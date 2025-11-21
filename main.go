package main

import (
	"context"
	"fmt"
	"go-base/pkg/app"
	"go-base/pkg/common/types"
	"go-base/pkg/config"
	"go-base/pkg/datasource/postgres/repository"
	"go-base/pkg/encrypt"
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

	//Crypto
	CryptoConfig encrypt.CryptoConfig `mapstructure:"CRYPTO"`
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

	encryptSvc,_ := encrypt.NewService(encrypt.CryptoConfig{
		PublicKeyPEM:  app.Config.Get("PUBLIC_KEY_PEM"),
		PrivateKeyPEM: app.Config.Get("PRIVATE_KEY_PEM"),
	})

	group := app.Group("v1",encryptSvc.PayloadMiddleware(encrypt.MiddlewareConfig{
		
	}))

	app.GET(group, "/test", HelloHandler(helloService))
	app.POST(group, "/test/:name/:test", new(UpdatePasswordData), TestPostHandler)

	app.ListenRMQ(map[string]mq.HandlerFunc{
		"test": test,
	})

	app.DB().MigrateEntities([]any{&TestWorkflow{}})

	wfctrl := NewWfController(app)
	wfctrl.setup()

	app.GET(group, "/test-workflow", func(c *restful.Context) (any, error) {

		wfctrl.Execute(c.Context)

		return true, nil
	})

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
	workflow.Workflow
}

type WFCtrl struct {
	*workflow.WorkflowExecutor[TestWorkflow]
}

const (
	StepA workflow.WorkflowStep = "STEP_A"
	StepB workflow.WorkflowStep = "STEP_B"
	StepC workflow.WorkflowStep = "STEP_C"
)

func NewWfController(app *app.App[AppConfig]) *WFCtrl {
	baseRepo := repository.NewBaseRepository(app.DB().DB, new(TestWorkflow))
	wfRepo := workflow.NewWorkflowRepository[TestWorkflow](baseRepo)

	wfExec := workflow.NewWorkflowExecutor(
		app.Container(),
		workflow.WorkflowProps{
			Name: "test",
			Payload: map[string]any{
				"a": "A",
				"b": "B",
			},
			MaxAttempt: 5,
			Schedule:   "*/5 * * * * *",
		},
		wfRepo,
		wfExec,
		workflow.RetryConfig{
			MaxAttempt: 5,
		},
	)

	return &WFCtrl{WorkflowExecutor: wfExec}
}

func (wctrl *WFCtrl) setup() {
	wctrl.WorkflowExecutor.SetProcessResults(types.JSONB{
		string(StepA): workflow.New,
		string(StepB): workflow.New,
		string(StepC): workflow.New,
	})

	wctrl.WorkflowExecutor.SetStepOperators(map[workflow.WorkflowStep]workflow.StepHandler{
		StepA: wctrl.stepA,
		StepB: wctrl.stepB,
		StepC: wctrl.stepC,
	})

	wctrl.WorkflowExecutor.SetStepResults(map[workflow.WorkflowStep]workflow.StepHandler{
		StepA: func(ctx context.Context, args any) (workflow.WorkflowResult, error) {
			return wctrl.GetStepResult(StepA), nil

		},
		StepB: func(ctx context.Context, args any) (workflow.WorkflowResult, error) {
			return wctrl.GetStepResult(StepB), nil

		},
		StepC: func(ctx context.Context, args any) (workflow.WorkflowResult, error) {
			return wctrl.GetStepResult(StepC), nil

		},
	})
}

func (ctrl *WFCtrl) stepA(ctx context.Context, args any) (workflow.WorkflowResult, error) {
	fmt.Println(args)
	return workflow.Succeed, nil
}

func (ctrl *WFCtrl) stepB(ctx context.Context, args any) (workflow.WorkflowResult, error) {
	fmt.Println(args)
	return workflow.Succeed, nil
}

func (ctrl *WFCtrl) stepC(ctx context.Context, args any) (workflow.WorkflowResult, error) {
	fmt.Println(args)
	return workflow.Succeed, nil
}

// Execute

func wfExec(
	ctx context.Context,
	executor *workflow.Executor[TestWorkflow],
	repo workflow.IWorkflowRepository[TestWorkflow],
) (workflow.WorkflowResult, error) {
	resA, err := executor.Execute(ctx, StepA, "execute step A")
	if err != nil || resA != workflow.Succeed {
		return workflow.Failed, err
	}

	resB, err := executor.Execute(ctx, StepB, "execute step B")
	if err != nil || resB != workflow.Succeed {
		return workflow.Rerun, err
	}

	resC, err := executor.Execute(ctx, StepC, "execute step C")
	if err != nil || resC != workflow.Succeed {
		return workflow.Rerun, err
	}

	repo.Save(ctx, &TestWorkflow{
		Workflow: workflow.Workflow{Status: workflow.Completed},
	})

	return workflow.Completed, nil
}
