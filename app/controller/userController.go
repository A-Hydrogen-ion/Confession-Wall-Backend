package controller

import (
	"fmt"
	"net/http"

	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/model"
	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/service"
	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/config/database"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserController struct {
	DB *gorm.DB
}

func NewUserController(db *gorm.DB) *UserController {
	return &UserController{DB: db}
}

func (userController *AuthController) GetMyProfile(c *gin.Context) {
	// 获取从中间件设置的user_id
	userID, _ := c.Get("user_id")
	var profile model.User
	// 如果 userID 为 0 或者avatars没有该用户的头像，返回一个默认头像和昵称
	switch {
	case userID == uint(0):
		// 未登录用户
		c.JSON(http.StatusOK, gin.H{
			"user_id":  0,
			"nickname": "匿名用户",
			"avatar":   "uploads/avatars/default.png",
		})
		return
	case database.DB.First(&profile, userID).Error != nil:
		// 用户没有上传头像
		c.JSON(http.StatusOK, gin.H{
			"avatar": "uploads/avatars/default.png",
		})
		return
	}
	result := database.DB.First(&profile, userID)
	if result.Error != nil {
		ReturnMsg(c, 400, "没有找到这个用户啊喵")
		return
	}
	avatarURL := "/uploads/avatars/" + profile.Avatar
	c.JSON(http.StatusOK, gin.H{ //返回用户的信息
		"user_id":  profile.UserID,
		"nickname": profile.Nickname,
		"avatar":   avatarURL,
	})
}

// 更新用户处理
func (authController *AuthController) UpdateUserProfile(c *gin.Context) {
	// 获取从中间件设置的 user_id
	userID, exists := c.Get("user_id")
	if !exists {
		ReturnMsg(c, 401, "用户没有登陆啊喵")
		return
	}
	var input model.User // 绑定请求的 JSON 数据
	if err := c.ShouldBindJSON(&input); err != nil {
		ReturnError400(c, err)
		return
	}
	var profile model.User // 查找用户
	result := database.DB.First(&profile, userID)
	if result.Error != nil {
		ReturnMsg(c, 400, "没有找到这个用户啊喵")
		return
	}
	isUsernameExist := authController.IsUserExist(c, input.Username)
	if !isUsernameExist {
		return
	}
	isNicknameExist := authController.IsUserExist(c, input.Nickname)
	if !isNicknameExist {
		return
	}
	profile.Nickname = input.Nickname // 更新用户信息
	profile.Avatar = input.Avatar
	if err := database.DB.Save(&profile).Error; err != nil { // 保存更新
		fmt.Print(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器娘宕机了，她不小心把你的信息弄丢了"})
		return
	}
	c.JSON(http.StatusOK, gin.H{ // 返回更新后的用户信息
		"msg":      "用户资料更新成功了喵",
		"user_id":  profile.UserID,
		"nickname": profile.Nickname,
		"avatar":   profile.Avatar,
	})
}

// UploadAvatar 上传用户头像
func (userController *UserController) UploadAvatar(c *gin.Context) {
	// 获取登录用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "你还没有登录喵~"})
		return
	}

	path, err := service.UploadAvatar(c, userID.(uint))
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 更新用户资料表 avatar 字段
	var user model.User
	if err := userController.DB.First(&user, userID).Error; err != nil {
		c.JSON(500, gin.H{"error": "获取用户失败，服务器娘不知道你是谁喵"})
		return
	}
	user.Avatar = path
	if err := userController.DB.Save(&user).Error; err != nil {
		c.JSON(500, gin.H{"error": "服务器娘宕机了，她不小心把你的头像弄丢了"})
		return
	}

	c.JSON(200, gin.H{"msg": "服务器娘收到你上传的头像了喵~", "avatar": path})
}
