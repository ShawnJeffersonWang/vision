package jwt

import (
	"agricultural_vision/settings"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var mySecret []byte

// MyClaims 自定义声明类型
type MyClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// Init 初始化JWT模块
func Init() error {
	if settings.Conf.JWTConfig == nil {
		return errors.New("jwt config not found")
	}

	if settings.Conf.JWTConfig.Secret == "" {
		return errors.New("jwt secret not found in config")
	}

	if len(settings.Conf.JWTConfig.Secret) < 32 {
		return errors.New("jwt secret is too short, recommend at least 32 characters")
	}

	mySecret = []byte(settings.Conf.JWTConfig.Secret)
	return nil
}

// GenToken 生成JWT（通用方法）
func GenToken(userID int64, username string, tokenType string, customExpireHours ...int) (string, error) {
	if len(mySecret) == 0 {
		return "", errors.New("jwt secret not initialized")
	}

	// 确定过期时间
	var expireHours int
	if len(customExpireHours) > 0 && customExpireHours[0] > 0 {
		expireHours = customExpireHours[0]
	} else {
		// 根据token类型使用不同的默认过期时间
		if tokenType == "refresh_token" {
			expireHours = settings.Conf.JWTConfig.RefreshExpireHours
			if expireHours <= 0 {
				expireHours = 720 // 默认30天
			}
		} else {
			expireHours = settings.Conf.JWTConfig.ExpireHours
			if expireHours <= 0 {
				expireHours = 2 // 默认2小时
			}
		}
	}

	expireDuration := time.Hour * time.Duration(expireHours)

	claims := MyClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    settings.Conf.JWTConfig.Issuer,
			Subject:   tokenType,
			ID:        generateJTI(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(mySecret)
}

// GenAccessToken 生成访问token
func GenAccessToken(userID int64, username string) (string, error) {
	return GenToken(userID, username, "access_token")
}

// GenRefreshToken 生成刷新token
func GenRefreshToken(userID int64, username string) (string, error) {
	return GenToken(userID, username, "refresh_token")
}

// 在 auth/jwt.go 中添加
func ParseRefreshToken(tokenString string) (*MyClaims, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	// 验证是否是刷新token
	if claims.Subject != "refresh_token" {
		return nil, errors.New("not a refresh token")
	}

	return claims, nil
}

// 解析JWT
func ParseToken(tokenString string) (*MyClaims, error) {
	if len(mySecret) == 0 {
		return nil, errors.New("jwt secret not initialized")
	}

	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return mySecret, nil
	})

	if err != nil {
		return nil, err
	}

	// 类型断言获取claims
	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// 生成JWT ID
func generateJTI() string {
	return time.Now().Format("20060102150405.000")
}
