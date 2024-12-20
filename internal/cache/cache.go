package cache

import "path/filepath"

var (
	UploadDir = filepath.Join("data", "uploads")           // 上传文件存储目录
	CacheFile = filepath.Join("data", "cache", "iptv.m3u") // 缓存文件路径
)
