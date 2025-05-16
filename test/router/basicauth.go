package router

import "github.com/golang-acexy/starter-gin/ginstarter"

type BasicAuthRouter struct {
}

func (a *BasicAuthRouter) Info() *ginstarter.RouterInfo {
	return &ginstarter.RouterInfo{
		GroupPath: "auth",

		// 为该路由添加中间件
		PreInterceptors: []ginstarter.PreInterceptor{
			ginstarter.BasicAuthInterceptor(&ginstarter.BasicAuthAccount{
				Username: "acexy",
				Password: "acexy",
			}),
		},
	}
}

func (a *BasicAuthRouter) Handlers(router *ginstarter.RouterWrapper) {
	router.GET("invoke", a.invoke())
}

func (a *BasicAuthRouter) invoke() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		return ginstarter.RespTextPlain([]byte("request auth success")), nil
	}
}
