package jwt

import (
	"errors"
	"log"
	"time"

	//"os"
	"github.com/golang-jwt/jwt/v5"
)

var CustomSecret = []byte("114514")

//var CustomSecret []byte
// CustomSecret 用于加盐的字符串,暂时没有想好用时间当字符串还是在服务器内部使用openssl生成一个密钥并传入环境变量JWT_SECRET中
// 在搞好逻辑关系之前，暂时使用不安全的文本当sercert
/*func init() {
    // 若使用生成的密钥，则应用启动时初始化密钥
    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        log.Fatal("JWT_SECRET环境不存在！")
    }
    CustomSecret = []byte(secret)
}*/

type CustomClaims struct {
	//我在这里加了自己申明的字段，这样你才能评鉴出这是我写的史
	UserID               int    `json:"user_id"`
	Username             string `json:"username"`
	jwt.RegisteredClaims        // 内嵌标准的声明
}

// 生成JWT
func GenerateToken(UserID int, username string) (string, error) {
	claims := CustomClaims{
		UserID,
		username,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), //设置token过期时间为1天
			Issuer:    "Server",
		},
	}
	// 使用SHA256算法创建签名对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	//返回完整的token
	return token.SignedString(CustomSecret)
}

// 解析JWT Token
func ParseToken(tokenString string) (*CustomClaims, error) {
	var claims = new(CustomClaims)
	// 由于自定义了Claim结构体，需要使用 ParseWithClaims 方法进行解析
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (i interface{}, err error) {
		return CustomSecret, nil // 返回用于验证的密钥
	})
	// 验证token有效性（包括过期时间等）
	if token.Valid { // 校验token
		return claims, nil
	}
	if err != nil {
		return nil, err
	}
	log.Printf("该token无效！请检查")
	return nil, errors.New("invalid token")
}
