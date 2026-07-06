// Package web 提供前端静态文件的 go:embed 嵌入。
//
// 注意：go:embed 不支持 ".." 路径，只能嵌入本包源文件所在目录及其子目录的文件。
// 本包源文件放在 frontend/ 目录下，才能访问其子目录 out（即 frontend/out 构建产物）。
package web

import (
	"embed"
	"io/fs"
)

// Dist 嵌入前端构建产物 out（即 frontend/out）。使用 all: 前缀确保 _next/ 等以下划线开头的目录也被嵌入.
//
//go:embed all:out
var Dist embed.FS

// DistFS 返回前端构建产物的 fs.FS，供 gin 挂载静态文件.
func DistFS() fs.FS {
	sub, err := fs.Sub(Dist, "out")
	if err != nil {
		panic(err)
	}
	return sub
}
