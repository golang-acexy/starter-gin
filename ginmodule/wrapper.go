package ginmodule

import (
	"github.com/acexy/golang-toolkit/logger"
	"github.com/gin-gonic/gin"
	"net/http"
)

type BasicAuthAccount struct {
	Username string
	Password string
}

type RouterInfo struct {
	// GroupPath 路由分组路径
	GroupPath string

	// BasicAuthAccount 如果指定基于BasicAuth认证的账户，则该GroupPath下资源将需要权限认证
	BasicAuthAccount *BasicAuthAccount
}

// RouterWrapper 定义路由包装器
type RouterWrapper struct {
	routerGroup *gin.RouterGroup
}

// HandlerWrapper 定义内部Handler
type HandlerWrapper func(request *Request) (Response, error)

type Router interface {
	// Info 定义路由信息
	Info() *RouterInfo

	// Handlers 注册处理器
	Handlers(router *RouterWrapper)
}

// 定义RouterWrapper的接收请求行为

func (r *RouterWrapper) POST(path string, handler ...HandlerWrapper) {
	r.handler([]string{http.MethodPost}, path, handler...)
}

func (r *RouterWrapper) GET(path string, handler ...HandlerWrapper) {
	r.handler([]string{http.MethodGet}, path, handler...)
}

func (r *RouterWrapper) HEAD(path string, handler ...HandlerWrapper) {
	r.handler([]string{http.MethodHead}, path, handler...)
}

func (r *RouterWrapper) PUT(path string, handler ...HandlerWrapper) {
	r.handler([]string{http.MethodPut}, path, handler...)
}

func (r *RouterWrapper) PATCH(path string, handler ...HandlerWrapper) {
	r.handler([]string{http.MethodPatch}, path, handler...)
}

func (r *RouterWrapper) DELETE(path string, handler ...HandlerWrapper) {
	r.handler([]string{http.MethodDelete}, path, handler...)
}

func (r *RouterWrapper) OPTIONS(path string, handler ...HandlerWrapper) {
	r.handler([]string{http.MethodOptions}, path, handler...)
}

func (r *RouterWrapper) TRACE(path string, handler ...HandlerWrapper) {
	r.handler([]string{http.MethodTrace}, path, handler...)
}

func (r *RouterWrapper) MATCH(method []string, path string, handler ...HandlerWrapper) {
	r.handler(method, path, handler...)
}

// 执行RouterWrapper行为

func (r *RouterWrapper) handler(methods []string, path string, handlerWrapper ...HandlerWrapper) {
	handlers := make([]gin.HandlerFunc, len(handlerWrapper))
	for i, handler := range handlerWrapper {
		handlers[i] = func(context *gin.Context) {
			if context.IsAborted() {
				logger.Logrus().Warning("Request is aborted")
				return
			}

			response, err := handler(&Request{context})
			if err != nil {
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

func httpResponse(context *gin.Context, response Response) {

	if instance, ok := response.(ginRawResp); ok {
		instance.ginFn(context)
		return
	}

	httpStatusCode := response.HttpStatusCode()
	if httpStatusCode == 0 {
		httpStatusCode = http.StatusOK
	}

	if len(response.Headers()) != 0 {
		for k, v := range response.Headers() {
			context.Header(k, v)
		}
	}

	contentType := response.ContentType()
	if contentType == "" {
		contentType = gin.MIMEJSON
		logger.Logrus().Traceln("ContentType is not set, use default", gin.MIMEJSON)
	}

	data := response.Data()
	if v, ok := data.(string); ok {
		context.Data(httpStatusCode, contentType, []byte(v))
	} else if v, ok := data.([]byte); ok {
		context.Data(httpStatusCode, contentType, v)
	} else {
		decode, err := defaultResponseDataDecoder.Decode(data)
		if err != nil {
			panic(err)
		}
		context.Data(httpStatusCode, contentType, decode)
	}
}

// 支持将gin statusCode重写的响应处理器
type responseStatusRewriter struct {
	gin.ResponseWriter
	statusCode int
}

func (r *responseStatusRewriter) WriteHeader(code int) {
	r.statusCode = code
}

func (r *responseStatusRewriter) Write(data []byte) (int, error) {
	if !r.Written() {
		r.ResponseWriter.WriteHeader(r.statusCode)
	}
	return r.ResponseWriter.Write(data)
}

func (r *responseStatusRewriter) Status() int {
	return r.statusCode
}
