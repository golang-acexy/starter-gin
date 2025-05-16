package router

import (
	"github.com/acexy/golang-toolkit/logger"
	"github.com/golang-acexy/starter-gin/ginstarter"
)

// RestStruct 自定义的Rest结构体
type RestStruct struct {
	Code int
	Msg  string
	Data any
}

type MyRestRouter struct {
}

func (m *MyRestRouter) Info() *ginstarter.RouterInfo {
	return &ginstarter.RouterInfo{
		GroupPath: "my-rest",
	}
}

func (m *MyRestRouter) Handlers(router *ginstarter.RouterWrapper) {
	router.GET("m1", m.m1())
	router.GET("m2", m.m2())
	router.GET("m3", m.m3())
}

// 使用框架自带的Rest响应默认Rest结构体
func (m *MyRestRouter) m1() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		logger.Logrus().Info("invoke m1")
		return ginstarter.RespRestSuccess("data part"), nil
	}
}

// 使用框架自带的Rest响应自定义结构体
func (m *MyRestRouter) m2() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		return ginstarter.NewRespRest().SetDataResponse(&RestStruct{
			Code: 200,
			Msg:  "success",
			Data: "invoke",
		}), nil
	}
}

// 自实现Response响应数据
func (m *MyRestRouter) m3() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		response := &MyRestResponse{}
		response.setData("my rest impl")
		return response, nil
	}
}

type MyRestResponse struct {
}

func (m *MyRestResponse) Data() *ginstarter.ResponseData {
	return ginstarter.NewEmptyResponseData()
}

func (m *MyRestResponse) setData(data any) {

}
