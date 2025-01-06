package mongodb

import (
	"context"
	"fmt"
	"log"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"tv-server/internal/model/types"
	"tv-server/utils/core"
)

type Provider struct {
	client   *mongo.Client
	m3u      types.M3URepository
	favorite types.FavoriteRepository
}

var (
	instance *Provider
	once     sync.Once
)

// NewProvider 创建 MongoDB 提供者实例
func NewProvider() (*Provider, error) {
	var err error
	once.Do(func() {
		instance = &Provider{}
		err = instance.connect()
		if err == nil {
			instance.m3u = newM3URepository(instance.client)
			instance.favorite = newFavoriteRepository(instance.client)
		}
	})
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func (p *Provider) connect() error {
	cfg := core.GetConfig()
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d",
		cfg.DB.MongoDB.Username,
		cfg.DB.MongoDB.Password,
		cfg.DB.MongoDB.Host,
		cfg.DB.MongoDB.Port,
	)

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	p.client = client
	log.Println("Successfully connected to MongoDB")
	return nil
}

func (p *Provider) M3U() types.M3URepository {
	return p.m3u
}

func (p *Provider) Favorite() types.FavoriteRepository {
	return p.favorite
}

func (p *Provider) Close() error {
	if p.client != nil {
		return p.client.Disconnect(context.Background())
	}
	return nil
}
