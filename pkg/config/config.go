package config

import (
	"cmp"

	"github.com/gin-contrib/cors"
	"github.com/spf13/viper"
)

type Config interface {
	Get(string) string
	GetCorsConfig() cors.Config
	GetOrDefault(string, string) string
}

type Environment[EnvInterface any] struct {
	Env        *EnvInterface
	CorsConfig cors.Config
}

type EnvOptions struct {
	Path         string
	EnvInterface any
}

func NewAppConfig[EnvInterface any](opt EnvOptions) Config {
	var env Environment[EnvInterface]

	appCnf, err := LoadConfig[EnvInterface](opt.Path)
	corsConfig := GetCorsConfig()

	if err != nil {
		panic(err)
	}

	env.Env = &appCnf
	env.CorsConfig = corsConfig

	return &env
}

func (c *Environment[EnvInterface]) Get(key string) string {
	value, ok := viper.Get(key).(string)

	if !ok {
		return ""
	}

	return value
}

func (c *Environment[EnvInterface]) GetOrDefault(key string, defaultValue string) string {
	value, ok := viper.Get(key).(string)

	if !ok {
		return defaultValue
	}

	return cmp.Or(value, defaultValue)
}

func (c *Environment[EnvInterface]) GetCorsConfig() cors.Config {
	return c.CorsConfig
}
