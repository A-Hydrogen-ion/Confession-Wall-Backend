package middleware

import (
	"net/http"
	"strings"

	//"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/controller"
	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/jwt"
	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Auth struct {
	db *gorm.DB
}

func NewAuth(db *gorm.DB) *Auth {
	return &Auth{db: db}
}

// JWTMiddleware JWT认证中间件
func (m *Auth) JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Header中获取token
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "你没有权限访问哦喵",
			})
			//确保请求完全终止
			c.Abort()
			return
		}

		// 检查token格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "token格式是错误的喵",
			})
			c.Abort()
			return
		}

		// 解析token
		claims, err := jwt.ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "token无效或已过期了喵",
				"error":   err.Error(),
			})
			c.Abort()
			return
		}
		var user model.User
		result := m.db.First(&user, claims.UserID)
		if result.Error != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "这个用户不存在喵",
			})
			c.Abort()
			return
		}
		// 如果之后写了小黑屋
		/*if user.blacklist = "active" {
		    c.JSON(http.StatusForbidden, gin.H{
		        "code":    403,
		        "message": "这个用户被拉近小黑屋了喵",
		    })
		    c.Abort()
		    return
		}*/
		// 将当前请求的 userID 信息保存到请求的上下文context，并从database中获取用户实时数据
		c.Set("user_id", user.UserID)
		c.Set("username", user.Username)
		c.Set("nickname", user.Nickname)

		c.Next()
	}
}
