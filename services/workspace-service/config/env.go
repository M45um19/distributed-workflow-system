package config

import (
	"log"
	"reflect"
	"strings"

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
	GoENV                  string `mapstructure:"GO_ENV"`
	KafkaBrokers           string `mapstructure:"KAFKA_BROKERS"`
	TemporalHost           string `mapstructure:"TEMPORAL_HOST"`
	SmtpHost               string `mapstructure:"SMTP_HOST"`
	SmtpPort               string `mapstructure:"SMTP_PORT"`
	SmtpFrom               string `mapstructure:"FROM"`
	SmtpPassword           string `mapstructure:"PASSWORD"`
}

func LoadConfig() (config *Config, err error) {
	viper.SetConfigFile("./.env")
	viper.AutomaticEnv()

	bindAllEnv(Config{})

	err = viper.ReadInConfig()
	if err != nil {
		log.Printf("Warning: .env file not found, using system env")
	}

	var c Config
	err = viper.Unmarshal(&c)
	config = &c
	return
}

func bindAllEnv(s interface{}) {
	t := reflect.TypeOf(s)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("mapstructure")
		if tag != "" {
			envKey := strings.Split(tag, ",")[0]
			viper.BindEnv(envKey)
		}
	}
}
