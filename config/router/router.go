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
	// 创建所有控制器实例
	authController := controller.NewAuthController(db)
	confessionController := controller.CreateConfessionController(db)

	// 认证路由
	auth := config.Engine.Group("/api/auth")
	{
		auth.POST("/register", authController.Register)
		auth.POST("/login", authController.Login)
	}
	//路由设置
	//需要jwt认证的API公共路由
	api := config.Engine.Group("/api")
	api.Use(authController.JWTMiddleware())
	{
		// 用户相关路由可以在这里添加
		user := api.Group("/user")
		{
			user.GET("/profile", authController.GetMyProfile)
			// user.PUT("/user/profile", controller.UpdateUserProfile) //更新用户信息
			// user.PUT("/user/password", controller.UpdateUserPassword)     //修改密码
		}
	}
	{
		//public.POST("/register", controllers.Register)
		//public.POST("/login", controllers.Login)
	}

	// 受保护的路由
	// protected := r.Group("/api")
	// protected.Use(middleware.AuthMiddleware())//改用jwt认证，此处需修改
	{

		// protected.POST("/upload/image", controller.UploadImage)            //上传图片
		// protected.DELETE("/upload/image/{imageId}", controller.DelImage)   //删除图片
		// protected.POST("/confessions", controller.PostConfession)          //发布表白
		// protected.GET("/confessions", controller.GetConfession)            //获取表白列表（社区）
		// protected.GET("/confessions/my", controller.GetMyConfession)       //获取个人表白列表
		// protected.GET("/confessions/{id}", controller.GetCOnfessionDetail) //获取表白详情
		// protected.PUT("/confessions/{id}", controller.UpdateConfession)    //更新表白
		// protected.POST("/comments", controller.PostComment)                //发表评论
		// protected.GET("/comments", controller.GetComment)                  //获取评论列表
		// protected.DELETE("/comments/{id}", controller.DelComment)          //删除评论
		// protected.POST("/blacklist", controller.BlacklistUser)             //拉黑用户
		// protected.POST("/blacklist/{userId}", controller.Unblock)          //取消拉黑
		// protected.GET("api/blacklist", controller.GetBlackList)            //获取拉黑列表
	}

	// 表白相关路由
	confession := config.Engine.Group("/api/confession")
	{
		// 发布表白（需要登录），上传图片已经集成到了controller里
		confession.POST("/post", authController.JWTMiddleware(), confessionController.CreateConfession)
		// 修改表白（需要登录）
		confession.POST("/update", authController.JWTMiddleware(), confessionController.UpdateConfession)
		// 发布评论（需要登录）
		confession.POST("/comment", authController.JWTMiddleware(), confessionController.AddComment)
		// 查看社区表白（无需登录）
		confession.GET("/list", confessionController.ListPublicConfessions)
		// 查看某条表白的评论（无需登录）
		confession.GET("/comments", confessionController.ListComments)
	}

	return config.Engine
}
