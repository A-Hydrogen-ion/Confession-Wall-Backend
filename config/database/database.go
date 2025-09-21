package database

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	// 构建 DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		viper.GetString("database.username"),
		viper.GetString("database.password"),
		viper.GetString("database.host"),
		viper.GetInt("database.port"),
		viper.GetString("database.name"))

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库失败", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("获取内容失败", err)
	}
	//数据库连接池配置
	sqlDB.SetMaxIdleConns(10)           //最大空闲连接数
	sqlDB.SetMaxOpenConns(100)          //最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Hour) //连接嘴大时间

	DB = db // 将连接赋值给全局变量
	//测试数据库连接
	var version string
	if err := DB.Raw("SELECT VERSION()").Scan(&version).Error; err != nil {
		fmt.Printf("数据库连接测试失败: %v", err)
	} else {
		log.Printf("数据库版本: %s", version)
	}
	fmt.Println("连接成功")
}

// 搞个可导出的函数
func GetDB() *gorm.DB {
	return DB
}

// 数据库保活检查
func Health() error {
	if DB == nil {
		return fmt.Errorf("数据库未连接")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Ping()
}

// 在启动后每次调用保证数据库没有悄悄四掉
func HealthMonitor(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			if err := Health(); err != nil {
				log.Printf("[Healthcheck] 数据库连接不上惹喵: %v", err)
				// 可选：自动重连逻辑
			} else {
				log.Printf("[Healthcheck] 数据库还活着！")
			}
		}
	}()
}
