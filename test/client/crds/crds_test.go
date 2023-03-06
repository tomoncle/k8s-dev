// https://mozillazg.com/2020/07/k8s-kubernetes-client-go-list-get-create-update-patch-delete-crd-resource-without-generate-client-code-update-or-create-via-yaml.html#hidlist
// sigs.k8s.io\controller-runtime@v0.14.1\pkg\client\namespaced_client_test.go
package main

import (
	"context"
	"fmt"
	devopsV1 "github.com/tomoncle/k8s-operator-nginx/api/v1"
	devopsClientV1 "github.com/tomoncle/k8s-operator-nginx/pkg/k8s/client"
	dev "k8s-dev/pkg/k8s"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func init() {
	fmt.Println("************* 注册 devopsv1.Nginx 到 scheme *************")
	//  注册 devopsv1.Nginx 到 scheme
	utilruntime.Must(devopsV1.AddToScheme(scheme.Scheme))
}

func TestClientSetGetCRDInstance(t *testing.T) {
	nginxClient, err := devopsClientV1.NewForConfig(dev.GetK8SDefaultConfig())
	if err != nil {
		t.Error(err)
	}
	nginxList, err := nginxClient.Nginxes("default").List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		t.Error(err)
	}
	for _, nginx := range nginxList.Items {
		t.Log(nginx.Name)
	}

	nginx, err := nginxClient.Nginxes("default").Get(context.TODO(), "nginx-sample", metaV1.GetOptions{})
	if err != nil {
		t.Error(err)
	}
	t.Log(nginx.Spec.Image)
}

func TestRuntimeClientGetCRDInstance(t *testing.T) {
	ctx := context.Background()
	cfg := dev.GetK8SDefaultConfig()
	nonNamespacedClient, _ := client.New(cfg, client.Options{Scheme: scheme.Scheme})
	runtimeClient := client.NewNamespacedClient(nonNamespacedClient, "default")

	var nginxList devopsV1.NginxList
	err := runtimeClient.List(ctx, &nginxList, &client.ListOptions{})
	if err != nil {
		t.Error(err, "查询失败.")
	} else {
		t.Log("查询 Nginx 列表：成功", "count", len(nginxList.Items))
		for _, nginx := range nginxList.Items {
			t.Log("查询 Nginx 列表：成功", "nginx", nginx.Name)
		}
	}

}

func TestRuntimeClientUpdateCRDInstance(t *testing.T) {
	ctx := context.Background()
	cfg := dev.GetK8SDefaultConfig()
	nonNamespacedClient, _ := client.New(cfg, client.Options{Scheme: scheme.Scheme})
	runtimeClient := client.NewNamespacedClient(nonNamespacedClient, "default")

	var nginxList devopsV1.NginxList
	err := runtimeClient.List(ctx, &nginxList, &client.ListOptions{})
	if err != nil {
		t.Error(err, "查询失败.")
	} else {
		t.Log("查询 Nginx 列表：成功", "count", len(nginxList.Items))
		for _, nginx := range nginxList.Items {
			t.Log("查询 Nginx 列表：成功", "nginx", nginx.Name)
			nginx.Spec.TLS[0].Hosts = append(nginx.Spec.TLS[0].Hosts, "dev-02.devops.com")
			err = runtimeClient.Update(ctx, &nginx, &client.UpdateOptions{})
			if err != nil {
				t.Error(err)
			} else {
				t.Log("修改成功", "当前ingress：", nginx.Spec.TLS)
			}
		}
	}

}

