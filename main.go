package main

import (
	"log"

	database "github.com/A-Hydrogen-ion/Confession-Wall-Backend/config/database"
	routes "github.com/A-Hydrogen-ion/Confession-Wall-Backend/config/router"
	"github.com/gin-gonic/gin"
)

func main() {
	// 连接数据库
	database.ConnectDB()
	// 检查数据库连接是否成功
	if database.DB == nil {
		log.Println("数据库连接失败，程序退出")
	}
	// 自动迁移数据库
	err := database.DB.AutoMigrate() //模型尚未构建
	if err != nil {
		log.Printf("数据库迁移失败: %v", err)
	}
	// 创建 Gin 引擎
	r := gin.Default()
	// 设置路由
	r = routes.SetupRouter(r) //函数暂时未定义
	log.Println("服务器启动在 :8080")
	// 启动服务器
	r.Run(":8080")
}
