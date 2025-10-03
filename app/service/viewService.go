package service

import (
	"context"
	"strconv"

	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/config/database"
)

// 增加浏览量
func ViewCount(confessionID uint) error {
	key := "confession:view:" + strconv.Itoa(int(confessionID))
	// 每次浏览就自增1
	_, err := database.RedisClient.Incr(context.Background(), key).Result()
	return err
}

// 获取浏览量
func GetViewCount(confessionID uint) (int64, error) {
	key := "confession:view:" + strconv.Itoa(int(confessionID))
	return database.RedisClient.Get(context.Background(), key).Int64()
}