func TestRESTClientGetCRDInstance(t *testing.T) {
	crdConfig := dev.GetK8SDefaultConfig()
	crdConfig.ContentConfig.GroupVersion = &devopsV1.GroupVersion
	crdConfig.APIPath = "/apis"
	crdConfig.NegotiatedSerializer = scheme.Codecs
	crdConfig.UserAgent = rest.DefaultKubernetesUserAgent()

	restClient, err := rest.UnversionedRESTClientFor(crdConfig)
	if err != nil {
		t.Error(err, "获取restClient失败")
	}

	data := &devopsV1.NginxList{}
	err = restClient.Get().Resource("nginxes").Do(context.TODO()).Into(data)
	if err != nil {
		t.Error(err, "获取nginxes失败")
	}

	t.Log("success!")
	for _, o := range data.Items {
		t.Log("nginx: ", ", name:", o.Name, ", gvk:", o.GroupVersionKind())
	}

}
func TestRESTClientUpdateCRDInstance(t *testing.T) {
	crdConfig := dev.GetK8SDefaultConfig()
	crdConfig.ContentConfig.GroupVersion = &devopsV1.GroupVersion
	crdConfig.APIPath = "/apis"
	crdConfig.NegotiatedSerializer = scheme.Codecs
	crdConfig.UserAgent = rest.DefaultKubernetesUserAgent()

	restClient, err := rest.UnversionedRESTClientFor(crdConfig)
	if err != nil {
		t.Error(err, "获取restClient失败")
	}

	data := &devopsV1.NginxList{}
	err = restClient.Get().Resource("nginxes").Do(context.TODO()).Into(data)
	if err != nil {
		t.Error(err, "获取nginxes失败")
	}

	t.Log("success!")
	for _, ngx := range data.Items {
		t.Log("nginx: ", ", name:", ngx.Name, ", gvk:", ngx.GroupVersionKind())
		result := &devopsV1.Nginx{}
		ngx.Spec.Image = "nginx:latest"
		err = restClient.Put().
			Namespace("default").
			Resource("nginxes").
			Name("nginx-sample").
			VersionedParams(&metaV1.UpdateOptions{}, scheme.ParameterCodec).
			Body(&ngx).
			Do(context.TODO()).
			Into(result)
		if err != nil {
			t.Error("更新Nginx失败，", err)
		} else {
			t.Log("更新Nginx成功！", "当前镜像：", result.Spec.Image)
		}
	}

}
func TestDynamicClientGetCRDInstance(t *testing.T) {
	crdConfig := dev.GetK8SDefaultConfig()
	dynamicClient, err := dynamic.NewForConfig(crdConfig)

	if err != nil {
		t.Error("获取k8s dynamicClient 异常：", err)
		return
	}
	// gvr 定义
	gvr := schema.GroupVersionResource{
		Resource: "nginxes",
		Version:  devopsV1.GroupVersion.Version,
		Group:    devopsV1.GroupVersion.Group,
	}
	// 返回非结构化对象 unstructuredData
	unstructuredData, err := dynamicClient.Resource(gvr).Namespace("default").List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		t.Error("使用DynamicClient客户端获取nginx异常：", err)
		return
	}
	// 声明一个结构化数据结构
	data := &devopsV1.NginxList{}
	// 转换 unstructuredData 为 data3
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredData.UnstructuredContent(), data)
	if err != nil {
		t.Error(err, "获取nginxes失败")
		return
	}

	t.Log("success!")
	for _, o := range data.Items {
		t.Log("nginx: ", ", name:", o.Name, ", gvk:", o.GroupVersionKind())
	}

}
func TestDynamicClientUpdateCRDInstance(t *testing.T) {
	crdConfig := dev.GetK8SDefaultConfig()
	dynamicClient, err := dynamic.NewForConfig(crdConfig)

	if err != nil {
		t.Error("获取k8s dynamicClient 异常：", err)
		return
	}
	// gvr 定义
	gvr := schema.GroupVersionResource{
		Resource: "nginxes",
		Version:  devopsV1.GroupVersion.Version,
		Group:    devopsV1.GroupVersion.Group,
	}

	// ******************************查询
	// 返回非结构化对象 unstructuredData
	unstructuredData, err := dynamicClient.Resource(gvr).Namespace("default").Get(
		context.TODO(), "nginx-sample", metaV1.GetOptions{})
	if err != nil {
		t.Error("使用DynamicClient客户端获取nginx异常：", err)
		return
	}
	// 声明一个结构化数据结构
	data := &devopsV1.Nginx{}
	// 转换 unstructuredData 为 data
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredData.UnstructuredContent(), data)
	if err != nil {
		t.Error(err, "获取nginxes失败")
		return
	} else {
		t.Log("success!", "nginx:", data.Name)
	}

	// ****************************** 更新
	data.Spec.Image = "nginx:1.14.2"
	target, err := runtime.DefaultUnstructuredConverter.ToUnstructured(data)
	updated := &unstructured.Unstructured{Object: target}
	unstructuredData, err = dynamicClient.Resource(gvr).
		Namespace("default").
		Update(context.TODO(), updated, metaV1.UpdateOptions{})
	if err != nil {
		t.Error(err, "更新nginxes失败")
		return
	} else {
		utilruntime.Must(runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredData.UnstructuredContent(), data))
		t.Log("update success!", "当前镜像:", data.Spec.Image)
	}

}
