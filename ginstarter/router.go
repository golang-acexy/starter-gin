package ginstarter

import (
	"github.com/acexy/golang-toolkit/util/coll"
	"github.com/gin-gonic/gin"
)

type RouterInfo struct {
	// GroupPath 路由分组路径
	GroupPath string

	// 该Router下的前置拦截器
	PreInterceptors []PreInterceptor
	// 该Router下的后置拦截器
	PostInterceptors []PostInterceptor
}

type Router interface {
	// Info 定义路由信息
	Info() *RouterInfo
	// Handlers 注册处理器
	Handlers(router *RouterWrapper)
}

func registerRouter(ginEngine *gin.Engine, routers []Router) {
	for _, router := range routers {
		routerInfo := router.Info()
		group := ginEngine.Group(routerInfo.GroupPath)

		routerInfo.PreInterceptors = coll.SliceFilter(routerInfo.PreInterceptors, func(p PreInterceptor) bool {
			return p != nil
		})
		routerInfo.PostInterceptors = coll.SliceFilter(routerInfo.PostInterceptors, func(p PostInterceptor) bool {
			return p != nil
		})

		if len(routerInfo.PreInterceptors) != 0 || len(routerInfo.PostInterceptors) != 0 {
			if len(routerInfo.PreInterceptors) > 0 {
				group.Use(func(ctx *gin.Context) {
					// 有group级别的前置拦截器
					for i := range routerInfo.PreInterceptors {
						currentHandler, ok := ctx.Get(ginCtxKeyContinueHandler)
						response, continuePreInterceptor, continueHandler := routerInfo.PreInterceptors[i](&Request{ctx: ctx})
						if !(ok && !currentHandler.(bool)) {
							ctx.Set(ginCtxKeyContinueHandler, continueHandler)
						}
						if response != nil {
							httpResponse(ctx, response)
						}
						if continuePreInterceptor {
							continue
						} else {
							break
						}
					}
					ctx.Next()
				})
			}
			group.Use(func(ctx *gin.Context) {
				v, exists := ctx.Get(ginCtxKeyContinueHandler)
				if !exists || v.(bool) {
					ctx.Next()
				}
				if len(routerInfo.PostInterceptors) > 0 {
					var response Response
					var newResponse Response
					var continuePostInterceptor bool
					currentResponse, exists := ctx.Get(ginCtxKeyCurrentResponse)
					if exists && currentResponse != nil {
						response = currentResponse.(Response)
					}
					for i := range routerInfo.PostInterceptors {
						interceptor := routerInfo.PostInterceptors[i]
						newResponse, continuePostInterceptor = interceptor(&Request{ctx: ctx}, response)
						if newResponse != nil {
							response = newResponse
						}
						if continuePostInterceptor {
							continue
						}
						break
					}
					if response != nil {
						httpResponse(ctx, response)
					}
				}
			})
		}
		router.Handlers(&RouterWrapper{routerGroup: group})
	}
}
