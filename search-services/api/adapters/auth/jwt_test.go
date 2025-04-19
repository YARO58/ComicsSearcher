package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	"yadro.com/course/api/config"
	"yadro.com/course/api/core"
)

func TestJWTAuth_GenerateToken(t *testing.T) {
	tests := []struct {
		name     string
		config   config.AuthConfig
		username string
		password string
		wantErr  error
	}{
		{
			name: "valid credentials",
			config: config.AuthConfig{
				AdminUser:     "admin",
				AdminPassword: "password",
				SecretKey:     "secret",
				TokenTTL:      time.Hour,
			},
			username: "admin",
			password: "password",
			wantErr:  nil,
		},
		{
			name: "invalid username",
			config: config.AuthConfig{
				AdminUser:     "admin",
				AdminPassword: "password",
				SecretKey:     "secret",
				TokenTTL:      time.Hour,
			},
			username: "wrong",
			password: "password",
			wantErr:  core.ErrInvalidCredentials,
		},
		{
			name: "invalid password",
			config: config.AuthConfig{
				AdminUser:     "admin",
				AdminPassword: "password",
				SecretKey:     "secret",
				TokenTTL:      time.Hour,
			},
			username: "admin",
			password: "wrong",
			wantErr:  core.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := NewJWTAuth(tt.config)
			token, err := auth.GenerateToken(tt.username, tt.password)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Empty(t, token)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, token)
		})
	}
}

func TestJWTAuth_ValidateToken(t *testing.T) {
	validConfig := config.AuthConfig{
		AdminUser:     "admin",
		AdminPassword: "password",
		SecretKey:     "secret",
		TokenTTL:      time.Hour,
	}

	tests := []struct {
		name      string
		config    config.AuthConfig
		token     string
		shouldErr bool
		setupFunc func() string
	}{
		{
			name:   "valid token",
			config: validConfig,
			setupFunc: func() string {
				auth := NewJWTAuth(validConfig)
				token, _ := auth.GenerateToken("admin", "password")
				return token
			},
			shouldErr: false,
		},
		{
			name:      "invalid token format",
			config:    validConfig,
			token:     "invalid.token.format",
			shouldErr: true,
		},
		{
			name:   "expired token",
			config: validConfig,
			setupFunc: func() string {
				claims := jwt.MapClaims{
					"sub": "superuser",
					"exp": time.Now().Add(-time.Hour).Unix(),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte(validConfig.SecretKey))
				return tokenString
			},
			shouldErr: true,
		},
		{
			name:   "wrong signing method",
			config: validConfig,
			setupFunc: func() string {
				claims := jwt.MapClaims{
					"sub": "superuser",
					"exp": time.Now().Add(time.Hour).Unix(),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
				tokenString, _ := token.SignedString([]byte(validConfig.SecretKey))
				return tokenString
			},
			shouldErr: true,
		},
		{
			name:   "wrong secret key",
			config: validConfig,
			setupFunc: func() string {
				claims := jwt.MapClaims{
					"sub": "superuser",
					"exp": time.Now().Add(time.Hour).Unix(),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte("wrong_secret"))
				return tokenString
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := NewJWTAuth(tt.config)
			token := tt.token
			if tt.setupFunc != nil {
				token = tt.setupFunc()
			}

			err := auth.ValidateToken(token)

			if tt.shouldErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}
