package main

import (
	"context"
	"encoding/json"
	"fmt"
	dev "k8s-dev/pkg/k8s"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"testing"
)

func output(pods *coreV1.PodList, s string) {
	fmt.Println("*******************", s, "*******************")
	for _, pod := range pods.Items {
		fmt.Println(pod.Name)
	}
	fmt.Println()
}

// TestPodsByClientSet
// 方法一：使用 kubernetes.Clientset 来操作 pods
// ClientSet是在RESTClient的基础上封装了对Resource和Version的管理方法。
// ClientSet仅能访问Kubernetes自身内置的资源，不能直接访问CRD自定义的资源。
// 如果要想ClientSet访问CRD自定义资源，可通过client-gin代码生成器重新生成ClientSet。
func TestPodsByClientSet(t *testing.T) {
	client, err := dev.GetDefaultK8SClient()
	if err != nil {
		t.Error("获取k8s客户端异常：", err)
		return
	}
	data, err := client.CoreV1().Pods("default").List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		t.Error("使用kubernetes.Clientset获取pods异常：", err)
		return
	}
	output(data, "kubernetes.Clientset")
}

// TestPodsByRESTClient
// 方法2：使用 RESTClient客户端 来操作 pods
// 同时支持Json 和 protobuf
// 支持所有原生资源和CRDs
func TestPodsByRESTClient(t *testing.T) {
	//client, err := dev.GetDefaultK8SClient()
	//if err != nil {
	//	fmt.Println("获取k8s客户端异常：", err)
	//	return
	//}
	//restClient:= client.CoreV1().RESTClient()

	config := dev.GetK8SDefaultConfig()
	// 配置必须的参数
	config.APIPath = "api"
	config.GroupVersion = &coreV1.SchemeGroupVersion
	config.NegotiatedSerializer = scheme.Codecs
	restClient, err := rest.RESTClientFor(config)
	if err != nil {
		t.Error("获取k8s restClient异常：", err)
		return
	}
	// 存储数据
	data2 := &coreV1.PodList{}
	err = restClient.Get().Resource("pods").Namespace("default").VersionedParams(
		&metaV1.ListOptions{}, scheme.ParameterCodec).Do(context.TODO()).Into(data2)
	if err != nil {
		t.Error("使用RESTClient客户端获取pods异常：", err)
		return
	}
	output(data2, "RESTClient")
}

// TestPodsByDynamicClient
// 方法3：使用 DynamicClient客户端 来操作 pods
// DynamicClient：Dynamic client 是一种动态的 client，它能处理 kubernetes 所有的资源，包括CRD自定义资源。
// DynamicClient返回的对象是一个 map[string]interface{}，
// DynamicClient与ClientSet最大的不同就是，ClientSet仅能访问Kubernetes自带的资源，不能直接访问CRD自定义资源。
// ClientSet需要预先实现每种Resource和Version的操作，内部的数据都是结构化数据。
// DynamicClient内部实现了Unstructured，用于处理非结构化数据结构，这是能够处理CRD自定义资源的关键。
// DynamicClient不是类型安全的，因此访问CRD自定义资源时需要特别注意。
// 只支持JSON
func TestPodsByDynamicClient(t *testing.T) {
	dynamicClient, err := dynamic.NewForConfig(dev.GetK8SDefaultConfig())
	if err != nil {
		t.Error("获取k8s dynamicClient 异常：", err)
		return
	}
	gvr := schema.GroupVersionResource{Resource: "pods", Version: "v1"}
	// 返回非结构化对象 unstructuredData
	unstructuredData, err := dynamicClient.Resource(gvr).Namespace("default").List(context.TODO(), metaV1.ListOptions{})
	// 声明一个结构化数据结构
	data3 := &coreV1.PodList{}
	// 转换 unstructuredData 为 data3
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredData.UnstructuredContent(), data3)
	if err != nil {
		t.Error("使用DynamicClient客户端获取pods异常：", err)
		return
	}
	output(data3, "DynamicClient")

}

// TestApiGroupsByDiscoveryClient
// 使用 DiscoveryClient客户端 “不能操作 pods”，最终还是会转到 restClient 操作
// DiscoveryClient是发现客户端，主要用于发现Kubernetes API Server所支持的资源组、资源版本、资源信息。
// 除此之外，还可以将这些信息存储到本地，用户本地缓存，以减轻对Kubernetes API Server访问的压力。
// 在运行 Kubernetes 组件的机器上，缓存信息默认存储于~/.kube/cache 和 ~/.kube/http-cache 下。
// kubectl的api-versions和api-resources命令输出也是通过DiscoversyClient实现的
// 类似于kubectl命令 下面通过 DiscoveryClient 列出 Kubernetes API Server 所支持的资源组、资源版本、资源信息
func TestApiGroupsByDiscoveryClient(t *testing.T) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(dev.GetK8SDefaultConfig())
	if err != nil {
		t.Error("获取k8s discoveryClient 异常：", err)
		return
	}
	apiGroupArray, apiResourceListArray, err := discoveryClient.ServerGroupsAndResources()
	t.Log("*******************", "DiscoveryClient", "*******************")
	t.Log("*******************", "列出当前支持的ApiGroups:", "*******************")
	for _, apiGroup := range apiGroupArray {
		bytes, _ := json.Marshal(apiGroup.Versions)
		t.Log(string(bytes))
	}
	t.Log("*******************", "列出当前支持的apiResource:", "*******************")
	for _, apiResourceList := range apiResourceListArray {
		gv := apiResourceList.GroupVersion
		groupVersion, _ := schema.ParseGroupVersion(gv)
		//fmt.Println(gv, " ******* ", groupVersion)
		for _, apiResource := range apiResourceList.APIResources {
			t.Log(groupVersion, "-->", apiResource.Name)
		}

	}
}
