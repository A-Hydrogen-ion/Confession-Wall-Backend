package middleware

import (
	"net/http"
	"strings"

	//"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/controller"
	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/jwt"

	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware JWT认证中间件
func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Header中获取token
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无权限访问",
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
				"message": "token格式错误",
			})
			c.Abort()
			return
		}

		// 解析token
		claims, err := jwt.ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "token无效或已过期",
				"error":   err.Error(),
			})
			c.Abort()
			return
		}

		// 将当前请求的 userID 信息保存到请求的上下文context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)

		c.Next()
	}
}
