package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	DBHost                 string `mapstructure:"DB_HOST"`
	DBPort                 string `mapstructure:"DB_PORT"`
	DBUser                 string `mapstructure:"DB_USER"`
	DBPassword             string `mapstructure:"DB_PASSWORD"`
	DBName                 string `mapstructure:"DB_NAME"`
	Port                   string `mapstructure:"PORT"`
	JWTSecret              string `mapstructure:"JWT_SECRET"`
	RedisURI               string `mapstructure:"REDIS_URI"`
	AuthServiceGRPCAddress string `mapstructure:"AUTH_SERVICE_GRPC_ADDR"`
}

func LoadConfig() (config *Config, err error) {
	viper.SetConfigFile("./.env")
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		log.Printf("Warning: .env file not found, using system env")
	}

	var c Config
	err = viper.Unmarshal(&c)
	config = &c
	return
}
