package router

import (
	"github.com/golang-acexy/starter-gin/ginmodule"
)

type AbortRouter struct {
}

func (a *AbortRouter) Info() *ginmodule.RouterInfo {
	return &ginmodule.RouterInfo{
		GroupPath: "abort",
	}
}

func (a *AbortRouter) Handlers(router *ginmodule.RouterWrapper) {
	router.GET("invoke", a.invoke())
}

func (a *AbortRouter) invoke() ginmodule.HandlerWrapper {
	return func(request *ginmodule.Request) (ginmodule.Response, error) {
		return ginmodule.RespAbortWithStatus(203), nil
	}
}
