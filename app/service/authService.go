package service

import (
	"fmt"
	"log"

	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/model"
	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/config/database"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

// 检查数据库指针函数
func NewUserService() *UserService {
	if database.DB == nil {
		log.Println("警告: database.DB 为 nil，UserService 将无法工作")
	}
	return &UserService{db: database.DB}
}

// 检查用户名是否存在
func (s *UserService) CheckUsernameExists(username string) (bool, error) {
	if s.db == nil {
		return false, fmt.Errorf("数据库连接未初始化 (s.db is nil)")
	}
	var user model.User
	result := s.db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return false, nil
		}
		fmt.Println(result.Error)
		return false, result.Error
	}
	return true, nil
}

// 创建用户
func (s *UserService) CreateUser(user *model.User) error {
	result := s.db.Create(user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// 根据用户名获取用户
func (s *UserService) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	result := s.db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}
