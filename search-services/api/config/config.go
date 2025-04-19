package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type HTTPConfig struct {
	Address string        `yaml:"address" env:"API_ADDRESS" env-default:"localhost:28080"`
	Timeout time.Duration `yaml:"timeout" env:"API_TIMEOUT" env-default:"5s"`
}

type AuthConfig struct {
	AdminUser     string        `yaml:"admin_user" env:"ADMIN_USER" env-default:"admin"`
	AdminPassword string        `yaml:"admin_password" env:"ADMIN_PASSWORD" env-default:"password"`
	TokenTTL      time.Duration `yaml:"token_ttl" env:"TOKEN_TTL" env-default:"2m"`
	SecretKey     string        `yaml:"jwt_secret" env:"JWT_SECRET" env-default:"dKJHSUDNI7b6*E#N(698MFD*#U98398m)"`
}

type Config struct {
	LogLevel          string     `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
	HTTPConfig        HTTPConfig `yaml:"api_server"`
	WordsAddress      string     `yaml:"words_address" env:"WORDS_ADDRESS" env-default:"words:81"`
	UpdateAddress     string     `yaml:"update_address" env:"UPDATE_ADDRESS" env-default:"update:82"`
	SearchAddress     string     `yaml:"search_address" env:"SEARCH_ADDRESS" env-default:"search:83"`
	AuthConfig        AuthConfig `yaml:"auth"`
	SearchRate        float64    `yaml:"search_rate" env:"SEARCH_RATE" env-default:"100"`
	SearchConcurrency int        `yaml:"search_concurrency" env:"SEARCH_CONCURRENCY" env-default:"10"`
}

func MustLoad(configPath string) Config {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			log.Fatalf("cannot read config %q: %s", configPath, err)
		}
	}
	return cfg
}
