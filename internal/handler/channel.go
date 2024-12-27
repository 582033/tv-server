package handler

import (
	"tv-server/internal/model/mongodb"

	"net/url"

	"github.com/gin-gonic/gin"
)

// 根据传入的频道名称获取当前频道下有多少记录,支持多频道
func GetRecordNums(c *gin.Context) {
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

func ListAllChannel(c *gin.Context) {
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
func HandleChannelPage(c *gin.Context) {
	c.HTML(200, "template/channels.html", gin.H{
		"title":  "频道分类",
		"active": "category",
	})
}
