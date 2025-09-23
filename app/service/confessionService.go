package service

import (
	"errors"
	"time"

	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/model"
	"gorm.io/gorm"
)

// 创建表白
// 在 service 层通过参数 db *gorm.DB 传递数据库连接
func CreateConfession(db *gorm.DB, confession *model.Confession) error {
	if len(confession.Images) > 9 {
		return errors.New("最多上传9张图片哦喵~")
	}
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

// 获取社区表白（过滤私密、拉黑等逻辑可在此扩展）
func ListPublicConfessions(db *gorm.DB) ([]model.Confession, error) {
	var confessions []model.Confession
	err := db.Where("is_private = ?", false).Find(&confessions).Error
	return confessions, err
}

// 修改表白
func UpdateConfession(db *gorm.DB, confessionID uint, newContent string, newImages []string) error {
	var confession model.Confession
	if err := db.First(&confession, confessionID).Error; err != nil {
		return err
	}
	if len(newImages) > 9 {
		return errors.New("最多上传9张图片")
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

//依然没有写完，只写了个框架
