package mongodb

import (
	"context"
	"fmt"
	"log"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	//"github.com/micro/go-micro/v2/config"
)

var (
	client     *mongo.Client
	clientOnce sync.Once // 确保 MongoDB 连接只会初始化一次
)

func initDB() {
	clientOnce.Do(func() {
		// MongoDB 连接字符串
		uri := "mongodb://root:123456@192.168.50.100:27017"
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

	return
}

func DB(dbName string) *mongo.Database {
	if dbName == "" {
		dbName = "tv-server"
	}

	initDB()
	return client.Database(dbName)
}

func Collection(dbName, collectionName string) *mongo.Collection {
	return DB(dbName).Collection(collectionName)
}
