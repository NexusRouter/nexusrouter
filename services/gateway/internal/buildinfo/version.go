// Package buildinfo 承载构建时 -ldflags 注入的版本号，供健康检查等使用。
package buildinfo

// Version 为发布版本或构建标识；未注入时默认为 dev。
var Version = "dev"
