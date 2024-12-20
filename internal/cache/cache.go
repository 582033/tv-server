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

		if time.Since(info.ModTime()) > MaxCacheAge {
			if err := os.Remove(filePath); err != nil {
				log.Printf("Failed to remove old cache file %s: %v", filePath, err)
			}
		}
	}

	// 如果文件数量超过限制，删除最旧的文件
	if len(files) > MaxCacheFiles {
		for i := MaxCacheFiles; i < len(files); i++ {
			filePath := filepath.Join(CacheDir, files[i].Name())
			if err := os.Remove(filePath); err != nil {
				log.Printf("Failed to remove excess cache file %s: %v", filePath, err)
			}
		}
	}
}
