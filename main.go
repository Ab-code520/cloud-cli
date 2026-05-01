package main

import "cloud-cli/cmd"

// 导入所有后端驱动
import (
	_ "cloud-cli/backends/quark"
)

func main() {
	cmd.Execute()
}
