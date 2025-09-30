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

// 搞个可导出的函数
func GetDB() *gorm.DB {
	return DB
}

// 通用数据库连接重试函数
func connectWithRetry(dsn string, maxRetries int, retryInterval time.Duration) (*gorm.DB, error) {
	var db *gorm.DB
	var err error
	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			return db, nil
		}
		log.Printf("呼叫数据库姬……数据库姬没有回应（第%d次），%v，%d秒后重试...", i+1, err, int(retryInterval.Seconds()))
		time.Sleep(retryInterval)
	}
	return nil, err
}

// 构建用于读取配置文件 DSN 的通用函数
func getDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		viper.GetString("database.username"),
		viper.GetString("database.password"),
		viper.GetString("database.host"),
		viper.GetInt("database.port"),
		viper.GetString("database.name"))
}

// 连接数据库
func ConnectDB() {
	dsn := getDSN()
	maxRetries := 10
	retryInterval := 5 * time.Second
	// 设置重连次数
	db, err := connectWithRetry(dsn, maxRetries, retryInterval)
	if err != nil {
		log.Fatal("数据库连接失败，服务器娘呼唤了好几次数据库姬对面还是冰冰凉: ", err) //在最终连接失败后直接退出
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("获取连接失败", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	// 对连接池进行配置
	DB = db
	var version string
	if err := DB.Raw("SELECT VERSION()").Scan(&version).Error; err != nil {
		fmt.Printf("数据库连接测试失败: %v", err)
	} else {
		log.Printf("数据库版本: %s", version)
	}
	fmt.Println("数据库姬回应了服务器娘")
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
				fmt.Printf("[Healthcheck] 数据库姬没有回应惹喵: %v", err)
				// 保活失败时自动重连
				maxRetries := 5
				retryInterval := 3 * time.Second
				var db *gorm.DB
				var connErr error
				// 调用传入的配置文件
				dsn := getDSN()
				//调用重试连接函数
				db, connErr = connectWithRetry(dsn, maxRetries, retryInterval)
				if connErr == nil {
					DB = db
					fmt.Printf("[Healthcheck] 数据库姬诈尸了！") //重连后报告成功
				} else {
					fmt.Printf("[Healthcheck] 数据库连接失败，服务器娘呼唤了好几次数据库姬对面还是冰冰凉:%v", connErr)
				}
			} else {
				log.Printf("[Healthcheck] 数据库还活着！")
			}
		}
	}()
}
