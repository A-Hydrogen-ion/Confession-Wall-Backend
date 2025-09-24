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

// JWTMiddleware JWT认证中间件主函数
func JWTMiddleware(m Auth) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := tokenCheck(c, authHeaderCheck(c))
		// 查询数据库验证用户是否存在
		var user model.User
		result := m.db.First(&user, claims.UserID)
		if result.Error != nil {
			controller.ReturnMsg(c, http.StatusUnauthorized, "这个用户不存在喵")
			c.Abort()
			return
		}
		// 检查用户状态（小黑屋功能）
		// if user.Status == "banned" {
		// 	controller.ReturnMsg(c, http.StatusForbidden, "这个用户被拉进小黑屋了喵")
		// 	c.Abort()
		// 	return
		// }
		// 将用户信息保存到请求的上下文
		c.Set("user_id", user.UserID)
		c.Set("username", user.Username)
		c.Set("nickname", user.Nickname)
		c.Set("user_claims", claims) // 保存完整的claims信息
		c.Next()
	}
}
func (m *Auth) JWTMiddlewareLight() gin.HandlerFunc { // 轻量级验证版本，不查询数据库（仅验证token）
	return func(c *gin.Context) {
		claims := tokenCheck(c, authHeaderCheck(c))
		// 仅从token中获取用户信息，不查询数据库
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username) // 确保你的claims包含username
		c.Set("user_claims", claims)

		c.Next()
	}
}
func authHeaderCheck(c *gin.Context) string { //预检查Authorization头
	authHeader := c.Request.Header.Get("Authorization") // 从Header中获取token
	if authHeader == "" {                               // 检查Authorization头
		controller.ReturnMsg(c, http.StatusUnauthorized, "你没有权限访问哦喵")
		c.Abort()
		return ""
	}
	if !strings.HasPrefix(authHeader, "Bearer ") { // 检查是否以"Bearer "开头
		controller.ReturnMsg(c, http.StatusUnauthorized, "token格式错误")
		c.Abort()
		return ""
	}
	return authHeader
}
func tokenCheck(c *gin.Context, authHeader string) *jwt.CustomClaims { //检查解析出的token并返回结构体
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	tokenString = strings.TrimSpace(tokenString)

	if tokenString == "" { //token为空
		controller.ReturnMsg(c, http.StatusUnauthorized, "token不能为空")
		c.Abort()
		return nil
	}
	claims, err := jwt.ParseToken(tokenString)
	if err != nil { //token无效或已过期
		controller.ReturnMsg(c, http.StatusUnauthorized, "token无效或已过期了喵")
		c.Abort()
		return nil
	}
	return claims
}
