package handler

import (
	"fmt"
	"time"
	"tv-server/internal/logic/m3u"
	"tv-server/internal/model/mongodb"
	"tv-server/utils/core"

	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

// 在文件开头添加这些结构体定义
type ChannelValidateRequest struct {
	ChannelNames []string `json:"channelNames"`
	Timeout      int      `json:"timeout"`
}

// 根据传入的频道名称获取当前频道下有多少记录,支持多频道
func GetRecordNums(c *core.Context) {
	channelNameList := make([]mongodb.Name, 0)
	channelNames := c.QueryArray("channelName[]")
	if len(channelNames) > 0 {
		for _, name := range channelNames {
			if name != "" && name != "all" {
				decodedName, err := url.QueryUnescape(name)
				if err != nil {
					continue
				}
				channelNameList = append(channelNameList, mongodb.Name(decodedName))
			}
		}
	}
	filter := &mongodb.QueryFilter{
		ChannelNameList: channelNameList,
	}
	recordNums, err := filter.GetRecordNums(c)
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

func ListAllChannel(c *core.Context) {
	filter := &mongodb.QueryFilter{}
	channelNameList, err := filter.GetAllChannel(c)
	if err != nil {
		c.JSON(500, gin.H{
			"code":    500,
			"message": "获取频道列表失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    200,
		"message": "success",
		"data":    channelNameList,
	})
}

// HandleChannelPage 处理频道分类页面
func HandleChannelPage(c *core.Context) {
	c.HTML(200, "template/channels.html", gin.H{
		"title":  "频道分类",
		"active": "category",
	})
}

// HandleChannelValidate 处理频道验证请求
func HandleChannelValidate(c *core.Context) {
	var req ChannelValidateRequest
	// 添加请求体解析
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
	channelNameList := make([]mongodb.Name, 0)
	for _, name := range req.ChannelNames {
		if name != "" && name != "all" {
			decodedName, err := url.QueryUnescape(name)
			if err != nil {
				continue
			}
			channelNameList = append(channelNameList, mongodb.Name(decodedName))
		}
	}

	// 创建查询过滤器
	filter := &mongodb.QueryFilter{
		ChannelNameList: channelNameList,
	}

	// 从MongoDB获取记录
	r, err := filter.GetList(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取频道记录失败",
		})
		return
	}

	allEntries := make([]m3u.Entry, 0, len(r))
	for _, v := range r {
		//如果url有多个，则都需要进行验证,最终去重
		metadata := fmt.Sprintf("#EXTINF:-1 tvg-name=\"%s\" tvg-logo=\"%s\",group-title=\"%s\",%s", v.ChannelName, v.StreamLogo, v.ChannelName, v.StreamName)
		for _, url := range v.StreamUrl {
			allEntries = append(allEntries, m3u.Entry{
				Metadata: metadata,
				URL:      url,
			})
		}
	}
	//req.Timeout单位是ms
	timeout := time.Duration(req.Timeout) * time.Millisecond
	//开始验证并去重
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
