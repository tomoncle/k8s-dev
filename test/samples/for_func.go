package main

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/wait"
	"reflect"
	"time"
)

var count = 0

func logic(obj interface{}) bool {
	fmt.Println("hello world!", time.Now(), count, obj, reflect.TypeOf(obj))
	time.Sleep(time.Second * 1)
	count = count + 1
	return count < 10
}

func main() {
	// 创建一个信号
	stopCh := make(chan struct{})
	defer close(stopCh)

	//启动一个协程，每隔一定的时间，就去运行logic函数，直到接收到结束信号 就关闭这个协程
	go wait.Until(func() {
		logic(stopCh)
	}, time.Second, stopCh)

	// 等待五秒
	time.Sleep(time.Second * 5)
	// 关闭信号
	stopCh <- struct{}{}

	for logic("test") {
	}
}
