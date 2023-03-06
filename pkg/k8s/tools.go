package k8s

import (
	"fmt"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
)

// GetK8SDefaultConfig
//
//	@Description: 如果 $HOME/.kube/config 存在，返回默认的k8s配置
//	@return *rest.Config
func GetK8SDefaultConfig() *rest.Config {
	return ctrl.GetConfigOrDie()
}

// GetK8SConfig
//
//	@Description: 方法返回使用 masterUrl, kubeConfigPath 构建的k8s配置
//	@param masterUrl: k8sApiServer地址
//	@param kubeConfigPath: .kube/config 系统路径
//	@return *rest.Config
//	@return error
func GetK8SConfig(masterUrl, kubeConfigPath string) (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags(masterUrl, kubeConfigPath)
}

// GetDefaultK8SClient
//
//	@Description: 方法返回默认的k8s客户端
//	@return *kubernetes.Clientset
//	@return error
func GetDefaultK8SClient() (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(GetK8SDefaultConfig())
}

// GetListWatchByDefaultConfig
//
//	@Description: 根据默认配置，创建 ListWatch
//	@param resource
//	@param namespace
//	@return *cache.ListWatch
func GetListWatchByDefaultConfig(resource Resource, namespace string) *cache.ListWatch {
	k8sClient, err := GetDefaultK8SClient()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return cache.NewListWatchFromClient(k8sClient.CoreV1().RESTClient(), resource.toString(), namespace, fields.Everything())
}

// GetListWatchByDefaultNamespace
//
//	@Description: 创建默认命名空间资源的 listWatch
//	@param resource
//	@return *cache.ListWatch
func GetListWatchByDefaultNamespace(resource Resource) *cache.ListWatch {
	return GetListWatchByDefaultConfig(resource, DefaultNamespace)
}
