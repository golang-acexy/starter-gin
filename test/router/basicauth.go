package router

import "github.com/golang-acexy/starter-gin/ginmodule"

type BasicAuthRouter struct {
}

func (a *BasicAuthRouter) Info() *ginmodule.RouterInfo {
	return &ginmodule.RouterInfo{
		GroupPath: "auth",
		BasicAuthAccount: &ginmodule.BasicAuthAccount{
			Username: "acexy",
			Password: "acexy",
		},
	}
}

func (a *BasicAuthRouter) Handlers(router *ginmodule.RouterWrapper) {
	router.GET("invoke", a.invoke())
}

func (a *BasicAuthRouter) invoke() ginmodule.HandlerWrapper {
	return func(request *ginmodule.Request) (ginmodule.Response, error) {
		return ginmodule.RespRestSuccess(), nil
	}
}
