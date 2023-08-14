package router

import "github.com/golang-acexy/starter-gin/ginmodule"

type BasicAuthRouter struct {
}

func (a *BasicAuthRouter) RouterInfo() *ginmodule.RouterInfo {
	return &ginmodule.RouterInfo{
		GroupPath:        "auth",
		BasicAuthAccount: map[string]string{"acexy": "acexy"},
	}
}

func (a *BasicAuthRouter) RegisterHandler(ginWrapper *ginmodule.GinWrapper) {
	ginWrapper.GET("invoke", a.invoke())
}

func (a *BasicAuthRouter) invoke() func(request *ginmodule.Request) (*ginmodule.Response, error) {
	return func(request *ginmodule.Request) (*ginmodule.Response, error) {
		return ginmodule.ResponseSuccess(), nil
	}
}
