package config

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"
)

var (
	EnvDev  = "dev"
	EnvTest = "test"
	EnvProd = "prod"
)

type DBConfig struct {
	Host         string
	User         string
	Password     string
	Name         string
	Port         string
	MaxOpenConns int
	MaxIdleConns int
}

func (dbc *DBConfig) GetDBConnString() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbc.Host, dbc.User, dbc.Password, dbc.Name, dbc.Port,
	)
}

type ServerConfig struct {
	Domain string
	Port   string
}

type TokenConfig struct {
	Secret               string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

type Config struct {
	AccessKey     string
	SecretKey     string
	BucketName    string
	URL           string
	SigningRegion string
	Environment   string

	DB     DBConfig
	Server ServerConfig
	Token  TokenConfig
}

func LoadConfig() *Config {
	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Config reading failed: %v", err)
	}

	// Дефолтные значения для конфига базы данных
	viper.SetDefault("DB_PORT", 5432)
	viper.SetDefault("DB_MAX_OPEN_CONNS", 25)
	viper.SetDefault("DB_MAX_IDLE_CONNS", 10)
	viper.SetDefault("SERVER_PORT", 8080)

	config := &Config{
		AccessKey:     viper.GetString("ACCESS_KEY"),
		SecretKey:     viper.GetString("SECRET_KEY"),
		BucketName:    viper.GetString("BUCKET_NAME"),
		URL:           viper.GetString("URL"),
		SigningRegion: viper.GetString("SIGNING_REGION"),
		Environment:   viper.GetString("ENVIRONMENT"),
		DB: DBConfig{
			Host:         viper.GetString("DB_HOST"),
			User:         viper.GetString("DB_USER"),
			Password:     viper.GetString("DB_PASSWORD"),
			Name:         viper.GetString("DB_NAME"),
			Port:         viper.GetString("DB_PORT"),
			MaxOpenConns: viper.GetInt("DB_MAX_OPEN_CONNS"),
			MaxIdleConns: viper.GetInt("DB_MAX_IDLE_CONNS"),
		},
		Server: ServerConfig{
			Domain: viper.GetString("SERVER_DOMAIN"),
			Port:   viper.GetString("SERVER_PORT"),
		},
		Token: TokenConfig{
			Secret:               viper.GetString("TOKEN_SECRET"),
			AccessTokenDuration:  viper.GetDuration("TOKEN_ACCESS_DURATION"),
			RefreshTokenDuration: viper.GetDuration("TOKEN_REFRESH_DURATION"),
		},
	}

	return config
}
