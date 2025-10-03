package jwt

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// CustomSecret 存储用于 HMAC 签名验证的密钥字节切片
// 通过环境变量 JWT_SECRET 初始化（startup 时）。如果未设置，将在 init() 中记录错误。
var CustomSecret []byte

func init() {
	// 若使用生成的密钥，则应用启动时初始化密钥
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET环境不存在！")
	}
	CustomSecret = []byte(secret)
}

// CustomClaims 为自定义的 JWT claims，携带 user 相关信息
type CustomClaims struct {
	//我在这里加了自己申明的字段，这样你才能评鉴出这是我写的史
	UserID               uint   `json:"user_id"`
	Username             string `json:"username"`
	jwt.RegisteredClaims        // 内嵌标准的声明
}

// GenerateToken 使用当前全局 CustomSecret 对 claims 进行签名并返回 token 字符串
// 如果 CustomSecret 为空，会返回错误，避免生成无效/不安全的 token
func GenerateToken(UserID uint, username string) (string, error) {
	if len(CustomSecret) == 0 {
		return "", errors.New("jwt secret is not configured")
	}
	claims := CustomClaims{
		UserID:   UserID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "ConfessionWall",
		},
	}
	// 使用SHA256算法创建签名对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	//返回完整的token
	return token.SignedString(CustomSecret)
}

// 解析JWT Token
func ParseToken(tokenString string) (*CustomClaims, error) {
	if len(CustomSecret) == 0 {
		return nil, errors.New("jwt secret is not configured")
	}
	var claims = new(CustomClaims) // 由于自定义了Claim结构体，需要使用 ParseWithClaims 方法进行解析
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return CustomSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if token == nil || !token.Valid {
		// token无效或已过期和非法的token全部报告无效，防止继续执行访问本就为空的claims造成空指针panic
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
