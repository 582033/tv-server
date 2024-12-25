package mongodb

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func BatchSave(c *gin.Context, msList []*MediaStream) error {
	if len(msList) == 0 {
		return nil
	}

	collection := (&MediaStream{}).Collection()
	batchSize := 1000 // 每批次插入 1000 条数据，避免一次性插入过多数据

	now := time.Now().Unix()

	// 每次处理批量数据
	for i := 0; i < len(msList); i += batchSize {
		end := i + batchSize
		if end > len(msList) {
			end = len(msList)
		}
		batch := msList[i:end]

		// 构建 BulkWrite 操作的切片
		var operations []mongo.WriteModel
		for _, ms := range batch {
			// 每个元素更新 CreatedAt 和 UpdatedAt
			ms.CreatedAt = now
			ms.UpdatedAt = now

			// 使用 $addToSet 来保证 StreamUrl 中的 URL 不重复
			filter := bson.M{
				"streamName":  ms.StreamName,
				"channelName": ms.ChannelName,
			}

			// 批量更新或插入，如果存在则更新，如果不存在则插入
			update := bson.M{
				"$addToSet": bson.M{
					"streamUrl": bson.M{
						"$each": ms.StreamUrl, // 批量添加 URL
					},
				},
				"$set": bson.M{
					"updatedAt": ms.UpdatedAt,
					"logo":      ms.StreamLogo,
				},
			}

			// 构造批量操作模型，设置 upsert 为 true，表示数据不存在时插入新记录
			operations = append(operations, mongo.NewUpdateOneModel().
				SetFilter(filter).
				SetUpdate(update).
				SetUpsert(true)) // 若文档不存在，则插入新文档
		}

		// 执行 BulkWrite 操作
		if len(operations) > 0 {
			_, err := collection.BulkWrite(c, operations, options.BulkWrite().SetOrdered(false)) // 设置 SetOrdered(false) 批量操作不按顺序
			if err != nil {
				fmt.Println("批量写入失败:", err)
				return err
			}
			fmt.Printf("成功处理了 %d 条数据\n", len(operations))
		}
	}

	return nil
}
