package sqlite

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"tv-server/internal/model/types"
	"tv-server/utils/core"

	_ "github.com/mattn/go-sqlite3"
)

type sqliteProvider struct {
	db  *sql.DB
	m3u types.M3URepository
}

var (
	instance *sqliteProvider
	once     sync.Once
)

// NewProvider 创建 SQLite 提供者实例
func NewProvider() (types.DBProvider, error) {
	var err error
	once.Do(func() {
		instance = &sqliteProvider{}
		err = instance.connect()
		if err == nil {
			instance.m3u = newM3URepository(instance.db)
		}
	})
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func (p *sqliteProvider) connect() error {
	cfg := core.GetConfig()
	db, err := sql.Open("sqlite3", cfg.DB.SQLite.Path)
	if err != nil {
		return fmt.Errorf("failed to open SQLite database: %v", err)
	}

	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping SQLite database: %v", err)
	}

	// 创建必要的表
	if err = p.createTables(db); err != nil {
		return fmt.Errorf("failed to create tables: %v", err)
	}

	p.db = db
	log.Println("Successfully connected to SQLite database")
	return nil
}

func (p *sqliteProvider) createTables(db *sql.DB) error {
	// 创建媒体流表
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS m3u (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            created_at INTEGER NOT NULL,
            updated_at INTEGER NOT NULL,
            stream_name TEXT NOT NULL,
            stream_logo TEXT,
            channel_name TEXT NOT NULL,
            UNIQUE(stream_name, channel_name)
        );

        CREATE TABLE IF NOT EXISTS stream_urls (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            m3u_id INTEGER NOT NULL,
            url TEXT NOT NULL,
            FOREIGN KEY(m3u_id) REFERENCES m3u(id) ON DELETE CASCADE,
            UNIQUE(m3u_id, url)
        );

        CREATE INDEX IF NOT EXISTS idx_m3u_channel_name ON m3u(channel_name);
        CREATE INDEX IF NOT EXISTS idx_m3u_stream_name ON m3u(stream_name);
    `)
	return err
}

func (p *sqliteProvider) M3U() types.M3URepository {
	return p.m3u
}

func (p *sqliteProvider) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}
