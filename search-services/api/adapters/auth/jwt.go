package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"yadro.com/course/api/config"
	"yadro.com/course/api/core"
)

type JWTAuth struct {
	config config.AuthConfig
}

func NewJWTAuth(config config.AuthConfig) *JWTAuth {
	return &JWTAuth{
		config: config,
	}
}

func (a *JWTAuth) GenerateToken(username, password string) (string, error) {
	if username != a.config.AdminUser || password != a.config.AdminPassword {
		return "", core.ErrInvalidCredentials
	}

	claims := jwt.MapClaims{
		"sub": "superuser",
		"exp": time.Now().Add(a.config.TokenTTL).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.config.SecretKey))
}

func (a *JWTAuth) ValidateToken(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, core.ErrInvalidToken
		}
		return []byte(a.config.SecretKey), nil
	})

	if err != nil {
		return err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid || claims["sub"] != "superuser" {
		return core.ErrInvalidToken
	}
	return nil
}
