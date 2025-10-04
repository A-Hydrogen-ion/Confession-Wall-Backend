package routes

import (
	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/controller"
	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RouterConfig struct {
	Engine     *gin.Engine
	Middleware *middleware.Auth
	DB         *gorm.DB
	// 其他配置按需添加，注意，添加了任何新字段都需要去main.go里的routerConfig设置路由
}

func SetupRouter(config *RouterConfig) *gin.Engine {
	db := config.DB
	authController := controller.NewAuthController(db) // 创建所有控制器实例
	userController := controller.NewUserController(db)
	confessionController := controller.CreateConfessionController(db)
	blockController := controller.NewBlockController(db)
	commentController := controller.NewCommentController(db)
	//路由设置
	config.Engine.Static("/uploads", "./uploads") // 静态文件服务，用于给前端展示图片
	auth := config.Engine.Group("/api/auth")      // 认证路由
	{
		auth.POST("/register", authController.Register)
		auth.POST("/login", authController.Login)
	}
	// 公共的 confession 路由（可选认证，仅list和comment）
	var m middleware.Auth
	m = *middleware.NewAuth(db)
	publicConfession := config.Engine.Group("/api/confession")
	{
		publicConfession.GET("/list", middleware.OptionalJWTMiddleware(m), confessionController.ListPublicConfessions) // 查看社区表白（可选认证）
		publicConfession.GET("/comment", middleware.OptionalJWTMiddleware(m), commentController.ListComments)          // 查看某条表白的评论（可选认证）
		publicConfession.GET("/detail", middleware.OptionalJWTMiddleware(m), confessionController.GetConfessionByID)   // 根据ID获取表白（可选认证）
		publicConfession.GET("/hot", middleware.OptionalJWTMiddleware(m), controller.GetHotConfessions)                // 获取热度榜（可选认证）
	}
	config.Engine.GET("api//user/detail", userController.GetUserProfileByID) // 获取用户详情，无需登录，知道用户ID就可以获取
	api := config.Engine.Group("/api")
	api.Use(middleware.JWTMiddleware(m)) //需要jwt认证的API公共路由
	{                                    // 用户相关路由可以在这里添加
		user := api.Group("/user")
		{
			user.GET("/profile", authController.GetMyProfile)      //获取自己的配置
			user.PUT("/profile", authController.UpdateUserProfile) //更新用户信息
			user.PUT("/avatar", userController.UploadAvatar)       // 上传头像
			user.PUT("/password", authController.ChangePassword)   //修改密码
		}
		// 受保护的 confession 路由（需要登录）
		privateConfession := api.Group("/confession")
		{
			privateConfession.POST("/post", confessionController.CreateConfession)     // 发布表白（需要登录），上传图片已经集成到了controller里
			privateConfession.POST("/update", confessionController.UpdateConfession)   // 修改表白（需要登录）
			privateConfession.DELETE("/delete", confessionController.DeleteConfession) // 删除表白（需要登录）
			privateConfession.POST("/comment", commentController.AddComment)           // 发布评论（需要登录）
			privateConfession.DELETE("/comment", commentController.DeleteComment)      // 删除评论（需要登录）
			privateConfession.GET("/user", confessionController.GetUserConfessions)    // 获取某用户所有表白（需要登录）
			privateConfession.POST("/like", controller.LikeConfession)                 // 点赞表白（需要登录）
		}
		block := api.Group("/blacklist")
		{
			block.POST("/add", blockController.BlockUser)
			block.POST("/remove", blockController.UnblockUser)
			block.GET("/list", blockController.GetBlockedUsers)
		}
	}
	return config.Engine
}
