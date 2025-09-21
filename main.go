package main

import (
	"fmt"
	"log"

	//middleware "github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/middleware"
	"time"

	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/config/config"
	database "github.com/A-Hydrogen-ion/Confession-Wall-Backend/config/database"
	routes "github.com/A-Hydrogen-ion/Confession-Wall-Backend/config/router"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	config.InitViper()      //读取配置
	database.ConnectDB()    // 连接数据库
	if database.DB == nil { // 检查数据库连接是否成功
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
	// 获取配置
	port := viper.GetInt("server.port")
	host := viper.GetString("server.host")
	// 创建 Gin 引擎
	r := gin.Default()
	// 设置路由
	r = routes.SetupRouter(r) //函数暂时未定义
	addr := fmt.Sprintf("%s:%d", host, port)
	log.Printf("服务启动在 :%s", addr)
	// 启动服务器
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败啦！: %v", err)
	}
	database.HealthMonitor(30 * time.Second) // 每30秒检查一次数据库是否还活着
}
