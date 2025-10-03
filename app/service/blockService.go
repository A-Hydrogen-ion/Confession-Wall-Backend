package service

import (
	"time"

	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/model"
	"gorm.io/gorm"
)

// 处理黑名单
func BlockUser(db *gorm.DB, userID uint, blockedID uint) error {
	block := &model.Block{
		UserID:    userID,
		BlockedID: blockedID,
		CreatedAt: time.Now(),
	}
	return db.Create(block).Error
}

// 移除黑名单
func UnblockUser(db *gorm.DB, userID uint, blockedID uint) error {
	return db.Where("user_id = ? AND blocked_id = ?", userID, blockedID).Delete(&model.Block{}).Error
}
