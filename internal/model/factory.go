package model

import (
	"fmt"
	"sync"
	"tv-server/internal/model/mongodb"
	"tv-server/internal/model/sqlite"
	"tv-server/internal/model/types"
)

var (
	provider types.DBProvider
	once     sync.Once
)

// InitDB 初始化数据库连接
func InitDB(dbType string) error {
	var err error
	once.Do(func() {
		switch dbType {
		case types.DBTypeMongoDB:
			provider, err = mongodb.NewProvider()
		case types.DBTypeSQLite:
			provider, err = sqlite.NewProvider()
		default:
			err = fmt.Errorf("unsupported database type: %s", dbType)
		}
	})
	return err
}

// GetDB 获取数据库实例
func GetDB() types.DBProvider {
	if provider == nil {
		panic("database not initialized")
	}
	return provider
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if provider != nil {
		return provider.Close()
	}
	return nil
}
