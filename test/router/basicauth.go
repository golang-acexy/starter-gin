package router

import "github.com/golang-acexy/starter-gin/ginstarter"

type BasicAuthRouter struct {
}

func (a *BasicAuthRouter) Info() *ginstarter.RouterInfo {
	return &ginstarter.RouterInfo{
		GroupPath: "auth",
		BasicAuthAccount: &ginstarter.BasicAuthAccount{
			Username: "acexy",
			Password: "acexy",
		},
	}
}

func (a *BasicAuthRouter) Handlers(router *ginstarter.RouterWrapper) {
	router.GET("invoke", a.invoke())
}

func (a *BasicAuthRouter) invoke() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		return ginstarter.RespTextPlain("request auth success"), nil
	}
}
