package main

import (
	"fmt"
	"go-base/pkg/app"
	"go-base/pkg/config"
	"go-base/pkg/restful"
	"go-base/proto"
	"go-base/rpc-test/test"
)

type RedisOptions struct {
	CacheHost string `mapstructure:"CACHE_HOST"`
	CachePort string `mapstructure:"CACHE_PORT"`
	CachePass string `mapstructure:"CACHE_PWD"`
	CacheDB   int    `mapstructure:"CACHE_DB"`
}

type AppConfig struct {
	//Environment
	AppPort string `mapstructure:"PORT"`
	RpcPort string `mapstructure:"RPC_PORT"`
	AppName string `mapstructure:"APP_NAME"`
	ENV     string `mapstructure:"ENV" json:"ENV"`

	//Base Route
	API_PATH string `mapstructure:"API_PATH"`

	//Redis
	CacheOptions RedisOptions `mapstructure:"CACHE"`
}

func HelloHandler(c *restful.Context) (any, error) {
	name := c.Request.Query("name")

	if name == "" {
		c.Logger().Info("Name came empty")
		name = "World"
	}

	return fmt.Sprintf("Hello %s!", name), nil
}

func main() {
	appEnv := config.EnvOptions{
		Path: "/rpc-test", EnvInterface: AppConfig{},
	}

	app := app.New[AppConfig](appEnv)
	group := app.Group("v1")
	app.GET(group, "/test", HelloHandler)

	proto.RegisterHelloServer(app, test.NewHelloService())

	app.Run()
}
