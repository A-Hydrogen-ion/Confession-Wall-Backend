package service

import (
	"fmt"
	"log"
	"strings"

	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/model"
	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/config/database"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService() *UserService { // 检查数据库指针函数
	if database.DB == nil {
		log.Println("警告: database.DB 为 nil，UserService 将无法工作")
	}
	return &UserService{db: database.DB}
}
func (s *UserService) CheckUsernameExists(username string) (bool, error) { // 检查用户名是否存在
	if s.db == nil { //数据库检查
		return false, fmt.Errorf("数据库连接未初始化")
	}
	var count int64 //查询用户名
	err := s.db.Model(&model.User{}).
		Where("username = ?", username).
		Count(&count).Error

	if err != nil { //其他错误
		log.Printf("数据库查询错误: %v", err)
		return false, fmt.Errorf("系统繁忙，请稍后重试")
	}

	return count > 0, nil
}
func (s *UserService) CreateUser(user *model.User) error { // 创建用户
	var count int64 // 唯一性检查
	if err := s.db.Model(&model.User{}).Where("username = ?", user.Username).Count(&count).Error; err != nil {
		return fmt.Errorf("检查用户名失败: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("用户名已存在")
	}
	if user.Nickname != "" {
		if err := s.db.Model(&model.User{}).Where("nickname = ?", user.Nickname).Count(&count).Error; err != nil {
			return fmt.Errorf("检查昵称失败: %w", err)
		}
		if count > 0 {
			return fmt.Errorf("昵称已存在")
		}
	}
	if err := s.db.Create(user).Error; err != nil { // 创建用户（Hook 会自动处理默认值）
		if strings.Contains(err.Error(), "Duplicate entry") {
			return fmt.Errorf("用户名或昵称已存在")
		}
		return fmt.Errorf("创建用户失败: %w", err)
	}

	return nil
}
func (s *UserService) GetUserByUsername(username string) (*model.User, error) { // 根据用户名获取用户
	var user model.User
	result := s.db.Where("username = ?", username).First(&user) //查找用户
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}
