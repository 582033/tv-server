package sqlite

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
	"tv-server/internal/model/types"
	"tv-server/utils/core"
)

type m3uRepository struct {
	db *sql.DB
}

func newM3URepository(db *sql.DB) types.M3URepository {
	return &m3uRepository{
		db: db,
	}
}

func (r *m3uRepository) Save(ctx *core.Context, stream *types.MediaStream) error {
	now := time.Now().Unix()
	if stream.CreatedAt == 0 {
		stream.CreatedAt = now
	}
	stream.UpdatedAt = now

	tx, err := r.db.BeginTx(ctx.StdCtx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 插入或更新主记录
	result, err := tx.ExecContext(ctx.StdCtx, `
        INSERT INTO m3u (stream_name, channel_name, stream_logo, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?)
        ON CONFLICT(stream_name, channel_name) DO UPDATE SET
        stream_logo = ?,
        updated_at = ?
    `, stream.StreamName, stream.ChannelName, stream.StreamLogo, stream.CreatedAt, stream.UpdatedAt,
		stream.StreamLogo, stream.UpdatedAt)
	if err != nil {
		return err
	}

	// 获取m3u记录的ID
	var m3uID int64
	if id, err := result.LastInsertId(); err == nil {
		m3uID = id
	} else {
		// 如果是更新操作，需要查询ID
		err = tx.QueryRowContext(ctx.StdCtx, `
            SELECT id FROM m3u WHERE stream_name = ? AND channel_name = ?
        `, stream.StreamName, stream.ChannelName).Scan(&m3uID)
		if err != nil {
			return err
		}
	}

	// 插入URL记录
	for _, url := range stream.StreamUrl {
		_, err = tx.ExecContext(ctx.StdCtx, `
            INSERT OR IGNORE INTO stream_urls (m3u_id, url)
            VALUES (?, ?)
        `, m3uID, url)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *m3uRepository) BatchSave(ctx *core.Context, streams []*types.MediaStream) error {
	if len(streams) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx.StdCtx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now().Unix()
	for _, stream := range streams {
		if stream.CreatedAt == 0 {
			stream.CreatedAt = now
		}
		stream.UpdatedAt = now

		// 插入或更新主记录
		result, err := tx.ExecContext(ctx.StdCtx, `
            INSERT INTO m3u (stream_name, channel_name, stream_logo, created_at, updated_at)
            VALUES (?, ?, ?, ?, ?)
            ON CONFLICT(stream_name, channel_name) DO UPDATE SET
            stream_logo = ?,
            updated_at = ?
        `, stream.StreamName, stream.ChannelName, stream.StreamLogo, stream.CreatedAt, stream.UpdatedAt,
			stream.StreamLogo, stream.UpdatedAt)
		if err != nil {
			return err
		}

		// 获取m3u记录的ID
		var m3uID int64
		if id, err := result.LastInsertId(); err == nil {
			m3uID = id
		} else {
			// 如果是更新操作，需要查询ID
			err = tx.QueryRowContext(ctx.StdCtx, `
                SELECT id FROM m3u WHERE stream_name = ? AND channel_name = ?
            `, stream.StreamName, stream.ChannelName).Scan(&m3uID)
			if err != nil {
				return err
			}
		}

		// 插入URL记录
		for _, url := range stream.StreamUrl {
			_, err = tx.ExecContext(ctx.StdCtx, `
                INSERT OR IGNORE INTO stream_urls (m3u_id, url)
                VALUES (?, ?)
            `, m3uID, url)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *m3uRepository) GetList(ctx *core.Context, filter *types.QueryFilter) ([]*types.MediaStream, error) {
	query := `
        SELECT m.id, m.created_at, m.updated_at, m.stream_name, m.stream_logo, m.channel_name, GROUP_CONCAT(u.url) as urls
        FROM m3u m
        LEFT JOIN stream_urls u ON m.id = u.m3u_id
    `

	var conditions []string
	var args []interface{}

	if len(filter.StreamNameList) > 0 {
		placeholders := make([]string, len(filter.StreamNameList))
		for i, name := range filter.StreamNameList {
			placeholders[i] = "?"
			args = append(args, name)
		}
		conditions = append(conditions, fmt.Sprintf("m.stream_name IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(filter.ChannelNameList) > 0 {
		placeholders := make([]string, len(filter.ChannelNameList))
		for i, name := range filter.ChannelNameList {
			placeholders[i] = "?"
			args = append(args, name)
		}
		conditions = append(conditions, fmt.Sprintf("m.channel_name IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " GROUP BY m.id"

	rows, err := r.db.QueryContext(ctx.StdCtx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var streams []*types.MediaStream
	for rows.Next() {
		var stream types.MediaStream
		var urls string
		if err := rows.Scan(&stream.ID, &stream.CreatedAt, &stream.UpdatedAt,
			&stream.StreamName, &stream.StreamLogo, &stream.ChannelName, &urls); err != nil {
			return nil, err
		}
		if urls != "" {
			stream.StreamUrl = strings.Split(urls, ",")
		}
		streams = append(streams, &stream)
	}

	return streams, nil
}

func (r *m3uRepository) GetAllChannel(ctx *core.Context, filter *types.QueryFilter) ([]string, error) {
	query := `
        SELECT DISTINCT channel_name
        FROM m3u
    `

	var conditions []string
	var args []interface{}

	if len(filter.StreamNameList) > 0 {
		placeholders := make([]string, len(filter.StreamNameList))
		for i, name := range filter.StreamNameList {
			placeholders[i] = "?"
			args = append(args, name)
		}
		conditions = append(conditions, fmt.Sprintf("stream_name IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(filter.ChannelNameList) > 0 {
		placeholders := make([]string, len(filter.ChannelNameList))
		for i, name := range filter.ChannelNameList {
			placeholders[i] = "?"
			args = append(args, name)
		}
		conditions = append(conditions, fmt.Sprintf("channel_name IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY channel_name"

	rows, err := r.db.QueryContext(ctx.StdCtx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []string
	for rows.Next() {
		var channel string
		if err := rows.Scan(&channel); err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}

	return channels, nil
}

func (r *m3uRepository) GetRecordNums(ctx *core.Context, filter *types.QueryFilter) (map[string]int64, error) {
	result := make(map[string]int64)

	for _, channelName := range filter.ChannelNameList {
		var count int64
		err := r.db.QueryRowContext(ctx.StdCtx, `
            SELECT COUNT(*) FROM m3u WHERE channel_name = ?
        `, channelName).Scan(&count)
		if err != nil {
			return nil, err
		}
		result[channelName] = count
	}

	return result, nil
}
