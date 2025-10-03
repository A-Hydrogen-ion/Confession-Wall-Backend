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
		if claims == nil { // 这里必须加，以防止传入错误的token继续执行访问本就为空的claims造成空指针panic
			return
		}
		// 查询数据库验证用户是否存在
		var user model.User
		result := m.db.First(&user, claims.UserID)
		if result.Error != nil {
			controller.ReturnMsg(c, http.StatusUnauthorized, "这个用户不存在喵")
			c.Abort()
			return
		}
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
	if err != nil || claims == nil { // token无效或已过期和非法的token全部报告无效，防止继续执行造成空指针panic
		controller.ReturnMsg(c, http.StatusUnauthorized, "token无效或已过期了喵")
		c.Abort()
		return nil
	}
	return claims
}

// 太好了，原来是预留了接口但是啥也没写
// OptionalJWTMiddleware 可选JWT认证中间件：有token就解析，无token直接放行
func OptionalJWTMiddleware(m Auth) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			tokenString = strings.TrimSpace(tokenString)
			claims, err := jwt.ParseToken(tokenString)
			if err == nil && claims != nil { //加一层判断以防止出现传入的非法token没有被解析访问空claims产生空指针导致panic
				// token合法，注入user_id等
				c.Set("user_id", claims.UserID)
				c.Set("username", claims.Username)
				c.Set("user_claims", claims)
			}
			// token不合法就啥也不做，直接放行
		}
		c.Next()
	}
}
