package hermes

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/apourchet/hermes/binding"
	"github.com/apourchet/hermes/endpoint"
	"github.com/gin-gonic/gin"
)

const (
	HERMES_CODE_BYPASS = -1
)

func findEndpointByHandler(svc IServer, name string) (*endpoint.Endpoint, error) {
	for _, ep := range svc.Endpoints() {
		if ep.Handler == name {
			return ep, nil
		}
	}
	return nil, fmt.Errorf("MethodNotFoundError")
}

func getGinHandler(svc IServer, binder binding.Binding, ep *endpoint.Endpoint, method reflect.Method) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input reflect.Value
		if ep.InputType != nil {
			input = reflect.New(ep.InputType)
		}

		var output reflect.Value
		if ep.OutputType != nil {
			output = reflect.New(ep.OutputType)
		}

		// Bind input to context
		if input.IsValid() {
			err := binder.Bind(ctx, input.Interface())
			if err != nil {
				ctx.JSON(http.StatusBadRequest, &gin.H{"message": err.Error()})
				return
			}
		}

		// Prepare arguments to function
		args := []reflect.Value{reflect.ValueOf(svc), reflect.ValueOf(ctx)}
		if input.IsValid() {
			args = append(args, input)
		}
		if output.IsValid() {
			args = append(args, output)
		}

		// Call function
		vals := method.Func.Call(args)
		code := int(vals[0].Int())
		if code == HERMES_CODE_BYPASS {
			// Bypass code, do nothing here
			return
		}

		if !vals[1].IsNil() { // If there was an error
			errVal := vals[1].Interface().(error)
			ctx.JSON(code, map[string]string{"message": errVal.Error()})
		} else if output.IsValid() {
			ctx.JSON(code, output.Interface())
		}
	}
}
