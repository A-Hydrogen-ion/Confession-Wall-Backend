package database
import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	dsn := "examplePasswor@tcp(127.0.0.1:3306)/forum_db?charset=utf8mb4&parseTime=True&loc=Local"
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

	DB = db
	fmt.Println("连接成功")
	//测试数据库连接
	var version string
	if err := DB.Raw("SELECT VERSION()").Scan(&version).Error; err != nil {
		log.Printf("数据库连接测试失败: %v", err)
	} else {
		log.Printf("数据库版本: %s", version)
	}
}
