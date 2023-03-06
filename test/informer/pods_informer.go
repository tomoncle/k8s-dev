package main

import (
	"fmt"
	dev "k8s-dev/pkg/k8s"
	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"strings"
	"time"
)

type Controller struct {
	indexer  cache.Indexer
	queue    workqueue.RateLimitingInterface
	informer cache.Controller
}

func NewController(queue workqueue.RateLimitingInterface, indexer cache.Indexer, informer cache.Controller) *Controller {
	return &Controller{
		informer: informer,
		indexer:  indexer,
		queue:    queue,
	}
}

func (c *Controller) processNextItem() bool {
	// 等待工作队列中有新 item
	key, quit := c.queue.Get()
	if quit {
		return false
	}

	//告诉队列我们已经处理完此key。这将为其他工作人员解锁key
	//这允许安全的并行处理，因为具有相同key的两个pod永远不会在
	defer c.queue.Done(key)

	// 调用包含业务逻辑的方法
	err := c.syncToStdout(key.(string))
	// 如果在执行业务逻辑过程中出现问题，则处理错误
	c.handleErr(err, key)
	return true
}

// syncToStdout 是控制器的业务逻辑。在这个控制器中，它只需打印有关pod到stdout的信息。如果发生错误，它只需返回错误。
// 重试逻辑不应是业务逻辑的一部分。
func (c *Controller) syncToStdout(key string) error {
	action, key := takeKey(key)

	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		fmt.Println("在存储获取对象 key: ", key, " 失败，原因为: ", err)
		return err
	}

	if !exists { // 下面我们将用一个Pod来预热缓存，这样我们将看到一个Pod的删除
		fmt.Println("Pod", key, "不存在")
		return nil
	}

	podInfo := obj.(*v1.Pod)
	// 注意，如果您有本地控制的资源，还必须检查uid
	// 取决于实际实例，以检测是否使用相同的名称重新创建了Pod
	switch action {
	case cache.Updated:
		fmt.Println("修改 Pod", podInfo.Name, "，创建时间：", podInfo.GetCreationTimestamp(), "，删除时间：", podInfo.GetDeletionTimestamp())
	case cache.Added:
		fmt.Println("添加 Pod", podInfo.Name, "，创建时间：", podInfo.GetCreationTimestamp())
	case cache.Deleted:
		fmt.Println("删除 Pod", podInfo.Name, "，删除时间：", podInfo.GetDeletionTimestamp())
	default:
		fmt.Println("action: ", action, "key: ", key)
	}
	return nil
}

// handleErr 检查是否发生错误，并确保稍后重试
func (c *Controller) handleErr(err error, key interface{}) {
	if err == nil {
		c.queue.Forget(key)
		return
	}

	// 如果出现问题，此控制器将重试5次。之后，它停止尝试。
	if c.queue.NumRequeues(key) < 5 {
		fmt.Println(fmt.Sprintf("Error syncing pod %v: %v", key, err))

		//限制key速率重新排队。基于队列和重新排队历史记录，稍后将再次处理key
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	// 多次重试失败，抛出异常到runtime.HandleError
	runtime.HandleError(err)
	fmt.Println("删除的 pod: ", key, "不在队列中:", err)
}

func (c *Controller) Run(workerSize int, stopCh chan struct{}) {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	fmt.Println("Starting Pod controller")
	go c.informer.Run(stopCh)

	// 在开始处理队列中的项目之前，等待所有相关的缓存同步
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("等待缓存同步时超时"))
		return
	}

	for i := 0; i < workerSize; i++ {
		// 启动一个协程，每隔一定的时间，就去运行runWorker函数，直到接收到结束信号 就关闭这个协程
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
	fmt.Println("Stopping Pod controller")
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
	}
}

func makeKey(deltaType cache.DeltaType, k string) string {
	return fmt.Sprintf("%s:%s", deltaType, k)
}

func takeKey(k string) (cache.DeltaType, string) {
	arr := strings.Split(k, ":")
	action := cache.DeltaType(arr[0])
	key := arr[1]
	return action, key
}

func main() {

	// 创建一个pod listWatch
	podListWatcher := dev.GetListWatchByDefaultNamespace(dev.POD)
	// 创建一个队列
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	resourceEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(makeKey(cache.Added, key))
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				queue.Add(makeKey(cache.Updated, key))
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(makeKey(cache.Deleted, key))
			}
		}}
	//创建 informer
	indexer, informer := cache.NewIndexerInformer(
		podListWatcher,
		&v1.Pod{},
		0,
		resourceEventHandler,
		cache.Indexers{})

	controller := NewController(queue, indexer, informer)

	// 模拟一个不存在的pod
	_ = indexer.Add(&v1.Pod{ObjectMeta: meta_v1.ObjectMeta{Name: "404-pod", Namespace: v1.NamespaceDefault}})

	// Now let's start the controller
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	// Wait forever
	select {}
}
