package mongodb

import (
	"fmt"
	"time"
	"tv-server/internal/model/types"
	"tv-server/utils/core"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type m3uRepository struct {
	client *mongo.Client
}

func newM3URepository(client *mongo.Client) types.M3URepository {
	return &m3uRepository{
		client: client,
	}
}

func (r *m3uRepository) collection() *mongo.Collection {
	return r.client.Database("tv-server").Collection("m3u")
}

func (r *m3uRepository) Save(ctx *core.Context, stream *types.MediaStream) error {
	now := time.Now().Unix()
	if stream.CreatedAt == 0 {
		stream.CreatedAt = now
	}
	stream.UpdatedAt = now

	collection := r.collection()
	filter := bson.M{
		"streamName":  stream.StreamName,
		"channelName": stream.ChannelName,
	}

	update := bson.M{
		"$addToSet": bson.M{
			"streamUrl": bson.M{
				"$each": stream.StreamUrl,
			},
		},
		"$set": bson.M{
			"updatedAt":  stream.UpdatedAt,
			"streamLogo": stream.StreamLogo,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(ctx.StdCtx, filter, update, opts)
	return err
}

func (r *m3uRepository) BatchSave(ctx *core.Context, streams []*types.MediaStream) error {
	if len(streams) == 0 {
		return nil
	}

	collection := r.collection()
	batchSize := 1000
	now := time.Now().Unix()

	for i := 0; i < len(streams); i += batchSize {
		end := i + batchSize
		if end > len(streams) {
			end = len(streams)
		}
		batch := streams[i:end]

		var operations []mongo.WriteModel
		for _, stream := range batch {
			if stream.CreatedAt == 0 {
				stream.CreatedAt = now
			}
			stream.UpdatedAt = now

			filter := bson.M{
				"streamName":  stream.StreamName,
				"channelName": stream.ChannelName,
			}

			update := bson.M{
				"$addToSet": bson.M{
					"streamUrl": bson.M{
						"$each": stream.StreamUrl,
					},
				},
				"$set": bson.M{
					"updatedAt":  stream.UpdatedAt,
					"streamLogo": stream.StreamLogo,
				},
			}

			operations = append(operations, mongo.NewUpdateOneModel().
				SetFilter(filter).
				SetUpdate(update).
				SetUpsert(true))
		}

		if len(operations) > 0 {
			_, err := collection.BulkWrite(ctx.StdCtx, operations, options.BulkWrite().SetOrdered(false))
			if err != nil {
				return fmt.Errorf("批量写入失败: %v", err)
			}
		}
	}

	return nil
}

func (r *m3uRepository) GetList(ctx *core.Context, filter *types.QueryFilter) ([]*types.MediaStream, error) {
	collection := r.collection()

	bsonFilter := bson.M{}
	if len(filter.StreamNameList) > 0 {
		bsonFilter["streamName"] = bson.M{"$in": filter.StreamNameList}
	}
	if len(filter.ChannelNameList) > 0 {
		bsonFilter["channelName"] = bson.M{"$in": filter.ChannelNameList}
	}

	cursor, err := collection.Find(ctx.StdCtx, bsonFilter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx.StdCtx)

	var streams []*types.MediaStream
	if err := cursor.All(ctx.StdCtx, &streams); err != nil {
		return nil, err
	}

	return streams, nil
}

func (r *m3uRepository) GetAllChannel(ctx *core.Context, filter *types.QueryFilter) ([]string, error) {
	collection := r.collection()

	bsonFilter := bson.M{}
	if len(filter.StreamNameList) > 0 {
		bsonFilter["streamName"] = bson.M{"$in": filter.StreamNameList}
	}
	if len(filter.ChannelNameList) > 0 {
		bsonFilter["channelName"] = bson.M{"$in": filter.ChannelNameList}
	}

	pipeline := []bson.M{
		{"$match": bsonFilter},
		{"$group": bson.M{"_id": "$channelName"}},
		{"$sort": bson.M{"_id": 1}},
	}

	cursor, err := collection.Aggregate(ctx.StdCtx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx.StdCtx)

	var results []struct {
		ID string `bson:"_id"`
	}
	if err := cursor.All(ctx.StdCtx, &results); err != nil {
		return nil, err
	}

	channels := make([]string, len(results))
	for i, result := range results {
		channels[i] = result.ID
	}

	return channels, nil
}

func (r *m3uRepository) GetRecordNums(ctx *core.Context, filter *types.QueryFilter) (map[string]int64, error) {
	result := make(map[string]int64)
	collection := r.collection()

	for _, channelName := range filter.ChannelNameList {
		count, err := collection.CountDocuments(ctx.StdCtx, bson.M{"channelName": channelName})
		if err != nil {
			return nil, err
		}
		result[channelName] = count
	}

	return result, nil
}
