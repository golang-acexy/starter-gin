package ginstarter

import (
	"github.com/gin-gonic/gin"
)

func registerRouter(g *gin.Engine, routers []Router) {
	for _, v := range routers {
		routerInfo := v.Info()
		if len(routerInfo.Middlewares) > 0 {
			group := g.Group(routerInfo.GroupPath)
			for i := range routerInfo.Middlewares {
				middleware := routerInfo.Middlewares[i]
				group.Use(func(ctx *gin.Context) {
					response, continued := middleware(&Request{ctx: ctx})
					if !continued {
						httpResponse(ctx, response)
						ctx.Abort()
					} else {
						ctx.Next()
					}
				})
				v.Handlers(&RouterWrapper{routerGroup: group})
			}
		} else {
			v.Handlers(&RouterWrapper{routerGroup: g.Group(routerInfo.GroupPath)})
		}
	}
}
