package router

import (
	"github.com/acexy/golang-toolkit/logger"
	"github.com/golang-acexy/starter-gin/ginstarter"
)

// RestStruct is the custom envelope used by the response example.
type RestStruct struct {
	// Code is the custom result code.
	Code int
	// Msg describes the custom result.
	Msg string
	// Data carries the example payload.
	Data any
}

// MyRestRouter implements the manual integration example.
type MyRestRouter struct {
}

// Info returns the example route-group configuration.
func (m *MyRestRouter) Info() *ginstarter.RouterInfo {
	return &ginstarter.RouterInfo{
		GroupPath: "my-rest",
	}
}

// Handlers registers example handlers during startup.
func (m *MyRestRouter) Handlers(router *ginstarter.RouterWrapper) {
	router.GET("m1", m.m1())
	router.GET("m2", m.m2())
	router.GET("m3", m.m3())
}

// m1 demonstrates the default REST envelope.
func (m *MyRestRouter) m1() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		logger.Logrus().Info("invoke m1")
		return ginstarter.RespRestSuccess("data part"), nil
	}
}

// m2 demonstrates a custom structured REST body.
func (m *MyRestRouter) m2() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		return ginstarter.NewRespRest().SetBodyResponse(&RestStruct{
			Code: 200,
			Msg:  "success",
			Data: "invoke",
		}), nil
	}
}

// m3 demonstrates a custom Response implementation.
func (m *MyRestRouter) m3() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		response := &MyRestResponse{}
		response.setData("my rest impl")
		return response, nil
	}
}

// MyRestResponse implements the manual integration example.
type MyRestResponse struct {
}

// Data returns the example response entity.
func (m *MyRestResponse) Data() *ginstarter.ResponseEntity {
	return ginstarter.NewEmptyResponseEntity()
}

func (m *MyRestResponse) setData(data any) {

}
