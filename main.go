package main

import (
	"log"
	//middleware "github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/middleware"
	"time"

	database "github.com/A-Hydrogen-ion/Confession-Wall-Backend/config/database"
	routes "github.com/A-Hydrogen-ion/Confession-Wall-Backend/config/router"
	"github.com/gin-gonic/gin"
)

func main() {
	// 连接数据库
	database.ConnectDB()
	// 检查数据库连接是否成功
	if database.DB == nil {
		log.Fatal("数据库连接失败，程序退出") //使用Fatal以使程序自动结束
	}
	// 健康检查
	if err := database.Health(); err != nil {
		log.Fatal("健康检查失败: ", err)
	}
	//db := database.GetDB()
	//authMiddleware := middleware.NewAuth(db)
	//暂时没有需要认证的路由，注释掉
	// 自动迁移数据库
	err := database.DB.AutoMigrate() //模型尚未构建
	if err != nil {
		log.Printf("数据库迁移失败: %v", err)
	}
	if err := database.Health(); err != nil {
		log.Fatal("数据库li: ", err)
	}
	// 创建 Gin 引擎
	r := gin.Default()
	// 设置路由
	r = routes.SetupRouter(r) //函数暂时未定义
	log.Println("服务启动在 :8080")
	// 启动服务器
	r.Run(":8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("服务启动失败啦！: ", err)
	}
	database.HealthMonitor(30 * time.Second) // 每30秒检查一次数据库是否还活着
}
