package main

import (
	"fmt"
	"log"
	"os"

	middleware "github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/middleware"
	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/model"
	"gorm.io/gorm"

	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/config/config"
	database "github.com/A-Hydrogen-ion/Confession-Wall-Backend/config/database"
	routes "github.com/A-Hydrogen-ion/Confession-Wall-Backend/config/router"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// 创建必要的文件夹，权限为755（拥有者可读写执行，组用户和其他用户可读执行）
func createUploadDirs() {
	dirs := []string{
		"uploads",
		"uploads/avatars",
		"uploads/confessions",
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				log.Fatalf("创建目录 %s 失败: %v", dir, err)
			}
		}
	}
}

func hel() *gorm.DB { // 数据库健康检查
	if err := database.Health(); err != nil {
		log.Fatal("健康检查失败: ", err)
	}
	db := database.GetDB()
	if db == nil {
		log.Fatal("无法获取数据库连接")
	}
	return db
}
func migrate(db *gorm.DB) { //数据库迁移及检查函数
	err := db.AutoMigrate(&model.User{})
	err = db.AutoMigrate(&model.Confession{}, &model.Comment{}, &model.Block{})
	if err != nil {
		log.Printf("数据库迁移失败: %v", err)
	}
	if err := database.Health(); err != nil {
		log.Fatal("数据库健康检查失败: ", err)
	}
}
func main() {
	createUploadDirs()           // 创建必要的文件夹
	config.InitViper()           //读取配置
	database.ConnectDB()         // 连接数据库
	if database.GetDB() == nil { // 检查数据库连接是否成功
		log.Fatal("数据库连接失败，程序退出") //使用Fatal以使程序自动结束
	}
	db := hel()                              //健康检查
	authMiddleware := middleware.NewAuth(db) //获取数据库实例并创建中间件
	migrate(db)                              // 自动迁移数据库
	port := viper.GetInt("server.port")      // 获取配置
	host := viper.GetString("server.host")
	if host == "" {
		host = "0.0.0.0" // 默认监听来自所有地址的请求
	}
	r := gin.Default()                    // 创建 Gin 引擎
	routerConfig := &routes.RouterConfig{ // 设置路由来配合router中的模块设计
		Engine:     r,
		Middleware: authMiddleware,
		DB:         db,
	}
	r = routes.SetupRouter(routerConfig)
	addr := fmt.Sprintf("%s:%d", host, port)
	log.Printf("服务启动在 :%s", addr)
	//go database.HealthMonitor(30 * time.Second) // 每30秒检查一次数据库是否还活着
	if err := r.Run(addr); err != nil { // 启动服务器
		log.Fatalf("服务启动失败啦！: %v", err)
	}
}
