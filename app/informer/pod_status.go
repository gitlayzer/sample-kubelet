package informer

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// SetPodStatus 用于修改 Pod 的状态
func SetPodStatus(client kubernetes.Interface, pod *corev1.Pod, phase corev1.PodPhase) error {
	// 定义 Pod 的状态
	pod.Status.Phase = phase

	// 更新 Pod 的状态
	_, err := client.CoreV1().Pods(pod.Namespace).UpdateStatus(context.Background(), pod, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

// RealDeletePod 用于删除 Pod
func RealDeletePod(client kubernetes.Interface, pod *corev1.Pod) {
	// 定义了 Pod 的状态为 0，表示已经删除
	var ps int64 = 0

	// 删除 Pod 操作，GracePeriodSeconds 参数表示优雅删除的时间，设置为 0 表示立即删除
	err := client.CoreV1().Pods(pod.Namespace).Delete(context.Background(), pod.Name, metav1.DeleteOptions{
		GracePeriodSeconds: &ps,
	})
	if err != nil {
		klog.Errorf("Failed to delete pod %s/%s: %v", pod.Namespace, pod.Name, err)
	} else {
		klog.Infof("Pod %s/%s deleted successfully", pod.Namespace, pod.Name)
	}
}
