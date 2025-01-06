package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"tv-server/internal/model/types"
)

type favoriteRepository struct {
	collection *mongo.Collection
}

func newFavoriteRepository(client *mongo.Client) types.FavoriteRepository {
	return &favoriteRepository{
		collection: client.Database("tv").Collection("favorites"),
	}
}

// CreateCategory 创建分类
func (r *favoriteRepository) CreateCategory(category *types.Category) error {
	ctx := context.Background()

	// 检查分类名是否已存在
	count, err := r.collection.CountDocuments(ctx, bson.M{"name": category.Name})
	if err != nil {
		return err
	}
	if count > 0 {
		return types.ErrCategoryExists
	}

	category.ID = primitive.NewObjectID().Hex()
	category.CreatedAt = time.Now().Unix()
	category.UpdatedAt = time.Now().Unix()

	_, err = r.collection.InsertOne(ctx, category)
	return err
}

// UpdateCategory 更新分类
func (r *favoriteRepository) UpdateCategory(category *types.Category) error {
	ctx := context.Background()
	category.UpdatedAt = time.Now().Unix()

	result, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": category.ID},
		bson.M{"$set": category},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return types.ErrCategoryNotFound
	}
	return nil
}

// DeleteCategory 删除分类
func (r *favoriteRepository) DeleteCategory(categoryID string) error {
	ctx := context.Background()

	// 删除分类
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": categoryID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return types.ErrCategoryNotFound
	}

	// 删除该分类下的所有收藏
	_, err = r.collection.UpdateMany(ctx,
		bson.M{},
		bson.M{"$pull": bson.M{"favorites": bson.M{"categoryId": categoryID}}},
	)
	return err
}

// GetCategories 获取所有分类
func (r *favoriteRepository) GetCategories() ([]*types.Category, error) {
	ctx := context.Background()
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var categories []*types.Category
	err = cursor.All(ctx, &categories)
	return categories, err
}

// AddFavorite 添加收藏
func (r *favoriteRepository) AddFavorite(favorite *types.Favorite) error {
	ctx := context.Background()

	// 检查是否已收藏
	count, err := r.collection.CountDocuments(ctx, bson.M{"streamUrl": favorite.StreamUrl})
	if err != nil {
		return err
	}
	if count > 0 {
		return types.ErrFavoriteExists
	}

	favorite.ID = primitive.NewObjectID().Hex()
	favorite.CreatedAt = time.Now().Unix()
	favorite.UpdatedAt = time.Now().Unix()

	_, err = r.collection.InsertOne(ctx, favorite)
	return err
}

// RemoveFavorite 移除收藏
func (r *favoriteRepository) RemoveFavorite(favoriteID string) error {
	ctx := context.Background()
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": favoriteID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return types.ErrFavoriteNotFound
	}
	return nil
}

// UpdateFavorite 更新收藏
func (r *favoriteRepository) UpdateFavorite(favorite *types.Favorite) error {
	ctx := context.Background()
	favorite.UpdatedAt = time.Now().Unix()

	result, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": favorite.ID},
		bson.M{"$set": favorite},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return types.ErrFavoriteNotFound
	}
	return nil
}

// GetFavorites 获取指定分类下的收藏
func (r *favoriteRepository) GetFavorites(categoryID string) ([]*types.Favorite, error) {
	ctx := context.Background()
	cursor, err := r.collection.Find(ctx, bson.M{"categoryId": categoryID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var favorites []*types.Favorite
	err = cursor.All(ctx, &favorites)
	return favorites, err
}

// GetAllFavorites 获取所有收藏
func (r *favoriteRepository) GetAllFavorites() ([]*types.Favorite, error) {
	ctx := context.Background()
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var favorites []*types.Favorite
	err = cursor.All(ctx, &favorites)
	return favorites, err
}

// MoveFavoriteToCategory 移动收藏到指定分类
func (r *favoriteRepository) MoveFavoriteToCategory(favoriteID string, categoryID string) error {
	ctx := context.Background()

	result, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": favoriteID},
		bson.M{
			"$set": bson.M{
				"categoryId": categoryID,
				"updatedAt":  time.Now().Unix(),
			},
		},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return types.ErrFavoriteNotFound
	}
	return nil
}
