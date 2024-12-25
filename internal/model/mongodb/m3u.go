package mongodb

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type (
	Name string
	Url  string
)

// MediaStream结构体用于表示媒体流信息
type MediaStream struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"` // 媒体流ID，omitempty表示在bson序列化时，如果ID为空，则忽略此字段
	CreatedAt   int64              `bson:"createdAt"`     // 创建时间戳
	UpdatedAt   int64              `bson:"updatedAt"`     // 更新时间戳
	StreamName  string             `bson:"streamName"`    // 媒体流名称
	StreamLogo  string             `bson:"streamLogo"`    // 媒体流Logo
	ChannelName string             `bson:"channelName"`   // 频道分类
	StreamUrl   []string           `bson:"streamUrl"`     // 媒体流链接
	// 其他需要的字段可以在这里添加
}

func (ms *MediaStream) Collection() *mongo.Collection {
	return Collection("tv-server", "m3u")
}

type QueryFilter struct {
	StreamNameList  []Name
	ChannelNameList []Name
}

func (ms *MediaStream) Save(c *gin.Context) error {
	if ms == nil {
		return nil
	}
	now := time.Now().Unix()
	ms.CreatedAt = now
	ms.UpdatedAt = now
	collection := ms.Collection()
	_, err := collection.InsertOne(c, ms)
	return err
}

func BatchSave(c *gin.Context, msList []*MediaStream) error {
	if len(msList) == 0 {
		return nil
	}

	collection := (&MediaStream{}).Collection()
	batchSize := 100

	//每次100条批量插入, 插入方法用BulkWrite
	for i := 0; i < len(msList); i += batchSize {
		end := i + batchSize
		if end > len(msList) {
			end = len(msList)
		}
		batch := msList[i:end]

		// 将batch切片转换为 interface{} 切片
		batchInterface := make([]interface{}, len(batch))
		for j := range batch {
			batchInterface[j] = batch[j]
		}
		if debugBytes, _ := json.Marshal(batchInterface); len(debugBytes) > 0 {
			fmt.Printf("RequestID:%v DebugMessage:%s Value:%s", nil, "batchInterface", string(debugBytes))
		}

		_, err := collection.InsertMany(c, batchInterface)
		if err != nil {
			fmt.Println(err)
		}
	}

	return nil
}

func (q QueryFilter) ToBson() bson.M {
	filter := bson.M{}
	if len(q.StreamNameList) > 0 {
		filter["streamName"] = bson.M{"$in": q.StreamNameList}
	}
	if len(q.ChannelNameList) > 0 {
		filter["channelName"] = bson.M{"$in": q.ChannelNameList}
	}
	return filter
}

func (ms *MediaStream) GetList(c *gin.Context, filter QueryFilter) ([]*MediaStream, error) {
	//选择collection
	collection := ms.Collection()

	cursor, err := collection.Find(c, filter.ToBson())
	if err != nil {
		return nil, err
	}
	defer cursor.Close(c)

	var msList []*MediaStream
	if err := cursor.All(c, &msList); err != nil {
		return nil, err
	}

	return msList, nil
}
