package controller

import (
	"net/http"

	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/model"
	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BlockController struct {
	DB *gorm.DB
}

func NewBlockController(db *gorm.DB) *BlockController {
	return &BlockController{DB: db}
}

// BlockUser 添加黑名单
func (blockController *BlockController) BlockUser(c *gin.Context) {
	userID := checkUserByID(c, "你还没有登录喵，你要拉黑全世界吗？") //使用辅助函数判断
	if userID == 0 {
		return
	}

	// 从请求参数获取 blocked_id，并调用queryUint辅助函数转换为 uint
	blockedID, err := QueryUint(c, "blocked_id")
	if err != nil {
		respondJSON(c, http.StatusBadRequest, "blocked_id 参数格式错误喵~", nil)
		return
	}
	// 调用 service
	if err := service.BlockUser(blockController.DB, userID, blockedID); err != nil {
		respondJSON(c, http.StatusInternalServerError, "添加黑名单失败喵~", nil)
		return
	}
	// 拉黑自己提示但不阻止
	if userID == blockedID {
		respondJSON(c, http.StatusOK, "你居然把自己拉黑了喵~（已执行）（服务器娘坏笑）", nil)
	} else {
		respondJSON(c, http.StatusOK, "这个用户被你拉进小黑屋了喵~", nil)
	}
}

// UnblockUser 将用户移除黑名单
func (blockController *BlockController) UnblockUser(c *gin.Context) {
	userID := checkUserByID(c, "你还没有登录喵，服务器娘不知道你拉黑了谁~") //使用辅助函数判断
	if userID == 0 {
		return
	}
	blockedIDStr := c.Query("blocked_id") // 从请求参数获取 blocked_id
	if blockedIDStr == "" {
		respondJSON(c, http.StatusBadRequest, "缺少参数 blocked_id 喵~", nil)
		return
	}
	blockedID, err := GetUintParam(c, "blocked_id") //参数格式转换
	if err != nil {
		respondJSON(c, http.StatusBadRequest, "blocked_id 参数格式错误喵~", nil)
		return
	}
	if err := service.UnblockUser(blockController.DB, userID, blockedID); err != nil { //调用service
		respondJSON(c, http.StatusInternalServerError, "移除黑名单失败喵~", nil)
		return
	}
	respondJSON(c, http.StatusOK, "已成功将该用户移出黑名单了喵~", nil)
}

// GetBlockedUsers 获取当前用户拉黑的用户列表
func (blockController *BlockController) GetBlockedUsers(c *gin.Context) {
	userID := checkUserByID(c, "你还没有登录喵，服务器娘不知道你的小黑屋") //使用辅助函数判断
	if userID == 0 {
		return
	}
	// 查 blocked_id 列表
	var blockedIDs []uint
	if err := blockController.DB.Model(&model.Block{}).
		Where("user_id = ?", userID).
		Pluck("blocked_id", &blockedIDs).Error; err != nil {
		respondJSON(c, http.StatusInternalServerError, "查询黑名单失败喵~", nil)
		return
	}
	// 查询对应用户信息
	var users []model.User //声明不赋值的切片默认为nil不必在为空时单独处理
	if len(blockedIDs) > 0 {
		if err := blockController.DB.Where("user_id IN ?", blockedIDs).Find(&users).Error; err != nil {
			respondJSON(c, http.StatusInternalServerError, "查询黑名单用户信息失败喵~", nil)
			return
		}
	}
	respondJSON(c, http.StatusOK, "success", gin.H{"blocked_users": users})
}

