package handler

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
	"tv-server/internal/logic/m3u"
	"tv-server/internal/model"
	"tv-server/internal/model/types"
	"tv-server/utils/core"
	"tv-server/utils/msg"

	"github.com/gin-gonic/gin"
)

// 在文件开头添加这些结构体定义
type ChannelValidateRequest struct {
	ChannelNames []string `json:"channelNames"`
	Timeout      int      `json:"timeout"`
}

// 根据传入的频道名称获取当前频道下有多少记录,支持多频道
func HandleGetRecordNums(c *core.Context) {
	channelNames := c.QueryArray("channelName[]")
	channelNameList := make([]string, 0)
	if len(channelNames) > 0 {
		for _, name := range channelNames {
			if name != "" && name != "all" {
				decodedName, err := url.QueryUnescape(name)
				if err != nil {
					continue
				}
				channelNameList = append(channelNameList, decodedName)
			}
		}
	}

	filter := &types.QueryFilter{
		ChannelNameList: channelNameList,
	}

	db := model.GetDB()
	recordNums, err := db.M3U().GetRecordNums(c, filter)
	if err != nil {
		c.JSON(500, gin.H{
			"code":    500,
			"message": "获取频道记录数失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    200,
		"message": "success",
		"data":    recordNums,
	})
}

func HandleListAllChannel(c *core.Context) {
	filter := &types.QueryFilter{}
	db := model.GetDB()
	channelNameList, err := db.M3U().GetAllChannel(c, filter)
	if err != nil {
		c.JSON(500, gin.H{
			"code":    500,
			"message": "获取频道列表失败",
			"error":   err.Error(),
		})
		c.WebResponse(msg.CodeError, nil, err)
		return
	}

	c.JSON(200, gin.H{
		"code":    200,
		"message": "success",
		"data":    channelNameList,
	})
}

// ChannelValidate 处理频道验证请求
func HandleChannelValidate(c *core.Context) {
	var req ChannelValidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数",
			"error":   err.Error(),
		})
		return
	}

	// 设置默认超时时间
	if req.Timeout <= 0 {
		req.Timeout = 5000 // 默认5秒
	}

	// 从请求中获取频道名称列表
	channelNameList := make([]string, 0)
	for _, name := range req.ChannelNames {
		if name != "" && name != "all" {
			decodedName, err := url.QueryUnescape(name)
			if err != nil {
				continue
			}
			channelNameList = append(channelNameList, decodedName)
		}
	}

	// 创建查询过滤器
	filter := &types.QueryFilter{
		ChannelNameList: channelNameList,
	}

	// 从数据库获取记录
	db := model.GetDB()
	r, err := db.M3U().GetList(c, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取频道记录失败",
		})
		return
	}

	allEntries := make([]m3u.Entry, 0, len(r))
	for _, v := range r {
		metadata := fmt.Sprintf("#EXTINF:-1 tvg-name=\"%s\" tvg-logo=\"%s\",group-title=\"%s\",%s",
			v.ChannelName, v.StreamLogo, v.ChannelName, v.StreamName)
		for _, url := range v.StreamUrl {
			allEntries = append(allEntries, m3u.Entry{
				Metadata: metadata,
				URL:      url,
			})
		}
	}

	timeout := time.Duration(req.Timeout) * time.Millisecond
	validEntries, finalValidEntries, err := m3u.ValidateAndUnique(allEntries, timeout, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "验证失败",
		})
		return
	}

	// 使用新的公共函数
	if err := SaveValidatedEntries(finalValidEntries); err != nil {
		c.JSON(http.StatusInternalServerError, ValidateResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// 返回响应
	c.JSON(http.StatusOK, ValidateResponse{
		Success: true,
		Message: "验证完成！",
		Stats: struct {
			Total  int `json:"total"`
			Unique int `json:"unique"`
			Valid  int `json:"valid"`
		}{
			Total:  len(allEntries),
			Unique: len(validEntries),
			Valid:  len(finalValidEntries),
		},
		M3ULink: fmt.Sprintf("http://%s/iptv.m3u", c.Request.Host),
	})
}

func HandleChannelDetail(c *core.Context) {
	channelName := c.Query("channelName")
	decodedName, err := url.QueryUnescape(channelName)
	log.Println("decodedName", decodedName)
	if err != nil {
		c.WebResponse(msg.CodeBadRequest, nil, fmt.Errorf("invalid channel name encoding: %v", err))
		return
	}
	if decodedName != "" {
		channelName = decodedName
	}

	filter := &types.QueryFilter{
		ChannelNameList: []string{channelName},
	}

	db := model.GetDB()
	streamList, err := db.M3U().GetList(c, filter)
	if err != nil {
		c.WebResponse(msg.CodeError, nil, err)
		return
	}

	c.WebResponse(msg.CodeOK, streamList, nil)
}
