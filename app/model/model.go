package model

import (
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User User相关
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

// Confession 表白数据类型
type Confession struct {
	ID          uint      `gorm:"primaryKey"`
	UserID      uint      `gorm:"not null" json:"userId"`
	Content     string    `gorm:"type:text;not null" json:"content"`
	Images      []string  `gorm:"type:json;serializer:json" json:"images"`
	Anonymous   bool      `gorm:"not null" json:"Anonymous"`
	Private     bool      `gorm:"not null" json:"Private"`
	ViewCount   uint      `gorm:"default:0" json:"viewCount"` // 浏览量
	LikeCount   uint      `gorm:"default:0" json:"likeCount"` // 点赞量
	PublishedAt time.Time `gorm:"column:publishedAt;not null" json:"publishedAt"`
	ChangedAt   time.Time `gorm:"column:changedAt;not null" json:"changedAt"`
}

// Comment 评论数据类型
type Comment struct {
	ID           uint      `gorm:"primaryKey"`
	UserID       uint      `gorm:"not null" json:"userId"`
	ConfessionID uint      `gorm:"not null;index" json:"confession_id"` // 来自表白数据类型的外键字段，以让评论和表白绑定在一起，同时在main.go中添加自动迁移来让gorm知道这个表结构和外键
	Content      string    `gorm:"type:text;not null" json:"content"`
	CreatedAt    time.Time `gorm:"column:createdAt;not null" json:"createdAt"`
	User         User      `gorm:"foreignKey:UserID;references:UserID" json:"user"` // 建立来自user的外键关系（GORM 会自动生成约束）

	Confession Confession `gorm:"foreignKey:ConfessionID;constraint:OnDelete:CASCADE;" json:"-"`
	// 建立外键关系（GORM 会自动生成约束）,默认情况下，任何模型的主键字段都是 ID，所以不需要加references来指向confession的ID
}

// Block 小黑屋数据类型
type Block struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null" json:"userId"`
	BlockedID uint      `gorm:"not null" json:"blockedID"`
	CreatedAt time.Time `gorm:"column:createdAt;not null" json:"createdAt"`
}

// RegisterRequest Register注册相关
type RegisterRequest struct {
	Username string `json:"username"   binding:"required,min=3,max=15"`
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
} // AuthResponse结构体用于登录成功后返回用户ID和JWT令牌
type Response struct {
	Data string `json:"data"`
	Msg  string `json:"msg"`
} // Response结构体用于统一API响应格式，包含数据和消息字段
// CreateConfessionRequest 表白创建请求相关
type CreateConfessionRequest struct {
	Content     string `form:"content" binding:"required"`
	Anonymous   bool   `form:"anonymous"`
	Private     bool   `form:"private"`
	PublishTime string `form:"publish_time"`
} // CreateConfessionRequest结构体用于处理创建表白请求的数据绑定和验证
type UpdateUserProfileRequest struct {
    Nickname string `json:"nickname"`
    Avatar   string `json:"avatar"`
    Username string `json:"username"`
} // UpdateUserProfileRequest结构体用于处理更新用户资料请求的数据绑定和验证
// BeforeSave 创建用户前哈希密码钩子
func (u *User) BeforeSave(tx *gorm.DB) error {
	if len(u.Password) == 0 || isBcryptHash(u.Password) { //如果已经hash过了则跳过hash
		return nil
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// 判断字符串是否看起来像 bcrypt 的 hash
func isBcryptHash(s string) bool {
	if len(s) != 60 {
		return false
	}
	return strings.HasPrefix(s, "$2a$") || strings.HasPrefix(s, "$2b$") || strings.HasPrefix(s, "$2y$")
}

// BeforeCreate 创建用户前创建检查钩子
func (u *User) BeforeCreate(tx *gorm.DB) error { // 在创建用户前设置默认值
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

// CheckPassword 检查密码是否正确
func (u *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}
