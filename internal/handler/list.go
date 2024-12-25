package handler

import (
	"encoding/json"
	"fmt"
	"tv-server/internal/logic/m3u"
	"tv-server/internal/model/mongodb"
	"tv-server/utils/cache"

	"github.com/gin-gonic/gin"
)

func List(c *gin.Context) {
	filter := &mongodb.QueryFilter{
		ChannelNameList: []mongodb.Name{
			"央视频道",
			"卫视频道",
		},
	}
	r, _ := filter.GetList(c)
	for k, v := range r {
		fmt.Printf("%v: ChannelName:%v, StreamName:%v, StreamUrl:%v, StreamLogo:%v\n", k, v.ChannelName, v.StreamName, v.StreamUrl, v.StreamLogo)
	}

	allEntries := make([]m3u.Entry, 0, len(r))
	for _, v := range r {
		allEntries = append(allEntries, m3u.Entry{
			Metadata: fmt.Sprintf("#EXTINF:-1 tvg-name=\"%s\" tvg-logo=\"%s\",group-title=\"%s\",%s", v.ChannelName, v.StreamLogo, v.ChannelName, v.StreamName),
			URL:      v.StreamUrl[0],
		})
	}
	//开始验证并去重
	_, finalValidEntries, err := m3u.ValidateAndUnique(allEntries, 1000, 100)
	if err != nil {
		return
	}
	if debugBytes, _ := json.Marshal(finalValidEntries); len(debugBytes) > 0 {
		fmt.Printf("RequestID:%v DebugMessage:%s Value:%s", nil, "finalValidEntries", string(debugBytes))
	}
	// 写入文件
	if len(finalValidEntries) > 0 {
		tempFile := cache.CacheFile
		if err := m3u.WriteToFile(finalValidEntries, tempFile); err != nil {
			fmt.Printf("写入文件失败: %v\n", err)
			return
		}
	}
}
