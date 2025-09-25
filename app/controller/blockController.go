package controller

import (
	"net/http"
	"strconv"

	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/model"
	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/service"
	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/config/database"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BlockController struct {
	DB *gorm.DB
}

func NewBlockController(db *gorm.DB) *BlockController {
	return &BlockController{DB: db}
}

// 添加黑名单
func (blockController *BlockController) BlockUser(c *gin.Context) {
	// 获取当前用户 ID
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "你还没有登录喵，你要拉黑全世界吗？"})
		return
	}
	userID := userIDValue.(uint)

	// 从请求参数获取 blocked_id，并调用queryUint辅助函数转换为 uint
	blockedIDStr := c.Query("blocked_id")
	if blockedIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少参数 blocked_id 喵~"})
		return
	}
	blockedID64, err := strconv.ParseUint(blockedIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "blocked_id 参数格式错误喵~"})
		return
	}
	blockedID := uint(blockedID64)

	// 调用 service
	if err := service.BlockUser(database.DB, userID, blockedID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "添加黑名单失败喵~"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "这个用户被你拉进小黑屋了喵~"})
}

// 将用户移除黑名单
func (blockController *BlockController) UnblockUser(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "你还没有登录喵，服务器娘不知道你拉黑了谁"})
		return
	}
	userID := userIDValue.(uint)

	blockedIDStr := c.Query("blocked_id")
	if blockedIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少参数 blocked_id 喵~"})
		return
	}
	blockedID64, err := strconv.ParseUint(blockedIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "blocked_id 参数格式错误喵~"})
		return
	}
	blockedID := uint(blockedID64)

	if err := service.UnblockUser(database.DB, userID, blockedID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "移除黑名单失败喵~"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "已成功将该用户移出黑名单了喵~"})
}

// 获取当前用户拉黑的用户列表
func (blockController *BlockController) GetBlockedUsers(c *gin.Context) {
	// 获取当前用户 ID
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "你还没有登录喵，服务器娘不知道你的小黑屋"})
		return
	}
	userID := userIDValue.(uint)

	// 查 blocked_id 列表
	var blockedIDs []uint
	if err := blockController.DB.Model(&model.Block{}).
		Where("user_id = ?", userID).
		Pluck("blocked_id", &blockedIDs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询黑名单失败喵~"})
		return
	}

	// 如果没拉黑任何人
	if len(blockedIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"blocked_users": []model.User{}})
		return
	}

	// 查询对应用户信息
	var users []model.User
	if err := blockController.DB.Where("user_id IN ?", blockedIDs).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询黑名单用户信息失败喵~"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"blocked_users": users})
}
