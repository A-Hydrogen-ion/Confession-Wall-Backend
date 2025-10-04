package service

import (
	"fmt"
	"log"
	"strings"

	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/model"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

// 通过NewUserService 创建一个新的 UserService 实例
// 这样做可以使得之后测试数据库和实际使用的数据库不互相干扰
func NewUserService(db *gorm.DB) *UserService {
	if db == nil {
		log.Println("警告: 传入的 db 为 nil，UserService 将无法工作")
	}
	return &UserService{db: db}
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
		// 记录底层错误，返回通用错误消息给调用方
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
		// MySQL 的重复键错误在不同驱动返回中可能包含不同文本，简单包含判断以提供友好信息
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

// 根据用户ID获取用户
func (s *UserService) GetUserByID(userID uint) (*model.User, error) {
	var user model.User
	result := s.db.First(&user, userID)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// 更新用户密码
func (s *UserService) UpdatePassword(user *model.User, newPassword string) error {
	user.Password = newPassword // User 的 BeforeSave 钩子会自动 hash 密码，不需要单独hash
	if err := s.db.Save(user).Error; err != nil {
		return err
	}
	return nil
}
