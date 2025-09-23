package model

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User相关
// 定义User数据类型
type User struct {
	UserID    uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"username"`
	Nickname  string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"nickname"`
	Avatar    string    `gorm:"column:avatar;not null" json:"avatar"`
	Password  string    `gorm:"column:password_hash;not null" json:"-"`
	CreatedAt time.Time `gorm:"column:createdAt;not null" json:"createdAt"`
	UpdateAt  time.Time `gorm:"column:updateAt;not null" json:"updateAt"`
}

type Confession struct {
	ID          uint      `gorm:"primaryKey"`
	UserID      uint      `gorm:"not null" json:"userId"`
	Content     string    `gorm:"type:text;not null" json:"content"`
	Images      []string  `gorm:"type:json" json:"images"`
	Anonymous   bool      `gorm:"not null" json:"Anonymous"`
	Private     bool      `gorm:"not null" json:"Private"`
	PublishedAt time.Time `gorm:"column:publishedAt;not null" json:"publishedAt"`
	ChangedAt   time.Time `gorm:"column:changedAt;not null" json:"changedAt"`
}

// 创建用户前哈希密码钩子
func (u *User) BeforeSave(tx *gorm.DB) error {
	if len(u.Password) > 0 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.Password = string(hashedPassword)
	}
	return nil
}

// 创建用户前创建检查钩子
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// 设置默认昵称
	if u.Nickname == "" {
		u.Nickname = u.Username
	}

	// 设置时间戳
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now()
	}
	if u.UpdateAt.IsZero() {
		u.UpdateAt = time.Now()
	}

	return nil
}

// 检查密码是否正确
func (u *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}

// 注册相关
type RegisterRequest struct {
	Username string `json:"username"   binding:"required,min=3,max=10"`
	Nickname string `json:"Nickname"   binding:"required,min=2"`
	Password string `json:"password"   binding:"required,min=8,max=16"`
} //RegisterRequest结构体将用于处理用户注册请求的数据绑定和验证
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
} //LoginRequest结构体将用于处理用户登录请求的数据绑定和验证
type AuthResponse struct {
	UserID uint   `json:"user_id"`
	Token  string `json:"token"`
} //返回结构体
type Response struct {
	Data string `json:"data"`
	Msg  string `json:"msg"`
}
