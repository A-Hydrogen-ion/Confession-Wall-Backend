package service

import (
	"context"
	"strconv"
	"time"

	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/model"
	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/config/database"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// SyncRedisToMySQL 定时批量同步 Redis 点赞量和浏览量到 MySQL
func SyncRedisToMySQL(db *gorm.DB, confessionIDs []uint) {
	for _, id := range confessionIDs {
		likeCount, _ := database.RedisClient.SCard(context.Background(), "confession:like:"+strconv.Itoa(int(id))).Result() // 获取点赞数（数组长度）
		viewCount, _ := database.RedisClient.Get(context.Background(), "confession:view:"+strconv.Itoa(int(id))).Int64()    // 更新到 MySQL
		db.Model(&model.Confession{}).Where("id = ?", id).Updates(map[string]interface{}{                                   // 类型转换为uint并同步
			"like_count": uint(likeCount),
			"view_count": uint(viewCount),
		})
	}
}

// UpdateHotRank 定时批量计算热度分并更新热度榜
func UpdateHotRank(confessionIDs []uint) {
	for _, id := range confessionIDs {
		viewCount, _ := database.RedisClient.Get(context.Background(), "confession:view:"+strconv.Itoa(int(id))).Int64()
		likeCount, _ := database.RedisClient.SCard(context.Background(), "confession:like:"+strconv.Itoa(int(id))).Result()
		score := float64(viewCount)*2 + float64(likeCount)*5              // 热度公式
		database.RedisClient.ZAdd(context.Background(), "confession:hot", // 更新热度榜
			redis.Z{Score: score, Member: strconv.Itoa(int(id))})
	}
}

// StartRedisSync 启动定时任务
func StartRedisSync(db *gorm.DB) {
	ticker := time.NewTicker(60 * time.Second) // 每60秒同步一次并更新热度榜单
	go func() {
		for range ticker.C {
			var confessionIDs []uint
			db.Model(&model.Confession{}).Pluck("id", &confessionIDs) // 获取所有表白ID并同步
			// 批量同步点赞数和浏览量到 MySQL
			SyncRedisToMySQL(db, confessionIDs)
			UpdateHotRank(confessionIDs)
		}
	}()
}
