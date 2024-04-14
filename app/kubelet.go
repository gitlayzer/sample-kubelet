package app

import (
	"context"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	"sample_kubelet/app/client"
	"sample_kubelet/app/informer"
	"sample_kubelet/app/node"
	"sample_kubelet/app/node/lease"
	"time"
)

func NewKubeletCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "kubelet",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())

			defer cancel()

			nodeName := "virtual-kubelet-node"

			// 获取 k8s client
			clientSet, err := client.GetClientSet()
			if err != nil {
				return err
			}

			// 注册 node
			registeredNode := node.RegisterNode(clientSet, nodeName)

			// 定时上报 node 信息
			go wait.Until(func() {
				err = node.ReportNodeStatus(clientSet, registeredNode)
				if err != nil {
					panic(err)
				}
			}, time.Minute*5, wait.NeverStop)

			// 启动 lease 租约控制器
			lease.StartLeaseController(clientSet, nodeName)

			// 启动 Informer
			informer.InitInformer(clientSet, nodeName)

			klog.Infoln("kubelet 启动成功")

			select {
			case <-ctx.Done():
				break
			case <-wait.NeverStop:
				break
			}
			return nil
		},
	}

	return cmd
}
