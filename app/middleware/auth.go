package middleware

import (

	//"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/controller"
	"net/http"
	"strings"

	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/controller"
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
		authHeader := c.Request.Header.Get("Authorization") // 从Header中获取token
		claims := tokenCheck(c, authHeader)
		var user model.User
		result := m.db.First(&user, claims.UserID)
		if result.Error != nil {
			controller.ReturnMsg(c, 401, "这个用户不存在喵")
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
func tokenCheck(c *gin.Context, authHeader string) *jwt.CustomClaims {
	if authHeader == "" {
		controller.ReturnMsg(c, http.StatusUnauthorized, "你没有权限访问哦喵")
		c.Abort() //确保请求完全终止
		return nil
	}
	parts := strings.SplitN(authHeader, " ", 2) // 检查token格式
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		controller.ReturnMsg(c, http.StatusUnauthorized, "token格式是错误的喵")
		c.Abort()
		return nil
	}
	claims, err := jwt.ParseToken(parts[1]) // 解析token
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "token无效或已过期了喵",
			"error":   err.Error(),
		})
		c.Abort()
		return nil
	}
	return claims
}
