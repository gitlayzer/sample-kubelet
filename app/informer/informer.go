package informer

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
	"time"
)

func InitInformer(clientSet kubernetes.Interface, nodeName string) {
	// 创建 InformerFactory
	// 使用 field 过滤，只过滤出调度到当前节点的 Pod
	informerFactory := informers.NewSharedInformerFactoryWithOptions(clientSet, time.Minute*5, informers.WithTweakListOptions(func(options *metav1.ListOptions) {
		options.FieldSelector = fmt.Sprintf("spec.nodeName=%s", nodeName)
	}))

	// 创建一个 PodInformer
	podInformer := informerFactory.Core().V1().Pods().Informer()

	// 注册事件处理函数到 Pod 到 Informer
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// 这里对 Pod 的类型进行了断言
			pod := obj.(*corev1.Pod)
			klog.Infof("Pod Added：%s", pod.Name)
			// 当然这里可以处理 Pod 的新增事件
			// 设置 Pod 的状态为 Running
			err := SetPodStatus(clientSet, pod, corev1.PodRunning)
			if err != nil {
				klog.Errorf("Failed to set pod status: %v", err)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			// 断言新 Pod 的类型
			newPod := newObj.(*corev1.Pod)
			// 这里处理 Pod 的更新事件
			klog.Infof("Pod Updated：%s", newPod.Name)

			// 更新 Pod 的状态
			if newPod.Status.Phase == corev1.PodRunning {
				// 创建一个一分钟的定时器来更新 Pod状态
				// 但仅仅是为了模拟 Pod 状态的更新
				ticker := time.NewTicker(time.Minute)
				select {
				case <-ticker.C:
					err := SetPodStatus(clientSet, newPod, corev1.PodSucceeded)
					if err != nil {
						klog.Errorf("Failed to set pod status: %v", err)
						break
					}
					klog.Infof("Pod status updated: %s", newPod.Name)
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			// 断言 Pod 的类型
			pod := obj.(*corev1.Pod)
			// 这里处理 Pod 的删除事件
			klog.Infof("Pod Deleted：%s", pod.Name)
			RealDeletePod(clientSet, pod)
		},
	})

	// 启动 InformerFactory
	informerFactory.Start(wait.NeverStop)
	// 等待 InformerFactory 同步
	for r, ok := range informerFactory.WaitForCacheSync(wait.NeverStop) {
		if !ok {
			klog.Fatalf("Failed to sync informer: %v", r)
		}
	}
	klog.Infoln("Informer started")
}
