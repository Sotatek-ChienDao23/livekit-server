package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	LivekitConfig LivekitConfig `mapstructure:"livekit"`
}

type LivekitApiConfig struct {
	Key    string `mapstructure:"key"`
	Secret string `mapstructure:"secret"`
}

type LivekitConfig struct {
	Api LivekitApiConfig `mapstructure:"api"`
	Url string           `mapstructure:"url"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return config, fmt.Errorf("error reading config file: %w", err)
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return config, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return config, nil
}
