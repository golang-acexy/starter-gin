package ginstarter

import (
	"net/http"

	"github.com/acexy/golang-toolkit/logger"
	"github.com/gin-gonic/gin"
)

// RouterWrapper registers routes during startup; its zero value is unusable and registration is not concurrency-safe.
// Handler entries must be nonnil. Gin panics on invalid paths, duplicate routes, or empty handler lists.
type RouterWrapper struct {
	routerGroup *gin.RouterGroup
}

// HandlerWrapper handles one request. A nonnil Response or error stops later handlers.
// Returning nil, nil continues; native committed responses stop later wrapped handlers.
type HandlerWrapper func(request *Request) (Response, error)

// POST registers an HTTP POST route; handlers run in order until a response/error or native commit.
func (r *RouterWrapper) POST(path string, handler ...HandlerWrapper) {
	r.handler([]string{http.MethodPost}, path, nil, handler...)
}

// POST1 registers an HTTP POST route; handlers run in order until a response/error or native commit. Nonempty contentType lists restrict accepted media types.
func (r *RouterWrapper) POST1(path string, contentType []string, handler ...HandlerWrapper) {
	r.handler([]string{http.MethodPost}, path, contentType, handler...)
}

// GET registers an HTTP GET route; handlers run in order until a response/error or native commit.
func (r *RouterWrapper) GET(path string, handler ...HandlerWrapper) {
	r.handler([]string{http.MethodGet}, path, nil, handler...)
}

// HEAD registers an HTTP HEAD route; handlers run in order until a response/error or native commit.
func (r *RouterWrapper) HEAD(path string, handler ...HandlerWrapper) {
	r.handler([]string{http.MethodHead}, path, nil, handler...)
}

// PUT registers an HTTP PUT route; handlers run in order until a response/error or native commit.
func (r *RouterWrapper) PUT(path string, handler ...HandlerWrapper) {
	r.handler([]string{http.MethodPut}, path, nil, handler...)
}

// PUT1 registers an HTTP PUT route; handlers run in order until a response/error or native commit. Nonempty contentType lists restrict accepted media types.
func (r *RouterWrapper) PUT1(path string, contentType []string, handler ...HandlerWrapper) {
	r.handler([]string{http.MethodPut}, path, contentType, handler...)
}

// PATCH registers an HTTP PATCH route; handlers run in order until a response/error or native commit.
func (r *RouterWrapper) PATCH(path string, handler ...HandlerWrapper) {
	r.handler([]string{http.MethodPatch}, path, nil, handler...)
}

// PATCH1 registers an HTTP PATCH route; handlers run in order until a response/error or native commit. Nonempty contentType lists restrict accepted media types.
func (r *RouterWrapper) PATCH1(path string, contentType []string, handler ...HandlerWrapper) {
	r.handler([]string{http.MethodPatch}, path, contentType, handler...)
}

// DELETE registers an HTTP DELETE route; handlers run in order until a response/error or native commit.
func (r *RouterWrapper) DELETE(path string, handler ...HandlerWrapper) {
	r.handler([]string{http.MethodDelete}, path, nil, handler...)
}

// DELETE1 registers an HTTP DELETE route; handlers run in order until a response/error or native commit. Nonempty contentType lists restrict accepted media types.
func (r *RouterWrapper) DELETE1(path string, contentType []string, handler ...HandlerWrapper) {
	r.handler([]string{http.MethodDelete}, path, contentType, handler...)
}

// OPTIONS registers an HTTP OPTIONS route; handlers run in order until a response/error or native commit.
func (r *RouterWrapper) OPTIONS(path string, handler ...HandlerWrapper) {
	r.handler([]string{http.MethodOptions}, path, nil, handler...)
}

// TRACE registers an HTTP TRACE route; handlers run in order until a response/error or native commit.
func (r *RouterWrapper) TRACE(path string, handler ...HandlerWrapper) {
	r.handler([]string{http.MethodTrace}, path, nil, handler...)
}

// TRACE1 registers an HTTP TRACE route; handlers run in order until a response/error or native commit. Nonempty contentType lists restrict accepted media types.
func (r *RouterWrapper) TRACE1(path string, contentType []string, handler ...HandlerWrapper) {
	r.handler([]string{http.MethodTrace}, path, contentType, handler...)
}

// MATCH registers the supplied HTTP methods; handlers run in order until a response/error or native commit.
func (r *RouterWrapper) MATCH(method []string, path string, handler ...HandlerWrapper) {
	r.handler(method, path, nil, handler...)
}

// MATCH1 registers the supplied HTTP methods; handlers run in order until a response/error or native commit. Nonempty contentType lists restrict accepted media types.
func (r *RouterWrapper) MATCH1(method []string, path string, contentType []string, handler ...HandlerWrapper) {
	r.handler(method, path, contentType, handler...)
}

// handler registers wrapped handlers and optional media-type checks.
func (r *RouterWrapper) handler(methods []string, path string, contentType []string, handlerWrapper ...HandlerWrapper) {
	handlers := make([]gin.HandlerFunc, len(handlerWrapper))
	for i, handler := range handlerWrapper {
		handler := handler
		handlers[i] = func(context *gin.Context) {
			defer recoverResponse(context)
			state := getRequestState(context)
			if state.stopped {
				return
			}
			if context.IsAborted() {
				logger.Logrus().Warning("Request is aborted")
				return
			}
			if len(contentType) > 0 {
				if !isMatchMediaType(contentType, context.GetHeader("Content-Type")) {
					panic(&internalPanic{
						statusCode: http.StatusUnsupportedMediaType,
						rawError:   ErrMediaTypeNotAllowed,
					})
				}
			}
			response, err := handler(&Request{context})
			if err != nil {
				panic(err)
			}
			if nativeResponseWritten(context) {
				state.stopped = true
				context.Abort()
				return
			}
			if response != nil {
				state.response = response
				state.stopped = true
				context.Abort()
			}
		}
	}
	r.routerGroup.Match(methods, path, handlers...)
}
