package main

import (
	"context"
	rpc "go-base/pkg/grpc"
	"go-base/pkg/restful"
	"go-base/proto"

	"google.golang.org/grpc"
)

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
