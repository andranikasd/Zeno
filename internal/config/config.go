package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	AWSAccessKeyID     string `mapstructure:"aws_access_key_id"`
	AWSSecretAccessKey string `mapstructure:"aws_secret_access_key"`
	S3Bucket           string `mapstructure:"s3_bucket"`
	S3Region           string `mapstructure:"s3_region"`
	S3Prefix           string `mapstructure:"s3_prefix"`
	CachePath          string `mapstructure:"cache_path"`
	DuckDBPath         string `mapstructure:"duckdb_path"`
	ServerAddr         string `mapstructure:"server_addr"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile("config.yaml")
	viper.AutomaticEnv()
	viper.SetDefault("server_addr", ":8080")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
