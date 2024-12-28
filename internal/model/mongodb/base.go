package mongodb

import (
	"context"
	"fmt"
	"log"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"tv-server/utils/core"
)

var (
	client     *mongo.Client
	clientOnce sync.Once // 确保 MongoDB 连接只会初始化一次
)

func initDB() {
	clientOnce.Do(func() {
		cfg := core.GetConfig()

		// 使用配置文件中的MongoDB连接信息
		uri := cfg.MongoDB.URI
		clientOptions := options.Client().ApplyURI(uri)

		// 连接到 MongoDB
		var err error
		client, err = mongo.Connect(context.Background(), clientOptions)
		if err != nil {
			log.Fatalf("Failed to connect to MongoDB: %v", err)
		}

		// 检查连接是否成功
		err = client.Ping(context.Background(), nil)
		if err != nil {
			log.Fatalf("Failed to ping MongoDB: %v", err)
		} else {
			fmt.Println("Successfully connected to MongoDB")
		}
	})
}

func DB(dbName string) *mongo.Database {
	if dbName == "" {
		cfg := core.GetConfig()
		dbName = cfg.MongoDB.Database // 从配置文件获取默认数据库名
	}

	initDB()
	return client.Database(dbName)
}

func Collection(dbName, collectionName string) *mongo.Collection {
	return DB(dbName).Collection(collectionName)
}
