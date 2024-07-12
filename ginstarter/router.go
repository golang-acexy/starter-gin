package ginstarter

import (
	"github.com/gin-gonic/gin"
)

func registerRouter(g *gin.Engine, routers []Router) {
	for _, v := range routers {
		routerInfo := v.Info()
		if len(routerInfo.Middlewares) > 0 {
			group := g.Group(routerInfo.GroupPath)
			for _, m := range routerInfo.Middlewares {
				group.Use(func(ctx *gin.Context) {
					response, continueExecute := m(&Request{ctx: ctx})
					if !continueExecute {
						httpResponse(ctx, response)
						ctx.Abort()
						return
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
