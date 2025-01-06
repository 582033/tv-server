package sqlite

import (
	"database/sql"
	"time"

	"tv-server/internal/model/types"
)

type favoriteRepository struct {
	db *sql.DB
}

func newFavoriteRepository(db *sql.DB) types.FavoriteRepository {
	return &favoriteRepository{db: db}
}

// CreateCategory 创建分类
func (r *favoriteRepository) CreateCategory(category *types.Category) error {
	// 检查分类名是否已存在
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM categories WHERE name = ?", category.Name).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return types.ErrCategoryExists
	}

	now := time.Now().Unix()
	_, err = r.db.Exec(
		"INSERT INTO categories (name, created_at, updated_at) VALUES (?, ?, ?)",
		category.Name, now, now,
	)
	if err != nil {
		return err
	}

	return nil
}

// UpdateCategory 更新分类
func (r *favoriteRepository) UpdateCategory(category *types.Category) error {
	result, err := r.db.Exec(
		"UPDATE categories SET name = ?, updated_at = ? WHERE id = ?",
		category.Name, time.Now().Unix(), category.ID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return types.ErrCategoryNotFound
	}

	return nil
}

// DeleteCategory 删除分类
func (r *favoriteRepository) DeleteCategory(categoryID string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 删除分类
	result, err := tx.Exec("DELETE FROM categories WHERE id = ?", categoryID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return types.ErrCategoryNotFound
	}

	// 删除该分类下的所有收藏
	_, err = tx.Exec("DELETE FROM favorites WHERE category_id = ?", categoryID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetCategories 获取所有分类
func (r *favoriteRepository) GetCategories() ([]*types.Category, error) {
	rows, err := r.db.Query("SELECT id, name, created_at, updated_at FROM categories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*types.Category
	for rows.Next() {
		category := &types.Category{}
		err := rows.Scan(&category.ID, &category.Name, &category.CreatedAt, &category.UpdatedAt)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, rows.Err()
}

// AddFavorite 添加收藏
func (r *favoriteRepository) AddFavorite(favorite *types.Favorite) error {
	// 检查是否已收藏
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM favorites WHERE stream_url = ?", favorite.StreamUrl).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return types.ErrFavoriteExists
	}

	now := time.Now().Unix()
	_, err = r.db.Exec(
		`INSERT INTO favorites (category_id, stream_name, stream_logo, stream_url, channel_name, created_at, updated_at) 
         VALUES (?, ?, ?, ?, ?, ?, ?)`,
		favorite.CategoryID, favorite.StreamName, favorite.StreamLogo, favorite.StreamUrl,
		favorite.ChannelName, now, now,
	)
	return err
}

// RemoveFavorite 移除收藏
func (r *favoriteRepository) RemoveFavorite(favoriteID string) error {
	result, err := r.db.Exec("DELETE FROM favorites WHERE id = ?", favoriteID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return types.ErrFavoriteNotFound
	}

	return nil
}

// UpdateFavorite 更新收藏
func (r *favoriteRepository) UpdateFavorite(favorite *types.Favorite) error {
	result, err := r.db.Exec(
		`UPDATE favorites SET 
         category_id = ?, stream_name = ?, stream_logo = ?, stream_url = ?, 
         channel_name = ?, updated_at = ? 
         WHERE id = ?`,
		favorite.CategoryID, favorite.StreamName, favorite.StreamLogo, favorite.StreamUrl,
		favorite.ChannelName, time.Now().Unix(), favorite.ID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return types.ErrFavoriteNotFound
	}

	return nil
}

// GetFavorites 获取指定分类下的收藏
func (r *favoriteRepository) GetFavorites(categoryID string) ([]*types.Favorite, error) {
	rows, err := r.db.Query(
		`SELECT id, category_id, stream_name, stream_logo, stream_url, channel_name, created_at, updated_at 
         FROM favorites WHERE category_id = ?`,
		categoryID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var favorites []*types.Favorite
	for rows.Next() {
		favorite := &types.Favorite{}
		err := rows.Scan(
			&favorite.ID, &favorite.CategoryID, &favorite.StreamName, &favorite.StreamLogo,
			&favorite.StreamUrl, &favorite.ChannelName, &favorite.CreatedAt, &favorite.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		favorites = append(favorites, favorite)
	}

	return favorites, rows.Err()
}

// GetAllFavorites 获取所有收藏
func (r *favoriteRepository) GetAllFavorites() ([]*types.Favorite, error) {
	rows, err := r.db.Query(
		`SELECT id, category_id, stream_name, stream_logo, stream_url, channel_name, created_at, updated_at 
         FROM favorites`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var favorites []*types.Favorite
	for rows.Next() {
		favorite := &types.Favorite{}
		err := rows.Scan(
			&favorite.ID, &favorite.CategoryID, &favorite.StreamName, &favorite.StreamLogo,
			&favorite.StreamUrl, &favorite.ChannelName, &favorite.CreatedAt, &favorite.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		favorites = append(favorites, favorite)
	}

	return favorites, rows.Err()
}

// MoveFavoriteToCategory 移动收藏到指定分类
func (r *favoriteRepository) MoveFavoriteToCategory(favoriteID string, categoryID string) error {
	result, err := r.db.Exec(
		"UPDATE favorites SET category_id = ?, updated_at = ? WHERE id = ?",
		categoryID, time.Now().Unix(), favoriteID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return types.ErrFavoriteNotFound
	}

	return nil
}
