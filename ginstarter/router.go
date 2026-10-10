package ginstarter

import (
	"github.com/acexy/golang-toolkit/util/coll"
	"github.com/gin-gonic/gin"
)

// RouterInfo defines a group and request interceptors; keep it stable while serving.
type RouterInfo struct {
	// GroupPath prefixes registered route paths.
	GroupPath string

	// PreInterceptors execute in order; nonnil Response/error short-circuits this scope.
	PreInterceptors []PreInterceptor
	// PostInterceptors execute in order, including after a route-scope short circuit.
	PostInterceptors []PostInterceptor
}

// Router describes application routes registered synchronously during startup.
type Router interface {
	// Info returns nonnil group information; startup filters nil interceptor entries in place.
	Info() *RouterInfo
	// Handlers registers routes before serving; it must not retain the wrapper for concurrent registration.
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
