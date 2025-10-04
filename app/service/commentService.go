package service

import (
	"time"

	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/model"
	"gorm.io/gorm"
)

// AddComment 添加评论
func AddComment(db *gorm.DB, comment *model.Comment) error {
	comment.CreatedAt = time.Now()
	return db.Create(comment).Error
}

// DeleteComment 删除评论
func DeleteComment(db *gorm.DB, commentID uint) error {
	return db.Delete(&model.Comment{}, commentID).Error
}

// ListComments 获取某个表白的所有评论，附带用户信息
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
