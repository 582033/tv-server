package types

import "tv-server/utils/core"

// 定义数据库类型常量
const (
	DBTypeMongoDB = "mongodb"
	DBTypeSQLite  = "sqlite"
)

// M3UEntry 定义M3U条目结构
type M3UEntry struct {
	Title   string
	Channel string
	Logo    string
	URL     string
}

// MediaStream 定义媒体流信息结构
type MediaStream struct {
	ID          string   `json:"id"`
	CreatedAt   int64    `json:"createdAt"`
	UpdatedAt   int64    `json:"updatedAt"`
	StreamName  string   `json:"streamName"`
	StreamLogo  string   `json:"streamLogo"`
	ChannelName string   `json:"channelName"`
	StreamUrl   []string `json:"streamUrl"`
}

// QueryFilter 定义查询过滤条件
type QueryFilter struct {
	StreamNameList  []string
	ChannelNameList []string
}

// M3URepository 定义了 M3U 数据的仓库接口
type M3URepository interface {
	// Save 保存单个媒体流信息
	Save(ctx *core.Context, stream *MediaStream) error

	// BatchSave 批量保存媒体流信息
	BatchSave(ctx *core.Context, streams []*MediaStream) error

	// GetList 根据查询条件获取媒体流列表
	GetList(ctx *core.Context, filter *QueryFilter) ([]*MediaStream, error)

	// GetAllChannel 获取所有频道名称
	GetAllChannel(ctx *core.Context, filter *QueryFilter) ([]string, error)

	// GetRecordNums 获取各频道的记录数
	GetRecordNums(ctx *core.Context, filter *QueryFilter) (map[string]int64, error)
}

// DBProvider 定义数据库提供者接口
type DBProvider interface {
	// M3U 返回 M3U 仓库实现
	M3U() M3URepository

	// Close 关闭数据库连接
	Close() error
}
