package main

import (
	"context"
	"fmt"
	"go-base/pkg/app"
	"go-base/pkg/common/types"
	"go-base/pkg/config"
	"go-base/pkg/datasource/postgres/repository"
	rpc "go-base/pkg/grpc"
	"go-base/pkg/mq"
	"go-base/pkg/restful"
	"go-base/pkg/workflow"
	"go-base/proto"

	"github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
)

// Config

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

// HTTP

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

func test(body []byte, metadata map[string]string, msg amqp091.Delivery) {
	fmt.Println(body, metadata)
}

// Main

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

	app.DB().MigrateEntities([]any{&TestWorkflow{}})

	app.Run()
}

// GRPC

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

// Workflow

type TestWorkflow struct {
	*workflow.Workflow
}

type WFCtrl struct {
	*workflow.WorkflowExecutor
}

const (
	StepA workflow.WorkflowStep = "STEP_A"
	StepB workflow.WorkflowStep = "STEP_B"
	StepC workflow.WorkflowStep = "STEP_C"
)

const (
	processA string = "isStepADone"
	processB string = "isStepBDone"
	processC string = "isStepCDone"
)

func NewWfController(app *app.App[AppConfig]) *WFCtrl {
	baseRepo := repository.NewBaseRepository(app.DB().DB, new(TestWorkflow))
	repo := workflow.NewWorkflowRepository(baseRepo)

	wfExec := workflow.NewWorkflowExecutor(
		app.Container(),
		workflow.WorkflowProps{
			Name: "test",
			Payload: map[string]any{
				"a": "A",
				"b": "B",
			},
			MaxAttempt: 5,
			Schedule:   "* * * * *",
		},
		repo,
		wfExec,
		workflow.RetryConfig{
			MaxAttempt: 5,
		},
	)

	wfExec.SetProcessResults(types.JSONB{
		"isStepADone": false,
		"isStepBDone": false,
		"isStepCDone": false,
	})

	wfExec.SetStepOperators(map[workflow.WorkflowStep]workflow.StepHandler{
		StepA: stepA,
		StepB: stepB,
		StepC: stepC,
	})

	wfExec.SetStepResults(map[workflow.WorkflowStep]workflow.StepHandler{
		StepA: func(args any) (workflow.WorkflowResult, error) {

		},
		StepB: stepB,
		StepC: stepC,
	})

	return &WFCtrl{WorkflowExecutor: wfExec}
}

func wfExec(executor *workflow.Executor) (workflow.WorkflowResult, error) {
	return workflow.Failed, nil
}

func stepA(args any) (workflow.WorkflowResult, error) {
	return workflow.Completed, nil
}

func stepB(args any) (workflow.WorkflowResult, error) {
	return workflow.Completed, nil
}

func stepC(args any) (workflow.WorkflowResult, error) {
	return workflow.Completed, nil
}
