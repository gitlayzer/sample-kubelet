package lease

import (
	"context"
	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/component-helpers/apimachinery/lease"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"
	"os"
	"time"
)

const (
	LeaseDurationSeconds = 40
	LeaseNameSpace       = "kube-node-lease"
)

// SetNodeOwnerFunc 设置 lease 持有者
func SetNodeOwnerFunc(c kubernetes.Interface, nodeName string) func(lease *coordinationv1.Lease) error {
	return func(lease *coordinationv1.Lease) error {
		if len(lease.OwnerReferences) == 0 {
			if node, err := c.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{}); err == nil {
				lease.OwnerReferences = []metav1.OwnerReference{
					{
						APIVersion: corev1.SchemeGroupVersion.WithKind("Node").Version,
						Kind:       corev1.SchemeGroupVersion.WithKind("Node").Kind,
						Name:       nodeName,
						UID:        node.UID,
					},
				}
			} else {
				klog.ErrorS(err, "Failed to get node when trying to set owner ref to the node lease", "node", klog.KRef("", nodeName))
				return err
			}
		}
		return nil
	}
}

// StartLeaseController 启动租约控制器
func StartLeaseController(clientSet kubernetes.Interface, nodeName string) {
	// 创建了一个真实的时钟对象
	Clock := clock.RealClock{}

	// 定义了租约的续约间隔
	renewInterval := time.Duration(LeaseDurationSeconds * 0.25)

	heartbeatFailureFunc := func() {
		klog.Infoln("Heartbeat failure, lease will be deleted")
		os.Exit(1)
	}

	klog.Infoln("Starting lease controller")

	// 创建了一个租约控制器
	ctl := lease.NewController(Clock, clientSet, nodeName, LeaseDurationSeconds, heartbeatFailureFunc, renewInterval, nodeName, LeaseNameSpace, SetNodeOwnerFunc(clientSet, nodeName))

	// 启动租约控制器
	go ctl.Run(context.Background())
}
