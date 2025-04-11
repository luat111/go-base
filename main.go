package main

import (
	"context"
	"fmt"
	"go-base/pkg/app"
	"go-base/pkg/config"
	rpc "go-base/pkg/grpc"
	"go-base/pkg/mq"
	"go-base/pkg/restful"
	"go-base/proto"

	"google.golang.org/grpc"
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

func HelloHandler(client *HelloService) func(c *restful.Context) (any, error) {
	return func(c *restful.Context) (any, error) {
		name := c.Request.Query("name")

		if name == "" {
			c.Logger().Warn("Name came empty")
			name = "World"
		}

		res, err := rpc.CallRPC(c.Logger(), c.Context, "SayHello", func(ctx context.Context) (any, error) {
			return client.SayHello(ctx, &proto.HelloRequest{Name: "ntl"})
		})

		return res, err
	}
}

type UpdatePasswordData struct {
	Password           string `json:"password" validate:"required,min=8,max=32"`
	NewPassword        string `json:"new_password" validate:"required,min=8,max=32"`
	NewPasswordConfirm string `json:"new_password_confirm" validate:"eqfield=NewPassword"`
}

func TestPostHandler(c *restful.Context) (any, error) {
	name := c.Request.PathParam("name")
	page := c.Request.Query("page")
	fmt.Println(name, page)

	if name == "" {
		c.Logger().Warn("Name came empty")
		name = "World"
	}

	return true, nil
}

func test(body []byte, metadata map[string]string) {
	fmt.Println(body, metadata)
}

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
		"test": test,
	})

	app.Run()
}

type HelloService struct {
	Client proto.HelloClient
}

func NewHelloService(con grpc.ClientConnInterface) *HelloService {
	client := proto.NewHelloClient(con)

	return &HelloService{
		Client: client,
	}

}

func (h *HelloService) SayHello(ctx context.Context, req *proto.HelloRequest, opts ...grpc.CallOption) (*proto.HelloResponse, error) {
	result, err := h.Client.SayHello(ctx, req)
	if err != nil {
		return nil, err
	}

	return result, nil
}
