package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	ChannelName string             `bson:"channelName"`   // 频道分类
	StreamUrl   []string           `bson:"streamUrl"`     // 媒体流链接
	// 其他需要的字段可以在这里添加
}

type QueryFilter struct {
	StreamNameList  []Name
	ChannelNameList []Name
}

func (ms *MediaStream) Save() error {
	if ms == nil {
		return nil
	}
	now := time.Now().Unix()
	ms.CreatedAt = now
	ms.UpdatedAt = now
	collection := Collection("tv-server", "m3u")
	_, err := collection.InsertOne(context.Background(), ms)
	return err
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

func (ms *MediaStream) GetList(filter QueryFilter) ([]*MediaStream, error) {
	//选择collection
	collection := Collection("tv-server", "m3u")

	cursor, err := collection.Find(context.TODO(), filter.ToBson())
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var msList []*MediaStream
	if err := cursor.All(context.TODO(), &msList); err != nil {
		return nil, err
	}

	return msList, nil
}
