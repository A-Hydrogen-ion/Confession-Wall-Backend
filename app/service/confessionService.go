package service

import (
	"time"

	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/model"
	"gorm.io/gorm"
)

// 创建表白
// 在 service 层通过参数 db *gorm.DB 传递数据库连接
func CreateConfession(db *gorm.DB, confession *model.Confession) error {
	confession.PublishedAt = time.Now()
	confession.ChangedAt = time.Now()
	return db.Create(confession).Error
}

// 获取某用户的所有表白
func GetAllConfessions(db *gorm.DB, userID uint) ([]model.Confession, error) {
	var confessions []model.Confession
	err := db.Where("user_id = ?", userID).Find(&confessions).Error
	return confessions, err
}

// 获取社区表白
func ListPublicConfessions(db *gorm.DB, currentUserID uint) ([]model.Confession, error) {
	var blockedIDs []uint
	var blockedByIDs []uint
	// 当前用户拉黑的
	db.Model(&model.Block{}).Where("user_id = ?", currentUserID).Pluck("blocked_id", &blockedIDs)
	// 拉黑了当前用户的
	db.Model(&model.Block{}).Where("blocked_id = ?", currentUserID).Pluck("user_id", &blockedByIDs)
	// 合并两个列表
	excludeIDs := append(blockedIDs, blockedByIDs...)
	var confessions []model.Confession
	query := db.Where("private = ?", false) //不展示私密表白
	// 如果有需要排除的用户ID，则添加条件
	if len(excludeIDs) > 0 {
		query = query.Where("user_id NOT IN ?", excludeIDs)
	}
	err := query.Find(&confessions).Error
	return confessions, err
}

// 修改表白
func UpdateConfession(db *gorm.DB, confessionID uint, newContent string, newImages []string) error {
	var confession model.Confession
	if err := db.First(&confession, confessionID).Error; err != nil {
		return err
	}
	confession.Content = newContent
	confession.Images = newImages
	confession.ChangedAt = time.Now()
	return db.Save(&confession).Error
}

// 删除表白
func DeleteConfession(db *gorm.DB, confessionID uint) error {
	return db.Delete(&model.Confession{}, confessionID).Error
}

// 根据ID获取单条表白
func GetConfessionByID(db *gorm.DB, confessionID uint) (model.Confession, error) {
	var confession model.Confession
	err := db.First(&confession, confessionID).Error
	return confession, err
}

// 获取某用户的所有表白（排除黑名单）
func GetUserConfessions(db *gorm.DB, targetUserID uint, currentUserID uint) ([]model.Confession, error) {
	var blockedIDs []uint
	var blockedByIDs []uint
	//熟悉的配方…………
	// 当前用户拉黑的
	db.Model(&model.Block{}).Where("user_id = ?", currentUserID).Pluck("blocked_id", &blockedIDs)
	// 拉黑了当前用户的
	db.Model(&model.Block{}).Where("blocked_id = ?", currentUserID).Pluck("user_id", &blockedByIDs)
	// 合并两个列表
	excludeIDs := append(blockedIDs, blockedByIDs...)

	var confessions []model.Confession
	query := db.Where("user_id = ? AND private = ?", targetUserID, false) // 排除私密表白
	if len(excludeIDs) > 0 {
		query = query.Where("user_id NOT IN ?", excludeIDs)
	}
	err := query.Find(&confessions).Error
	return confessions, err
}

// 添加评论
func AddComment(db *gorm.DB, comment *model.Comment) error {
	comment.CreatedAt = time.Now()
	return db.Create(comment).Error
}

// 删除评论
func DeleteComment(db *gorm.DB, commentID uint) error {
	return db.Delete(&model.Comment{}, commentID).Error
}

// 获取某个表白的所有评论，附带用户信息
func ListComments(db *gorm.DB, confessionID uint) ([]model.Comment, error) {
	var comments []model.Comment
	err := db.Preload("User"). // 关联用户信息
					Where("confession_id = ?", confessionID). //查询对应的表白
					Order("created_at ASC").
					Find(&comments).Error

	if err != nil {
		return nil, err
	}

	return comments, nil
}

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
