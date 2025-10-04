package controller

import (
	"context"
	"strconv"

	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/service"
	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/config/database"
	"github.com/gin-gonic/gin"
)

func GetHotConfessions(c *gin.Context) {
	// 获取前10个热度最高的 confessionID
	// 由于数组默认从小到大排序，所以使用ZRevRange
	ids, err := database.RedisClient.ZRevRange(context.Background(), "confession:hot", 0, 9).Result()
	if err != nil {
		respondJSON(c, 500, "热度榜获取失败了喵", nil)
		return
	}
	// 转换为 uint 并查数据库详情
	var confessionIDs []uint
	for _, idStr := range ids { // 返回字符串数组
		id, _ := strconv.ParseUint(idStr, 10, 64)       // 转换为 uint64
		confessionIDs = append(confessionIDs, uint(id)) // 转换为 uint
	}
	// 根据ID列表获取表白详情
	confessions, err := service.GetConfessionsByID(database.DB, confessionIDs)
	if err != nil {
		respondJSON(c, 500, "热度榜详情获取失败了喵", nil)
		return
	}
	c.JSON(200, gin.H{"data": confessions})
}
