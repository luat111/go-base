package main

import (
	"go-base/pkg/app"
	"go-base/pkg/config"
	"go-base/pkg/mq"
)

func main() {
	appEnv := config.EnvOptions{
		Path: "/", EnvInterface: AppConfig{},
	}

	app := app.New[AppConfig](appEnv)

	app.ConnectClients(map[string]string{"test": ":3003"})
	helloService := NewHelloService(app.GetClient("test"))

	group := app.Group("v1")

	app.GET(group, "/test", HelloHandler(helloService))
	app.POST(group, "/test/:name/:test", new(UpdatePasswordData), TestPostHandler)

	app.ListenRMQ(map[string]mq.HandlerFunc{
		"test": TestMQ,
	})

	app.Run()
}
