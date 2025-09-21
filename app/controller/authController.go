package controller

import (
	"fmt"
	"net/http"

	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/jwt"
	models "github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/model"
	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/service"

	//"gorm.io/gorm"
	"github.com/gin-gonic/gin"
)

func getUserService() *service.UserService {
	return service.NewUserService()
}
func Register(c *gin.Context) {
	var input models.RegisterRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查用户名是否已存在
	exists, err := getUserService().CheckUsernameExists(input.Username)
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
		Password: input.Password, // 这里会映射到数据库的 password_hash 列
		Nickname: input.Nickname, // 直接使用输入的昵称
	}

	if err := service.NewUserService().CreateUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败"})
		return
	}
	// 生成 token
	token, err := jwt.GenerateToken(user.UserID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"token": token}, "msg": "success"})
}

func Login(c *gin.Context) {
	var input models.LoginRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userService := getUserService()
	// 查找用户
	user, err := userService.GetUserByUsername(input.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 验证密码
	if err := user.CheckPassword(input.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
	}
	// 生成 token
	token, err := jwt.GenerateToken(user.UserID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	//创建会话
	response := models.AuthResponse{
		UserID: user.UserID,
		Token:  token,
	}
	//返回参数
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": response, "msg": "success"})
}
