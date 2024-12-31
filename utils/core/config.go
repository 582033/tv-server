package core

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// Config 配置结构体
type Config struct {
	Server struct {
		Port int `json:"port"`
	} `json:"server"`

	DB struct {
		Type   string `json:"type"` // mongodb 或 sqlite
		SQLite struct {
			Path string `json:"path"`
		} `json:"sqlite"`
		MongoDB struct {
			Host     string `json:"host"`
			Port     int    `json:"port"`
			Username string `json:"username"`
			Password string `json:"password"`
			Database string `json:"database"`
		} `json:"mongodb"`
	} `json:"db"`
}

var (
	config     *Config
	configOnce sync.Once
	configLock sync.RWMutex
)

// LoadConfig 加载配置文件
func LoadConfig(configPath string) error {
	configLock.Lock()
	defer configLock.Unlock()

	var loadErr error
	configOnce.Do(func() {
		config = &Config{}
		data, err := os.ReadFile(configPath)
		if err != nil {
			loadErr = fmt.Errorf("读取配置文件失败: %v", err)
			return
		}

		if err = json.Unmarshal(data, config); err != nil {
			loadErr = fmt.Errorf("解析配置文件失败: %v", err)
			return
		}
	})

	return loadErr
}

// GetConfig 获取配置
func GetConfig() *Config {
	configLock.RLock()
	defer configLock.RUnlock()
	return config
}

// UpdateConfig 更新配置
func UpdateConfig(newConfig *Config) {
	configLock.Lock()
	defer configLock.Unlock()
	config = newConfig
}
