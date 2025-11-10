package version

import (
	"fmt"
	"runtime"
)

var (
	Version   = "unknown"         // Version 应用版本号
	GitCommit = "unknown"         // GitCommit Git 提交哈希
	BuildDate = "unknown"         // BuildDate 构建日期
	GoVersion = runtime.Version() // GoVersion Go 版本
)

// PrintVersion 打印版本信息
func PrintVersion() {
	fmt.Printf("版本号: %s\n", Version)
	fmt.Printf("Git 提交: %s\n", GitCommit)
	fmt.Printf("构建日期: %s\n", BuildDate)
	fmt.Printf("Go 版本: %s\n", GoVersion)
	fmt.Printf("系统架构: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}
