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
func (s *UserService) checkFieldExists(fieldName, value string) (bool, error) { //检查特定字段是否存在
	if s.db == nil {
		return false, fmt.Errorf("数据库连接未初始化")
	}
	var count int64
	err := s.db.Model(&model.User{}).
		Where(fieldName+" = ?", value).
		Count(&count).Error
	if err != nil {
		log.Printf("数据库查询错误[字段%s]: %v", fieldName, err)
		return false, fmt.Errorf("系统繁忙，请稍后重试")
	}
	return count > 0, nil
}
func (s *UserService) CheckUsernameExists(username string) (bool, error) { //检查用户名是否存在

	return s.checkFieldExists("username", username)
}
func (s *UserService) CheckNicknameExists(nickname string) (bool, error) { //检查昵称是否存在
	return s.checkFieldExists("nickname", nickname)
}
func (s *UserService) CreateUser(user *model.User) error {
	// 使用 CheckUsernameExists 检查用户名
	exists, err := s.CheckUsernameExists(user.Username)
	if err != nil {
		return fmt.Errorf("检查用户名失败: %w", err)
	}
	if exists {
		return fmt.Errorf("用户名已存在")
	}
	// 如果设置了昵称，使用 CheckNicknameExists 检查昵称
	if user.Nickname != "" {
		exists, err := s.CheckNicknameExists(user.Nickname)
		if err != nil {
			return fmt.Errorf("检查昵称失败: %w", err)
		}
		if exists {
			return fmt.Errorf("昵称已存在")
		}
	}
	// 创建用户
	if err := s.db.Create(user).Error; err != nil {
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
