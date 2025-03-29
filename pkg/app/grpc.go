package app

import (
	"errors"
	"go-base/pkg/container"
	"reflect"

	"google.golang.org/grpc"
)

var (
	errNonAddressable = errors.New("cannot inject container as it is not addressable or is fail")
)

func (a *App[EnvInterface]) ConnectClients(listAddr map[string]string) error {
	var err error

	for svcName, value := range listAddr {
		err = errors.Join(err, a.grpcServer.RegisterClient(svcName, value))
	}

	return err
}

func (a *App[EnvInterface]) GetClient(name string) *grpc.ClientConn {
	return a.grpcServer.Services[name]
}

// RegisterService adds a gRPC service to the GoFr application.
func (a *App[EnvInterface]) RegisterService(desc *grpc.ServiceDesc, impl any) {
	a.container.Logger.Info("Registering GRPC server:", "name", desc.ServiceName)
	a.grpcServer.Server.RegisterService(desc, impl)

	err := injectContainer(impl, a.container)
	if err != nil {
		return
	}

	a.grpcRegistered = true
}

func injectContainer(impl any, c *container.Container) error {
	val := reflect.ValueOf(impl)

	// Note: returning nil for the cases where user does not want to inject the container altogether and
	// not to break any existing implementation for the users that are using gRPC server. If users are
	// expecting the container to be injected and are passing non-addressable server struct, we have the
	// DEBUG log for the same.
	if val.Kind() != reflect.Pointer {
		c.Logger.Error("Cannot inject container into non-addressable implementation, consider using pointer",
			val.Type().Name())

		return nil
	}

	val = val.Elem()
	tVal := val.Type()

	for i := 0; i < val.NumField(); i++ {
		f := tVal.Field(i)
		v := val.Field(i)

		if f.Type == reflect.TypeOf(c) {
			if !v.CanSet() {
				c.Logger.Error(errNonAddressable)
				return errNonAddressable
			}

			v.Set(reflect.ValueOf(c))

			// early return expecting only one container field necessary for one gRPC implementation
			return nil
		}

		if f.Type == reflect.TypeOf(*c) {
			if !v.CanSet() {
				c.Logger.Error(errNonAddressable)
				return errNonAddressable
			}

			v.Set(reflect.ValueOf(*c))

			// early return expecting only one container field necessary for one gRPC implementation
			return nil
		}
	}

	return nil
}
