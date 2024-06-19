package ginmodule

import (
	"github.com/gin-gonic/gin"
)

func registerRouter(g *gin.Engine, routers []Router) {
	for _, v := range routers {
		routerInfo := v.Info()
		if routerInfo.BasicAuthAccount != nil {
			v.Handlers(&RouterWrapper{routerGroup: g.Group(routerInfo.GroupPath,
				gin.BasicAuth(map[string]string{routerInfo.BasicAuthAccount.Username: routerInfo.BasicAuthAccount.Password}),
			)})
		} else {
			v.Handlers(&RouterWrapper{routerGroup: g.Group(routerInfo.GroupPath)})
		}
	}
}
