package controller

import (
	"fmt"
	"net/http"

	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/model"
	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/config/database"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserController struct {
	DB *gorm.DB
}

func NewUserController() *UserController {
	return &UserController{}
}

func (userController *AuthController) GetMyProfile(c *gin.Context) {
	// 获取从中间件设置的user_id
	userID, _ := c.Get("user_id")
	// 如果 userID 为 0，返回一个默认头像和昵称
	if userID == uint(0) {
		c.JSON(http.StatusOK, gin.H{
			"user_id":  0,
			"nickname": "匿名用户",
			"avatar":   "～/avatar/default.png",
		})
		return
	}
	var profile model.User
	result := database.DB.First(&profile, userID)
	if result.Error != nil {
		ReturnMsg(c, 400, "没有找到这个用户啊喵")
		return
	}
	c.JSON(http.StatusOK, gin.H{ //返回用户的信息
		"user_id":  profile.UserID,
		"nickname": profile.Nickname,
		"avatar":   profile.Avatar,
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
