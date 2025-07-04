package router

import (
	"errors"
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
	router.GET("panic", a.panic())
}

func (a *AbortRouter) invoke() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		return ginstarter.RespHttpStatusCode(203), nil
	}
}

func (a *AbortRouter) panic() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		request.Panic(403, errors.New("panic"))
		return nil, nil
	}
}
