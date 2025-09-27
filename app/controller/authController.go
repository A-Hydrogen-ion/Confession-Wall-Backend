package controller

import (
	"fmt"
	"log"
	"net/http"

	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/jwt"
	models "github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/model"
	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// 已经写不动controller的星期五的小绵羊也写不动注释了
type AuthController struct {
	userService *service.UserService
	db          *gorm.DB
}

func ReturnError400(c *gin.Context, err error) error { // 自定义返回错误函数
	c.JSON(http.StatusBadRequest, gin.H{
		"code":  http.StatusBadRequest,
		"msg":   err.Error(),
		"token": nil,
	})
	return nil
}

func ReturnMsg(c *gin.Context, state int, msg string) error { //自定义返回消息函数
	c.JSON(state, gin.H{
		"code":  state,
		"msg":   msg,
		"token": nil,
	})
	return nil
}

// 创建AuthController
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
			ReturnMsg(c, http.StatusBadRequest, fmt.Sprintf("%s 是必填字段", fieldError.Field()))
			return
		case "min": //字段长度过短
			ReturnMsg(c, http.StatusBadRequest, fmt.Sprintf("%s 长度不能少于 %s 个字符", fieldError.Field(), fieldError.Param()))
			return
		case "max": //字段长度过长
			ReturnMsg(c, http.StatusBadRequest, fmt.Sprintf("%s 长度不能超过 %s 个字符", fieldError.Field(), fieldError.Param()))
			return
		}
	}
}

// 存在验证函数
func (authController *AuthController) checkExists(c *gin.Context, input string,
	checkFunc func(string) (bool, error), errorMsg string) bool {
	exists, err := checkFunc(input)
	if err != nil {
		log.Println("检查存在性错误:", err)
		ReturnError400(c, err)
		return false
	}
	if exists {
		ReturnMsg(c, http.StatusBadRequest, errorMsg)
		return false
	}
	return true
}

// 用户名存在验证函数
func (authController *AuthController) IsUserExist(c *gin.Context, input string) bool {
	return authController.checkExists(c, input, authController.userService.CheckUsernameExists, "用户名已存在") // 这里传递的是函数本身，不是调用结果
}

// 昵称存在验证函数
func (authController *AuthController) IsNicknameExist(c *gin.Context, input string) bool {
	return authController.checkExists(c, input, authController.userService.CheckNicknameExists, "昵称已存在") // 这里传递的是函数本身，不是调用结果
}

// 注册主函数// 修改以提示史山可读性
func (authController *AuthController) Register(c *gin.Context) {
	var input models.RegisterRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if ok {
			checkInputRequirement(c, validationErrors) //调用输入错误处理函数
			return
		}
		// 其他类型的错误处理（？）
		ReturnError400(c, err)
		return
	}
	isExist := authController.IsUserExist(c, input.Username) // 检查用户名是否已存在
	if !isExist {
		return
	}
	user := models.User{ // 创建用户
		Username: input.Username,
		Password: input.Password, // 映射到数据库 password_hash 列，自动使用BeforeSave钩子hash密码
		Nickname: input.Nickname, // 使用输入的昵称
	}
	if err := authController.userService.CreateUser(&user); err != nil { //错误处理
		ReturnError400(c, err)
		return
	}
	token, err := jwt.GenerateToken(user.UserID, user.Username) // 生成 token
	if err != nil {
		ReturnMsg(c, http.StatusBadRequest, "token生成失败了喵")
		return
	}
	c.JSON(http.StatusOK, gin.H{ //返回成功
		"code": http.StatusOK,
		"data": gin.H{"token": token},
		"msg":  "success",
	})
}

// 登录主函数
func (authController *AuthController) Login(c *gin.Context) {
	var input models.LoginRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		ReturnError400(c, err)
		return
	}
	// 查找用户模块
	user, err := authController.userService.GetUserByUsername(input.Username)
	if err != nil {
		ReturnMsg(c, http.StatusBadRequest, "用户名或密码错误喵")
		return
	}
	// 验证密码模块
	if err := user.CheckPassword(input.Password); err != nil {
		ReturnMsg(c, http.StatusUnauthorized, "用户名或密码错误喵")
		return
	}
	// 生成 token 模块
	token, err := jwt.GenerateToken(user.UserID, user.Username)
	if err != nil {
		ReturnMsg(c, http.StatusInternalServerError, "token生成失败喵")
		return
	}
	response := models.AuthResponse{ // 返回响应模块
		UserID: user.UserID,
		Token:  token,
	}
	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"data": response,
		"msg":  "success",
	})
}
