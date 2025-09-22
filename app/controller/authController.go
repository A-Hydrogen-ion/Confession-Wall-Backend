package controller

import (
	"fmt"
	"net/http"

	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/jwt"
	models "github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/model"
	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 已经写不动controller的星期五的小绵羊也写不动注释了
type AuthController struct {
	userService *service.UserService
	db          *gorm.DB
}

// 你只需要直到这样做能用而且router调用没有问题，别问，问就是Artificial Intellgence大手笔
func NewAuthController(db *gorm.DB) *AuthController {
	return &AuthController{
		userService: service.NewUserService(),
		db:          db,
	}
}

// 修改以提示史山可读性
// 注册
func (authController *AuthController) Register(c *gin.Context) {
	var input models.RegisterRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查用户名是否已存在
	exists, err := authController.userService.CheckUsernameExists(input.Username)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库查询错误"})
		return
	}
	if exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户已存在"})
		return
	}

	// 创建用户
	user := models.User{
		Username: input.Username,
		Password: input.Password, // 映射到数据库 password_hash 列
		Nickname: input.Nickname, // 使用输入的昵称
	}

	if err := authController.userService.CreateUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败"})
		return
	}

	// 生成 token
	token, err := jwt.GenerateToken(user.UserID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token 生成失败了喵"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{"token": token},
		"msg":  "success",
	})
}

// 登录
func (authController *AuthController) Login(c *gin.Context) {
	var input models.LoginRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查找用户
	user, err := authController.userService.GetUserByUsername(input.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 验证密码
	if err := user.CheckPassword(input.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 生成 token
	token, err := jwt.GenerateToken(user.UserID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token 生成失败了喵"})
		return
	}

	// 返回响应
	response := models.AuthResponse{
		UserID: user.UserID,
		Token:  token,
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": response,
		"msg":  "success",
	})
}

// JWT 中间件
func (authController *AuthController) JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取 token
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token不见了喵"})
			c.Abort()
			return
		}

		// 校验 token
		claims, err := jwt.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token 无效喵"})
			c.Abort()
			return
		}

		// 将 user_id 设置到上下文
		c.Set("user_id", claims.UserID)
		c.Next()
	}
}
