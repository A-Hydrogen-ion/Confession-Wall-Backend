package service

import (
	"time"

	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/model"
	"gorm.io/gorm"
)

// CreateConfession 创建表白
// 在 service 层通过参数 db *gorm.DB 传递数据库连接
func CreateConfession(db *gorm.DB, confession *model.Confession) error {
	confession.ChangedAt = time.Now()
	return db.Create(confession).Error
}

// ListPublicConfessions 获取社区表白（带分页）
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

// UpdateConfession 修改表白
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

// DeleteConfession 删除表白
func DeleteConfession(db *gorm.DB, confessionID uint) error {
	// 先删除评论
	if err := db.Where("confession_id = ?", confessionID).Delete(&model.Comment{}).Error; err != nil {
		return err
	}
	// 再删除表白
	return db.Delete(&model.Confession{}, confessionID).Error
}

// GetConfessionByID 根据ID获取单条表白
func GetConfessionByID(db *gorm.DB, confessionID uint) (model.Confession, error) {
	var confession model.Confession
	err := db.First(&confession, confessionID).Error
	return confession, err
}

// GetConfessionsByID 根据多个ID获取表白列表（只给热度榜使用）
func GetConfessionsByID(db *gorm.DB, ids []uint) ([]model.Confession, error) {
	var confessions []model.Confession
	err := db.Where("id IN ?", ids).Find(&confessions).Error
	return confessions, err
}

// GetUserConfessions 获取某用户的所有表白（排除黑名单，带分页）
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
