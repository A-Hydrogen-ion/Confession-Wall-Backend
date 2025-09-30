package service

import (
	"time"

	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/model"
	"gorm.io/gorm"
)

// 创建表白
// 在 service 层通过参数 db *gorm.DB 传递数据库连接
func CreateConfession(db *gorm.DB, confession *model.Confession) error {
	confession.ChangedAt = time.Now()
	return db.Create(confession).Error
}

// 获取社区表白（带分页）
func ListPublicConfessions(db *gorm.DB, currentUserID uint, limit int, offset int) ([]model.Confession, error) {
	var blockedIDs []uint
	var blockedByIDs []uint
	// 当前用户拉黑的
	db.Model(&model.Block{}).Where("user_id = ?", currentUserID).Pluck("blocked_id", &blockedIDs)
	// 拉黑了当前用户的
	db.Model(&model.Block{}).Where("blocked_id = ?", currentUserID).Pluck("user_id", &blockedByIDs)
	// 合并两个列表
	excludeIDs := append(blockedIDs, blockedByIDs...)
	var confessions []model.Confession
	now := time.Now()
	query := db.Where("private = ? AND publishedAt <= ?", false, now) // 不展示私密表白和未来发布的表白
	// 如果有需要排除的用户ID，则添加条件
	if len(excludeIDs) > 0 {
		query = query.Where("user_id NOT IN ?", excludeIDs)
	}
	err := query.Limit(limit).Offset(offset).Find(&confessions).Error
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
	// 先删除评论
	if err := db.Where("confession_id = ?", confessionID).Delete(&model.Comment{}).Error; err != nil {
		return err
	}
	// 再删除表白
	return db.Delete(&model.Confession{}, confessionID).Error
}

// 根据ID获取单条表白
func GetConfessionByID(db *gorm.DB, confessionID uint) (model.Confession, error) {
	var confession model.Confession
	err := db.First(&confession, confessionID).Error
	return confession, err
}

// 获取某用户的所有表白（排除黑名单，带分页）
func GetUserConfessions(db *gorm.DB, targetUserID uint, currentUserID uint, limit int, offset int) ([]model.Confession, error) {
	var blockedIDs []uint
	var blockedByIDs []uint
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
	// 只有自己能看到未到发布时间的表白，别人只能看到已发布的
	now := time.Now()
	if targetUserID != currentUserID {
		query = query.Where("publishedAt <= ?", now)
	}
	err := query.Limit(limit).Offset(offset).Find(&confessions).Error
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
