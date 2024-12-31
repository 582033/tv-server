package mongodb

import (
	"fmt"
	"time"
	"tv-server/utils/core"

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

func (ms *MediaStream) Save(ctx *core.Context) error {
	if ms == nil {
		return nil
	}
	now := time.Now().Unix()
	ms.CreatedAt = now
	ms.UpdatedAt = now
	collection := ms.Collection()
	_, err := collection.InsertOne(ctx.StdCtx, ms)
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

func (filter *QueryFilter) GetList(ctx *core.Context) ([]*MediaStream, error) {
	//选择collection
	collection := (&MediaStream{}).Collection()

	cursor, err := collection.Find(ctx.StdCtx, filter.ToBson())
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx.StdCtx)

	var msList []*MediaStream
	if err := cursor.All(ctx.StdCtx, &msList); err != nil {
		return nil, err
	}

	return msList, nil
}

// UpdateDoc 表示更新文档的结构
type UpdateDoc struct {
	StreamUrl  []string
	UpdatedAt  int64
	StreamLogo string
}

// ToBsonM 将 UpdateDoc 转换为 bson.M 格式
func (u *UpdateDoc) ToBsonM() bson.M {
	return bson.M{
		"$addToSet": bson.M{
			"streamUrl": bson.M{
				"$each": u.StreamUrl,
			},
		},
		"$set": bson.M{
			"updatedAt":  u.UpdatedAt,
			"streamLogo": u.StreamLogo,
		},
	}
}

func BatchSave(ctx *core.Context, msList []*MediaStream) error {
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
			ms.CreatedAt = now
			ms.UpdatedAt = now

			filter := bson.M{
				"streamName":  ms.StreamName,
				"channelName": ms.ChannelName,
			}

			updateDoc := &UpdateDoc{
				StreamUrl:  ms.StreamUrl,
				UpdatedAt:  ms.UpdatedAt,
				StreamLogo: ms.StreamLogo,
			}

			operations = append(operations, mongo.NewUpdateOneModel().
				SetFilter(filter).
				SetUpdate(updateDoc.ToBsonM()).
				SetUpsert(true))
		}

		// 执行 BulkWrite 操作
		if len(operations) > 0 {
			_, err := collection.BulkWrite(ctx.StdCtx, operations, options.BulkWrite().SetOrdered(false)) // 设置 SetOrdered(false) 批量操作不按顺序
			if err != nil {
				fmt.Println("批量写入失败:", err)
				return err
			}
			fmt.Printf("成功处理了 %d 条数据\n", len(operations))
		}
	}

	return nil
}

// 查出所有的distinct ChannelName
func (filter *QueryFilter) GetAllChannel(ctx *core.Context) ([]Name, error) {
	collection := (&MediaStream{}).Collection()
	//拼接条件
	cursor, err := collection.Aggregate(ctx.StdCtx, []bson.M{
		{"$match": filter.ToBson()},
		{"$group": bson.M{"_id": "$channelName"}},
		{"$sort": bson.M{"channelName": 1}},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx.StdCtx)

	var channelNameList []Name
	for cursor.Next(ctx.StdCtx) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		channelName := Name(result["_id"].(string))
		channelNameList = append(channelNameList, channelName)
	}
	return channelNameList, nil
}

// GetRecordNums 获取频道记录数
func (f *QueryFilter) GetRecordNums(ctx *core.Context) (map[Name]int64, error) {
	result := make(map[Name]int64)

	// 按照原始顺序获取记录数
	for _, channelName := range f.ChannelNameList {
		collection := (&MediaStream{}).Collection()
		count, err := collection.CountDocuments(ctx.StdCtx, bson.M{"channelName": channelName})
		if err != nil {
			return nil, err
		}
		result[channelName] = count
	}

	return result, nil
}
