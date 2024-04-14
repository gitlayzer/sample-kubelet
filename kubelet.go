package main

import (
	"k8s.io/component-base/cli"
	"os"
	"sample_kubelet/app"
)

func main() {
	command := app.NewKubeletCommand()
	code := cli.Run(command)
	os.Exit(code)
}
