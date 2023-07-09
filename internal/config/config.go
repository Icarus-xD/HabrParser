package config

import "github.com/spf13/viper"

type Config struct {
	PostgresDriver   string `mapstructure:"POSTGRES_DRIVER"`
	PostgresUser     string `mapstructure:"POSTGRES_USER"`
	PostgresPassword string `mapstructure:"POSTGRES_PASSWORD"`
	PostgresHost     string `mapstructure:"POSTGRES_HOST"`
	PostgresPort     string `mapstructure:"POSTGRES_PORT"`
	PostgresDB       string `mapstructure:"POSTGRES_DB"`
	PostgresSSLMode  string `mapstructure:"POSTGRES_SSLMODE"`
}

func GetConfig() (*Config, error) {
	cfg := new(Config)

	viper.SetConfigFile(".env")

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}