package client

import (
	"fmt"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"testing"
)

func TestPodsByDynamicClient(t *testing.T) {
	namespace := &coreV1.Namespace{}
	pod := &coreV1.Pod{}
	deploy := appsV1.Deployment{}
	fmt.Println(namespace)
	fmt.Println(pod)
	fmt.Println(deploy)

}
