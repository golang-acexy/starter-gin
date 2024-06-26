package router

import (
	"github.com/golang-acexy/starter-gin/ginstarter"
)

type AbortRouter struct {
}

func (a *AbortRouter) Info() *ginstarter.RouterInfo {
	return &ginstarter.RouterInfo{
		GroupPath: "abort",
	}
}

func (a *AbortRouter) Handlers(router *ginstarter.RouterWrapper) {
	router.GET("invoke", a.invoke())
}

func (a *AbortRouter) invoke() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		return ginstarter.RespAbortWithHttpStatusCode(203), nil
	}
}
