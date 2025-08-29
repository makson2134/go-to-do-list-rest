package config

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Server ServerCfg `env-prefix:"SERVER_"`
	DB     DBCfg     `env-prefix:"DB_"`
	JWT    JWTCfg    `env-prefix:"JWT_"`
}

type ServerCfg struct {
	Port        string        `env:"PORT" env-required:"true"`
	Timeout     time.Duration `env:"TIMEOUT" env-required:"true"`
	IdleTimeout time.Duration `env:"IDLETIMEOUT" env-required:"true"`
}

type DBCfg struct {
	User            string        `env:"USER" env-required:"true"`
	Password        string        `env:"PASSWORD" env-required:"true"`
	Name            string        `env:"NAME" env-required:"true"`
	Host            string        `env:"HOST" env-required:"true"`
	Port            int           `env:"PORT" env-required:"true"`
	SSLMode         string        `env:"SSLMODE" env-required:"true"`
	MaxIdleConns    int           `env:"MAXIDLECONNS" env-required:"true"`
	MaxOpenConns    int           `env:"MAXOPENCONNS" env-required:"true"`
	ConnMaxLifetime time.Duration `env:"CONNMAXLIFETIME" env-required:"true"`
}

type JWTCfg struct {
	SecretKey string        `env:"SECRET_KEY" env-required:"true"`
	TTL       time.Duration `env:"TTL" env-required:"true"`
}

func NewConfig() (*Config, error) {
	err := godotenv.Load(".env")
	if err != nil {

	}
	var cfg Config
	err = cleanenv.ReadEnv(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
