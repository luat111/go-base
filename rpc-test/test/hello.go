package test

import (
	"context"
	"go-base/proto"
	"time"
)

type HelloService struct {
	proto.UnimplementedHelloServer
}

func (s *HelloService) SayHello(ctx context.Context, payload *proto.HelloRequest) (*proto.HelloResponse, error) {
	time.Sleep(5 * time.Second)
	return &proto.HelloResponse{Message: "hello"}, nil
}

func NewHelloService() proto.HelloServer {
	return &HelloService{}
}
