package controller

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/model"
	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/service"
	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/config/database"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
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

	// 未登录用户
	if userID == uint(0) {
		c.JSON(http.StatusOK, gin.H{
			"user_id":  0,
			"nickname": "匿名用户",
			"avatar":   "/uploads/avatars/default.png",
		})
		return
	}
	// 查询用户信息
	var profile model.User
	result := database.DB.First(&profile, userID)
	if result.Error != nil {
		ReturnMsg(c, http.StatusBadRequest, "没有找到这个用户啊喵")
		return
	}
	// 如果 Avatar 字段为空，使用默认头像
	avatarURL := "/uploads/avatars/default.png"
	if profile.Avatar != "" {
		avatarURL = "/uploads/avatars/" + profile.Avatar
	}

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
		ReturnMsg(c, http.StatusUnauthorized, "用户没有登陆啊喵")
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
		ReturnMsg(c, http.StatusBadRequest, "没有找到这个用户啊喵")
		return
	}
	if isUsernameExist := authController.IsUserExist(c, input.Username); !isUsernameExist {
		return
	}
	if isNicknameExist := authController.IsUserExist(c, input.Nickname); !isNicknameExist {
		return
	}
	profile.Nickname = input.Nickname // 更新用户信息
	profile.Avatar = input.Avatar
	if err := database.DB.Save(&profile).Error; err != nil { // 保存更新
		// 优先尝试通过 mysql 错误类型检测
		var me *mysql.MySQLError
		if errors.As(err, &me) && me.Number == 1062 {
			// 根据错误信息判断具体字段，优先返回昵称冲突
			if strings.Contains(err.Error(), "idx_users_nickname") || strings.Contains(err.Error(), "nickname") {
				ReturnMsg(c, http.StatusBadRequest, "该昵称已被占用喵")
				return
			}
			ReturnMsg(c, http.StatusBadRequest, "存在重复的唯一字段")
			return
		}
		// 回退到字符串匹配（在某些情况下，错误未直接暴露为 mysql.MySQLError）
		if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "ER_DUP_ENTRY") {
			if strings.Contains(err.Error(), "idx_users_nickname") || strings.Contains(err.Error(), "nickname") {
				ReturnMsg(c, http.StatusBadRequest, "该昵称已被占用喵")
				return
			}
			ReturnMsg(c, http.StatusBadRequest, "存在重复的唯一字段")
			return
		}

		fmt.Print(err.Error())
		ReturnMsg(c, http.StatusInternalServerError, "存在重复的唯一字段")
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
		ReturnMsg(c, http.StatusUnauthorized, "你还没有登录喵~")
		return
	}

	path, err := service.UploadAvatar(c, userID.(uint))
	if err != nil {
		ReturnError400(c, err)
		return
	}

	// 更新用户资料表 avatar 字段
	var user model.User
	if err := userController.DB.First(&user, userID).Error; err != nil {
		ReturnMsg(c, http.StatusInternalServerError, "获取用户失败，服务器娘不知道你是谁喵")
		return
	}
	user.Avatar = path
	if err := userController.DB.Save(&user).Error; err != nil {
		ReturnMsg(c, http.StatusInternalServerError, "服务器娘宕机了，她不小心把你的头像弄丢了")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":   http.StatusOK,
		"avatar": path,
		"msg":    "success",
	})
}
