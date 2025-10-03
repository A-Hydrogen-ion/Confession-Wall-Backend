package service

import (
	"context"
	"strconv"

	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/config/database"
)

// 赞赞赞赞赞赞赞！
func LikeConfession(confessionID uint, userID uint) error {
	// 用 Set 记录每个用户是否点赞过，防止重复点赞
	key := "confession:like:" + strconv.Itoa(int(confessionID)) //向 Redis 的 Set（集合）中添加一个用户 ID（member）
	member := strconv.Itoa(int(userID))
	// SADD 返回 1 表示新点赞，0 表示已点过
	_, err := database.RedisClient.SAdd(context.Background(), key, member).Result()
	if err != nil {
		return err
	}
	return nil
}

// 取消点赞
func UnlikeConfession(confessionID uint, userID uint) error {
	key := "confession:like:" + strconv.Itoa(int(confessionID))
	member := strconv.Itoa(int(userID))
	// SREM 返回 1 表示成功移除，0 表示成员不存在
	_, err := database.RedisClient.SRem(context.Background(), key, member).Result()
	if err != nil {
		return err
	}
	return nil
}

// 获取点赞数
func GetLikeCount(confessionID uint) (int64, error) {
	key := "confession:like:" + strconv.Itoa(int(confessionID))
	// SCARD 返回集合的元素数量，即成员的数量
	return database.RedisClient.SCard(context.Background(), key).Result()
}
