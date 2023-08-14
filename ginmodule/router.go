package ginmodule

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type GinWrapper struct {
	routerGroup *gin.RouterGroup
}

type HandlerWrapper func(request *Request) (*Response, error)

type RouterInfo struct {
	GroupPath string
	// 如果指定基于BasicAuth认证的账户，则该GroupPath下资源将需要权限认证 如果不满足验证规则，则会返回相应的httpStatus错误码，并且不会被本框架包装
	BasicAuthAccount map[string]string
}

type Router interface {
	RouterInfo() *RouterInfo
	RegisterHandler(ginWrapper *GinWrapper)
}

func loadRouter(g *gin.Engine, routers []Router) {
	for _, v := range routers {
		routerInfo := v.RouterInfo()
		if len(routerInfo.BasicAuthAccount) > 0 {
			v.RegisterHandler(&GinWrapper{routerGroup: g.Group(routerInfo.GroupPath, gin.BasicAuth(routerInfo.BasicAuthAccount))})
		} else {
			v.RegisterHandler(&GinWrapper{routerGroup: g.Group(routerInfo.GroupPath)})
		}
	}
}

func (g *GinWrapper) handler(methods []string, path string, handlerWrapper ...HandlerWrapper) {
	handlers := make([]gin.HandlerFunc, len(handlerWrapper))
	for i, handler := range handlerWrapper {
		handlers[i] = func(context *gin.Context) {
			response, err := handler(&Request{context})
			if err != nil {
				panic(err)
			}
			if !context.IsAborted() {
				context.JSON(http.StatusOK, response)
			}
		}
	}
	g.routerGroup.Match(methods, path, handlers...)
}

func (g *GinWrapper) POST(path string, handler ...HandlerWrapper) {
	g.handler([]string{http.MethodPost}, path, handler...)
}
func (g *GinWrapper) GET(path string, handler ...HandlerWrapper) {
	g.handler([]string{http.MethodGet}, path, handler...)
}
func (g *GinWrapper) HEAD(path string, handler ...HandlerWrapper) {
	g.handler([]string{http.MethodHead}, path, handler...)
}
func (g *GinWrapper) PUT(path string, handler ...HandlerWrapper) {
	g.handler([]string{http.MethodPut}, path, handler...)
}
func (g *GinWrapper) PATCH(path string, handler ...HandlerWrapper) {
	g.handler([]string{http.MethodPatch}, path, handler...)
}
func (g *GinWrapper) DELETE(path string, handler ...HandlerWrapper) {
	g.handler([]string{http.MethodDelete}, path, handler...)
}
func (g *GinWrapper) OPTIONS(path string, handler ...HandlerWrapper) {
	g.handler([]string{http.MethodOptions}, path, handler...)
}
func (g *GinWrapper) TRACE(path string, handler ...HandlerWrapper) {
	g.handler([]string{http.MethodTrace}, path, handler...)
}
func (g *GinWrapper) MATCH(method []string, path string, handler ...HandlerWrapper) {
	g.handler(method, path, handler...)
}
