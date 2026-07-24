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

func registerRouter(ginEngine *gin.Engine, routers []Router) error {
	for _, router := range routers {
		info := router.Info()
		if info == nil {
			return ErrRouterInfoNil
		}
		routerInfo := info
		group := ginEngine.Group(routerInfo.GroupPath)

		routerInfo.PreInterceptors = coll.SliceFilter(routerInfo.PreInterceptors, func(p PreInterceptor) bool {
			return p != nil
		})
		routerInfo.PostInterceptors = coll.SliceFilter(routerInfo.PostInterceptors, func(p PostInterceptor) bool {
			return p != nil
		})

		if len(routerInfo.PreInterceptors) != 0 || len(routerInfo.PostInterceptors) != 0 {
			group.Use(func(ctx *gin.Context) {
				defer recoverResponse(ctx)
				state := getRequestState(ctx)
				runPreInterceptors(ctx, routerInfo.PreInterceptors)
				if !state.stopped {
					nextWithRecovery(ctx)
				}
				normalizeResponse(ctx)
				runPostInterceptors(ctx, routerInfo.PostInterceptors)
			})
		}
		router.Handlers(&RouterWrapper{routerGroup: group})
	}
	return nil
}
