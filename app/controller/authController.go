package controller

import (
	"fmt"
	"net/http"

	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/jwt"
	models "github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/model"
	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// 自定义返回函数
func Return400(c *gin.Context, err error) error {
	c.JSON(http.StatusBadRequest, gin.H{
		"code":  400,
		"user":  nil,
		"msg":   err.Error(),
		"token": nil,
	})
	return nil
}

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

// 输入要求控制函数
func checkInputRequirement(c *gin.Context, validationErrors validator.ValidationErrors) {

	for _, fieldError := range validationErrors {
		switch fieldError.Tag() {
		//各种类型的错误处理
		case "required": //不存在必须字段
			c.JSON(http.StatusBadRequest, gin.H{
				"code":  400,
				"user":  nil,
				"msg":   fmt.Sprintf("%s 是必填字段", fieldError.Field()),
				"token": nil,
			})
			return
		case "min": //字段长度过短
			c.JSON(http.StatusBadRequest, gin.H{
				"code":  400,
				"user":  nil,
				"msg":   fmt.Sprintf("%s 长度不能少于 %s 个字符", fieldError.Field(), fieldError.Param()),
				"token": nil,
			})
			return
		case "max": //字段长度过长
			c.JSON(http.StatusBadRequest, gin.H{
				"code":  400,
				"user":  nil,
				"msg":   fmt.Sprintf("%s 长度不能超过 %s 个字符", fieldError.Field(), fieldError.Param()),
				"token": nil,
			})
			return
		}
	}
}

// 用户存在验证函数
func (authController *AuthController) isUserExist(c *gin.Context, input models.RegisterRequest) {
	exists, err := authController.userService.CheckUsernameExists(input.Username)
	//其他错误处理
	if err != nil {
		fmt.Println(err)
		Return400(c, err)
		return
	}
	//存在返回逻辑
	if exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":  400,
			"user":  nil,
			"msg":   "用户已存在",
			"token": nil,
		})
		return
	}
}

// 修改以提示史山可读性
// 注册主函数
func (authController *AuthController) Register(c *gin.Context) {
	var input models.RegisterRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if ok {
			checkInputRequirement(c, validationErrors)//调用输入错误处理函数
			return
		}
		// 其他类型的错误处理（？）
		c.JSON(http.StatusBadRequest, gin.H{
			"code":  400,
			"user":  nil,
			"msg":   err.Error(),
			"token": nil,
		})
		return
	}

	// 检查用户名是否已存在
	authController.isUserExist(c, input)
	// 创建用户
	user := models.User{
		Username: input.Username,
		Password: input.Password, // 映射到数据库 password_hash 列，自动使用BeforeSave钩子hash密码
		Nickname: input.Nickname, // 使用输入的昵称
	}
	//错误处理
	if err := authController.userService.CreateUser(&user); err != nil {
		Return400(c, err)
		return
	}

	// 生成 token
	token, err := jwt.GenerateToken(user.UserID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":  500,
			"user":  nil,
			"msg":   "token生成失败了喵",
			"token": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{"token": token},
		"msg":  "success",
	})
}

// 登录函数
func (authController *AuthController) Login(c *gin.Context) {
	var input models.LoginRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		Return400(c, err)
		return
	}

	// 查找用户模块
	user, err := authController.userService.GetUserByUsername(input.Username)
	if err != nil {
		Return400(c, err)
		return
	}

	// 验证密码模块
	if err := user.CheckPassword(input.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"data": nil,
			"msg":  "用户名或密码错误喵"})
		return
	}

	// 生成 token 模块
	token, err := jwt.GenerateToken(user.UserID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"data": nil,
			"msg":  "token生成失败喵"})
		return
	}

	// 返回响应模块
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

// JWT 中间件函数
func (authController *AuthController) JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取 token
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"data": nil,
				"msg":  "token不见了喵"})
			c.Abort()
			return
		}

		// 校验 token模块
		claims, err := jwt.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"data": nil,
				"msg":  "token 无效喵"})
			c.Abort()
			return
		}

		// 将 user_id 设置到上下文
		c.Set("user_id", claims.UserID)
		c.Next()
	}
}
