package controller

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/model"
	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/service"
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
		respondJSON(c, http.StatusOK, "", gin.H{"user_id": 0, "nickname": "匿名用户", "avatar": "/uploads/avatars/default.png"})
		return
	}
	// 查询用户信息
	var profile model.User
	// 使用注入到 AuthController 的 db 实例（接收者名为 userController）
	result := userController.db.First(&profile, userID)
	if result.Error != nil {
		respondJSON(c, http.StatusBadRequest, "没有找到这个用户啊喵", nil)
		return
	}
	// 如果 Avatar 字段为空，使用默认头像
	avatarURL := "/uploads/avatars/default.png"
	if profile.Avatar != "" {
		avatarURL = "/uploads/avatars/" + profile.Avatar
	}

	respondJSON(c, http.StatusOK, "", gin.H{"user_id": profile.UserID, "nickname": profile.Nickname, "avatar": avatarURL})
}

// 更新用户处理，好悬差点没改死我
func (authController *AuthController) UpdateUserProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		respondJSON(c, http.StatusUnauthorized, "用户没有登陆啊喵", nil)
		return
	}
	//绑定输入的model
	var input struct {
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
		Username string `json:"username"`
	}
	// 绑定输入的JSON
	if err := c.ShouldBindJSON(&input); err != nil {
		ReturnError400(c, err)
		return
	}
	// 查询用户信息
	var profile model.User
	if err := authController.db.First(&profile, userID).Error; err != nil {
		respondJSON(c, http.StatusBadRequest, "没有找到这个用户啊喵", nil)
		return
	}
	// 唯一性校验抽离
	if err := checkUnique(authController, c, &input, &profile); err != nil {
		return
	}
	// 更新字段检查
	if input.Nickname != "" {
		profile.Nickname = input.Nickname
	}
	if input.Avatar != "" {
		profile.Avatar = input.Avatar
	}
	//处理唯一性错误
	if err := authController.db.Save(&profile).Error; err != nil {
		processError(c, err)
		return
	}
	respondJSON(c, http.StatusOK, "用户资料更新成功了喵", gin.H{"user_id": profile.UserID, "nickname": profile.Nickname, "avatar": profile.Avatar})
}

// 唯一性校验函数
func checkUnique(authController *AuthController, c *gin.Context, input interface{}, profile *model.User) error {
	in, ok := input.(*struct {
		Nickname string
		Avatar   string
		Username string
	})
	if !ok {
		return nil
	}
	// 用户名和昵称唯一性检查
	if in.Username != "" && in.Username != profile.Username {
		if !authController.IsUserExist(c, in.Username) {
			return errors.New("用户名已存在")
		}
	}
	if in.Nickname != "" && in.Nickname != profile.Nickname {
		if !authController.IsNicknameExist(c, in.Nickname) {
			return errors.New("昵称已存在")
		}
	}
	return nil
}

// 唯一性错误处理
func processError(c *gin.Context, err error) {
	var me *mysql.MySQLError
	if errors.As(err, &me) && me.Number == 1062 {
		// 处理昵称保证其唯一性
		if strings.Contains(err.Error(), "nickname") {
			ReturnMsg(c, http.StatusBadRequest, "该昵称已被占用喵")
			return
		}
		ReturnMsg(c, http.StatusBadRequest, "存在重复的唯一字段")
		return
	}
	if strings.Contains(err.Error(), "Duplicate entry") {
		if strings.Contains(err.Error(), "nickname") {
			ReturnMsg(c, http.StatusBadRequest, "该昵称已被占用喵")
			return
		}
		ReturnMsg(c, http.StatusBadRequest, "存在重复的唯一字段")
		return
	}
	fmt.Print(err.Error())
	ReturnMsg(c, http.StatusInternalServerError, "存在重复的唯一字段")
}

// UploadAvatar 上传用户头像
func (userController *UserController) UploadAvatar(c *gin.Context) {
	// 获取登录用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		respondJSON(c, http.StatusUnauthorized, "你还没有登录喵~", nil)
		return
	}
	// 调用服务层处理上传的头像，服务层将对头像自动裁剪和压缩
	path, err := service.UploadAvatar(c, userID.(uint))
	if err != nil {
		ReturnError400(c, err)
		return
	}

	// 更新用户资料表 avatar 字段
	var user model.User
	if err := userController.DB.First(&user, userID).Error; err != nil {
		respondJSON(c, http.StatusInternalServerError, "获取用户失败，服务器娘不知道你是谁喵", nil)
		return
	}
	user.Avatar = path
	if err := userController.DB.Save(&user).Error; err != nil {
		respondJSON(c, http.StatusInternalServerError, "服务器娘宕机了，她不小心把你的头像弄丢了", nil)
		return
	}
	respondJSON(c, http.StatusOK, "success", gin.H{"avatar": path})
}
