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

type ChangePassword struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=100"`
}

func ReturnError400(c *gin.Context, err error) error { // 自定义返回错误函数
	respondJSON(c, http.StatusBadRequest, err.Error(), nil)
	return nil
}

func ReturnMsg(c *gin.Context, state int, msg string) error { //自定义返回消息函数
	respondJSON(c, state, msg, nil)
	return nil
}

// 创建AuthController
func NewAuthController(db *gorm.DB) *AuthController {
	return &AuthController{
		userService: service.NewUserService(db),
		db:          db,
	}
}

// 输入要求控制函数
func checkInputRequirement(c *gin.Context, validationErrors validator.ValidationErrors) {
	// 只返回第一个校验错误，消息尽量友好
	for _, fe := range validationErrors {
		switch fe.Tag() {
		case "required":
			ReturnMsg(c, http.StatusBadRequest, fmt.Sprintf("%s 是必填字段", fe.Field()))
			return
		case "min":
			ReturnMsg(c, http.StatusBadRequest, fmt.Sprintf("%s 长度不能少于 %s 个字符", fe.Field(), fe.Param()))
			return
		case "max":
			ReturnMsg(c, http.StatusBadRequest, fmt.Sprintf("%s 长度不能超过 %s 个字符", fe.Field(), fe.Param()))
			return
		default:
			// 兜底：使用验证器提供的错误信息
			ReturnMsg(c, http.StatusBadRequest, fe.Error())
			return
		}
	}
}

// 存在验证函数
func (authController *AuthController) checkExists(c *gin.Context, input string,
	checkFunc func(string) (bool, error), errorMsg string) bool {
	exists, err := checkFunc(input)
	if err != nil {
		log.Printf("检查%s存在性时发生错误: %v", input, err)
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

// 注册主函数
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
		// token 生成失败：返回 500 更符合语义，但保持原有行为为 400
		ReturnMsg(c, http.StatusBadRequest, "token生成失败了喵")
		return
	}
	respondJSON(c, http.StatusOK, "success", gin.H{"token": token})
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
	respondJSON(c, http.StatusOK, "success", response)
}

func (authController *AuthController) ChangePassword(c *gin.Context) {
	var req ChangePassword
	// 绑定输入的JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		// 判断输入
		if ok {
			checkInputRequirement(c, validationErrors)
			return
		}
		ReturnError400(c, err)
		return
	}

	// 获取当前登录用户ID
	userIDValue, exists := c.Get("user_id")
	if !exists {
		ReturnMsg(c, http.StatusUnauthorized, "你还没有登录喵，服务器娘不知道你是谁")
		return
	}
	userID := userIDValue.(uint)

	// 查找用户
	user, err := authController.userService.GetUserByID(userID)
	if err != nil {
		ReturnMsg(c, http.StatusBadRequest, "你要修改的用户不存在喵~")
		return
	}

	// 验证旧密码
	if err := user.CheckPassword(req.OldPassword); err != nil {
		ReturnMsg(c, http.StatusUnauthorized, "旧密码输入错误了喵~")
		return
	}

	// 更新新密码
	if err := authController.userService.UpdatePassword(user, req.NewPassword); err != nil {
		ReturnError400(c, err)
		return
	}

	ReturnMsg(c, http.StatusOK, "密码修改成功了喵~")
}

func (authController *AuthController) ChangePassword(c *gin.Context) {
	var req ChangePassword
	// 绑定输入的JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		// 判断输入
		if ok {
			checkInputRequirement(c, validationErrors)
			return
		}
		ReturnError400(c, err)
		return
	}

	// 获取当前登录用户ID
	userIDValue, exists := c.Get("user_id")
	if !exists {
		ReturnMsg(c, http.StatusUnauthorized, "你还没有登录喵，服务器娘不知道你是谁")
		return
	}
	userID := userIDValue.(uint)

	// 查找用户
	user, err := authController.userService.GetUserByID(userID)
	if err != nil {
		ReturnMsg(c, http.StatusBadRequest, "你要修改的用户不存在喵~")
		return
	}

	// 验证旧密码
	if err := user.CheckPassword(req.OldPassword); err != nil {
		ReturnMsg(c, http.StatusUnauthorized, "旧密码输入错误了喵~")
		return
	}

	// 更新新密码
	if err := authController.userService.UpdatePassword(user, req.NewPassword); err != nil {
		ReturnError400(c, err)
		return
	}

	ReturnMsg(c, http.StatusOK, "密码修改成功了喵~")
}
