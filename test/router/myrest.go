package router

import (
	"github.com/golang-acexy/starter-gin/ginmodule"
)

// RestStruct 自定义的Rest结构体
type RestStruct struct {
	Code int
	Msg  string
	Data any
}

type MyRestRouter struct {
}

func (m *MyRestRouter) Info() *ginmodule.RouterInfo {
	return &ginmodule.RouterInfo{
		GroupPath: "my-rest",
	}
}

func (m *MyRestRouter) Handlers(router *ginmodule.RouterWrapper) {
	router.GET("invoke", m.invoke())
}

func (m *MyRestRouter) invoke() ginmodule.HandlerWrapper {
	return func(request *ginmodule.Request) (ginmodule.Response, error) {
		return ginmodule.NewRespRest().RestDataResponse(&RestStruct{
			Code: 200,
			Msg:  "success",
			Data: "invoke",
		}), nil
	}
}
