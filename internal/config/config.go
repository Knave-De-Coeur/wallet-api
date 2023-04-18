package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

func fallbackConfigs() {
	viper.SetDefault("DB_CONNECTION_FORMAT", "%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true")
	viper.SetDefault("MYSQL_DB", "wallet")
	viper.SetDefault("MYSQL_USER", "alex")
	viper.SetDefault("MYSQL_PASSWORD", "alexsecret")
	viper.SetDefault("MYSQL_HOST", "localhost")
	viper.SetDefault("PORT", 8080)
	viper.SetDefault("MAX_CONNECTIONS", 100)
	viper.SetDefault("MAX_IDLE_CONNECTIONS", 10)
	viper.SetDefault("MAX_LIFETIME", 1)
	viper.SetDefault("REDIS_ADDRESS", "localhost:6379")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("REDIS_DB", 0)
}

// Configurations app configs from env file, env params or fallback configs
type Configurations struct {
	DBConnectionFormat     string        `mapstructure:"DB_CONNECTION_FORMAT"`
	DBName                 string        `mapstructure:"MYSQL_DB"`
	DBUser                 string        `mapstructure:"MYSQL_USER"`
	DBPassword             string        `mapstructure:"MYSQL_PASSWORD"`
	Host                   string        `mapstructure:"MYSQL_HOST"`
	Port                   string        `mapstructure:"PORT"`
	MaxConnections         int           `mapstructure:"MAX_CONNECTIONS"`
	MaxIdleConnections     int           `mapstructure:"MAX_IDLE_CONNECTIONS"`
	MaxLifetime            int           `mapstructure:"MAX_LIFETIME"`
	RedisAddress           string        `mapstructure:"REDIS_ADDRESS"`
	RedisPassword          string        `mapstructure:"REDIS_PASSWORD"`
	RedisDB                int           `mapstructure:"REDIS_DB"`
	AccessTokenPrivateKey  string        `mapstructure:"ACCESS_TOKEN_PRIVATE_KEY"`
	AccessTokenPublicKey   string        `mapstructure:"ACCESS_TOKEN_PUBLIC_KEY"`
	RefreshTokenPrivateKey string        `mapstructure:"REFRESH_TOKEN_PRIVATE_KEY"`
	RefreshTokenPublicKey  string        `mapstructure:"REFRESH_TOKEN_PUBLIC_KEY"`
	AccessTokenExpiresIn   time.Duration `mapstructure:"ACCESS_TOKEN_EXPIRED_IN"`
	RefreshTokenExpiresIn  time.Duration `mapstructure:"REFRESH_TOKEN_EXPIRED_IN"`
	AccessTokenMaxAge      int           `mapstructure:"ACCESS_TOKEN_MAXAGE"`
	RefreshTokenMaxAge     int           `mapstructure:"REFRESH_TOKEN_MAXAGE"`
}

var WalletConfigs Configurations

// initConfig reads in config file and ENV variables if set.
func init() {

	var err error

	WalletConfigs, err = LoadConfig(".")
	if err != nil {
		log.Fatalf("something went wrong setting up configs: %+v", err)
	}
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (config Configurations, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	if err = viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fallbackConfigs()
		} else {
			return
		}
	}

	err = viper.Unmarshal(&config)

	return
}
