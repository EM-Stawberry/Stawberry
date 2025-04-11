package config

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"
)

type TokenConfig struct {
	Secret               string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

type Config struct {
	DBHost        string
	DBUser        string
	DBPassword    string
	DBName        string
	DBPort        string
	ServerPort    string
	AccessKey     string
	SecretKey     string
	BucketName    string
	URL           string
	SigningRegion string
	Token         TokenConfig
}

func LoadConfig() *Config {
	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Config reading failed: %v", err)
	}

	config := &Config{
		DBHost:        viper.GetString("DB_HOST"),
		DBUser:        viper.GetString("DB_USER"),
		DBPassword:    viper.GetString("DB_PASSWORD"),
		DBName:        viper.GetString("DB_NAME"),
		DBPort:        viper.GetString("DB_PORT"),
		ServerPort:    viper.GetString("SERVER_PORT"),
		AccessKey:     viper.GetString("ACCESS_KEY"),
		SecretKey:     viper.GetString("SECRET_KEY"),
		BucketName:    viper.GetString("BUCKET_NAME"),
		URL:           viper.GetString("URL"),
		SigningRegion: viper.GetString("SIGNING_REGION"),
		Token: TokenConfig{
			Secret:               viper.GetString("TOKEN_SECRET"),
			AccessTokenDuration:  viper.GetDuration("TOKEN_ACCESS_DURATION"),
			RefreshTokenDuration: viper.GetDuration("TOKEN_REFRESH_DURATION"),
		},
	}

	return config
}

func (c *Config) GetDBConnString() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		c.DBHost, c.DBUser, c.DBPassword, c.DBName, c.DBPort,
	)
}
