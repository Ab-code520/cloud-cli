package main

import "github.com/Ab-code520/cloud-cli/cmd"

// 导入所有后端驱动
import (
	_ "github.com/Ab-code520/cloud-cli/backends/quark"
)

func main() {
	cmd.Execute()
}
