package ginstarter

import (
	"errors"
	"github.com/acexy/golang-toolkit/logger"
	"github.com/gin-gonic/gin"
	"net/http"
)

// RouterWrapper 定义路由包装器
type RouterWrapper struct {
	routerGroup *gin.RouterGroup
}

// HandlerWrapper 定义内部Handler
type HandlerWrapper func(request *Request) (Response, error)

// 定义RouterWrapper的接收请求行为

func (r *RouterWrapper) POST(path string, handler ...HandlerWrapper) {
	r.handler([]string{http.MethodPost}, path, nil, handler...)
}

func (r *RouterWrapper) POST1(path string, contentType []string, handler ...HandlerWrapper) {
	r.handler([]string{http.MethodPost}, path, contentType, handler...)
}

func (r *RouterWrapper) GET(path string, handler ...HandlerWrapper) {
	r.handler([]string{http.MethodGet}, path, nil, handler...)
}

func (r *RouterWrapper) HEAD(path string, handler ...HandlerWrapper) {
	r.handler([]string{http.MethodHead}, path, nil, handler...)
}

func (r *RouterWrapper) PUT(path string, handler ...HandlerWrapper) {
	r.handler([]string{http.MethodPut}, path, nil, handler...)
}
func (r *RouterWrapper) PUT1(path string, contentType []string, handler ...HandlerWrapper) {
	r.handler([]string{http.MethodPut}, path, contentType, handler...)
}

func (r *RouterWrapper) PATCH(path string, handler ...HandlerWrapper) {
	r.handler([]string{http.MethodPatch}, path, nil, handler...)
}
func (r *RouterWrapper) PATCH1(path string, contentType []string, handler ...HandlerWrapper) {
	r.handler([]string{http.MethodPatch}, path, contentType, handler...)
}

func (r *RouterWrapper) DELETE(path string, handler ...HandlerWrapper) {
	r.handler([]string{http.MethodDelete}, path, nil, handler...)
}
func (r *RouterWrapper) DELETE1(path string, contentType []string, handler ...HandlerWrapper) {
	r.handler([]string{http.MethodDelete}, path, contentType, handler...)
}

func (r *RouterWrapper) OPTIONS(path string, handler ...HandlerWrapper) {
	r.handler([]string{http.MethodOptions}, path, nil, handler...)
}

func (r *RouterWrapper) TRACE(path string, handler ...HandlerWrapper) {
	r.handler([]string{http.MethodTrace}, path, nil, handler...)
}
func (r *RouterWrapper) TRACE1(path string, contentType []string, handler ...HandlerWrapper) {
	r.handler([]string{http.MethodTrace}, path, contentType, handler...)
}

func (r *RouterWrapper) MATCH(method []string, path string, handler ...HandlerWrapper) {
	r.handler(method, path, nil, handler...)
}
func (r *RouterWrapper) MATCH1(method []string, path string, contentType []string, handler ...HandlerWrapper) {
	r.handler(method, path, contentType, handler...)
}

// 执行RouterWrapper行为

func (r *RouterWrapper) handler(methods []string, path string, contentType []string, handlerWrapper ...HandlerWrapper) {
	handlers := make([]gin.HandlerFunc, len(handlerWrapper))
	for i, handler := range handlerWrapper {
		handlers[i] = func(context *gin.Context) {
			v, exists := context.Get(ginCtxKeyContinueHandler)
			if exists && !v.(bool) {
				return
			}
			if context.IsAborted() {
				logger.Logrus().Warning("Request is aborted")
				return
			}
			if len(contentType) > 0 {
				if !isMatchMediaType(contentType, context.ContentType()) {
					panic(&internalPanic{
						statusCode: http.StatusUnsupportedMediaType,
						rawError:   errors.New(statusMessageMediaTypeNotAllowed),
					})
				}
			}
			response, err := handler(&Request{context})
			if err != nil {
				context.Status(http.StatusInternalServerError)
				panic(err)
			}
			if response != nil {
				httpResponse(context, response)
			} else {
				context.Status(http.StatusOK)
			}
		}
	}
	r.routerGroup.Match(methods, path, handlers...)
}
