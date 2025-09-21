package routes

import (
	"github.com/gin-gonic/gin"
	//"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/controller"
	//"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/middleware"
)

func SetupRouter(r *gin.Engine) *gin.Engine {
	//路由设置
	// 公共路由
	//public := r.Group("/api/auth")
	{
		//public.POST("/register", controllers.Register)
		//public.POST("/login", controllers.Login)
	}

	// 受保护的路由
	// protected := r.Group("/api")
	// protected.Use(middleware.AuthMiddleware())//改用jwt认证，此处需修改
	{

		// protected.GET("/user/profile", controller.GetUserProfile)          //获取当前用户信息
		// protected.PUT("/user/profile", controller.UpdateUserProfile)       //更新用户信息
		// protected.PUT("/user/password", controller.UpdateUserPassword)     //修改密码
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

	return r
}
