// Package buildinfo 承载构建时 -ldflags 注入的版本号，供健康检查等使用。
package buildinfo

import "time"

// Version 为发布版本或构建标识；未注入时默认为 dev。
var Version = "dev"

// ProcessStart 为当前进程启动时刻（包初始化时记录），供健康检查返回运行时长等。
var ProcessStart time.Time

func init() {
	ProcessStart = time.Now().UTC()
}
