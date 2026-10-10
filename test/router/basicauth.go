package router

import "github.com/golang-acexy/starter-gin/ginstarter"

// BasicAuthRouter implements the manual integration example.
type BasicAuthRouter struct {
}

// Info returns the example route-group configuration.
func (a *BasicAuthRouter) Info() *ginstarter.RouterInfo {
	return &ginstarter.RouterInfo{
		GroupPath: "auth",

		// Attach a pre-interceptor to this route group.
		PreInterceptors: []ginstarter.PreInterceptor{
			ginstarter.BasicAuthInterceptor(&ginstarter.BasicAuthAccount{
				Username: "acexy",
				Password: "acexy",
			}),
		},
	}
}

// Handlers registers example handlers during startup.
func (a *BasicAuthRouter) Handlers(router *ginstarter.RouterWrapper) {
	router.GET("invoke", a.invoke())
}

func (a *BasicAuthRouter) invoke() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		return ginstarter.RespTextPlain([]byte("request auth success")), nil
	}
}
