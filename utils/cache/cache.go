package cache

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	LastValidation time.Time
	CacheMutex     sync.Mutex
	CacheDir       = filepath.Join(os.TempDir(), "tv-server")
	CacheFile      = filepath.Join(CacheDir, "validated.m3u")
)

const (
	MaxCacheFiles = 10             // 最大缓存文件数
	MaxCacheAge   = 24 * time.Hour // 缓存文件最大保存时间
)

func Init() error {
	return os.MkdirAll(CacheDir, 0755)
}

func Cleanup() {
	files, err := os.ReadDir(CacheDir)
	if err != nil {
		log.Printf("Failed to read cache directory: %v", err)
		return
	}

	// 删除过期文件
	for _, file := range files {
		filePath := filepath.Join(CacheDir, file.Name())
		info, err := file.Info()
		if err != nil {
			continue
		}

		// 删除超过24小时的文件
		if time.Since(info.ModTime()) > MaxCacheAge {
			if err := os.Remove(filePath); err != nil {
				log.Printf("Failed to remove old cache file %s: %v", filePath, err)
			}
		}
	}
}
