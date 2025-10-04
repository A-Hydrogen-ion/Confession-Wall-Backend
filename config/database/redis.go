package database

import (
	"context"
	"fmt"
	"time"

	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/model"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// RedisClient 将RedisClient作为全局的 Redis 客户端用于操作Redis数据
var RedisClient *redis.Client

// InitRedis 初始化 Redis 连接，使用从viper传入的配置文件
func InitRedis(gormDB *gorm.DB) {
	addr := viper.GetString("redis.addr")
	password := viper.GetString("redis.password")
	db := viper.GetInt("redis.db")
	// 从配置文件读取 Redis 连接信息
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	Reconnect()
	// 初始化后自动预热排行榜
	if gormDB != nil {
		if err := PreheatHotRank(gormDB); err != nil {
			fmt.Printf("预热热度榜失败: %v\n", err)
		}
	}
}

// PreheatHotRank 预热热度榜
func PreheatHotRank(db *gorm.DB) error {
	var confessions []model.Confession //从数据库里面导入文件
	if err := db.Find(&confessions).Error; err != nil {
		return err
	}
	// 计算每个表白的热度分数并存入 Redis 有序集合
	for _, c := range confessions {
		// 将浏览量和点赞量计算来作为每个confession的score
		score := float64(c.ViewCount)*2 + float64(c.LikeCount)*5
		//使用上下文传入的ConfessionID作为member
		member := fmt.Sprintf("%d", c.ID)
		// 新建名为confession:hot的有序数组
		RedisClient.ZAdd(context.Background(), "confession:hot", redis.Z{
			Score:  score,
			Member: member,
		})
	}
	return nil
}

// Reconnect 一样的自动重连
func Reconnect() {
	maxRetries := 10                 //最大重试次数
	retryInterval := 5 * time.Second //每次重试事件
	var err error
	for i := 0; i < maxRetries; i++ {
		err = RedisClient.Ping(context.Background()).Err()
		if err == nil {
			return
		}
		fmt.Printf("呼叫Redis姬……没有回应（第%d次），%v，%d秒后重试...\n", i+1, err, int(retryInterval.Seconds()))
		time.Sleep(retryInterval)
	}
	panic("Redis 连接失败，服务器娘呼唤了好几次Redis姬对面还是冰冰凉: " + err.Error())
}
