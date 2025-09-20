package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type CustomClaims struct {
	UserID   uint    `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

var jwtSecret = []byte("example-key")

// 生成JWT Token
func GenerateToken(userID uint, username string) (string, error) {
	// 设置Token过期时间
	expirationTime := time.Now().Add(24 * time.Hour)

	// 创建声明
	claims := &CustomClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "ConfessionWall",
		},
	}

	// 创建Token对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 使用密钥签名并获取完整Token
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
