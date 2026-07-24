package router

import (
	"errors"
	"net/http"
	"strings"

	"github.com/acexy/golang-toolkit/logger"
	"github.com/golang-acexy/starter-gin/ginstarter"
)

const interceptorTestPathPrefix = "/interceptor-test/"

// InterceptorTestRouter 验证全局拦截器、路由拦截器、Handler 和异常响应之间的执行链路。
//
// 正常链路预期顺序：
// GLOBAL_PRE_A -> GLOBAL_PRE_B -> ROUTER_PRE_A -> ROUTER_PRE_B -> HANDLER ->
// ROUTER_POST_A -> ROUTER_POST_B -> GLOBAL_POST_A -> GLOBAL_POST_B
//
// 测试命令：
//
// curl -i 'http://127.0.0.1:8080/interceptor-test/run?case=normal'
// curl -i 'http://127.0.0.1:8080/interceptor-test/run?case=global-pre-response'
// curl -i 'http://127.0.0.1:8080/interceptor-test/run?case=global-pre-error'
// curl -i 'http://127.0.0.1:8080/interceptor-test/run?case=global-pre-panic'
// curl -i 'http://127.0.0.1:8080/interceptor-test/run?case=router-pre-response'
// curl -i 'http://127.0.0.1:8080/interceptor-test/run?case=router-pre-error'
// curl -i 'http://127.0.0.1:8080/interceptor-test/run?case=router-pre-panic'
// curl -i 'http://127.0.0.1:8080/interceptor-test/run?case=handler-error'
// curl -i 'http://127.0.0.1:8080/interceptor-test/run?case=handler-panic'
// curl -i 'http://127.0.0.1:8080/interceptor-test/run?case=bad-status'
// curl -i 'http://127.0.0.1:8080/interceptor-test/run?case=router-post-replace'
// curl -i 'http://127.0.0.1:8080/interceptor-test/run?case=router-post-error'
// curl -i 'http://127.0.0.1:8080/interceptor-test/run?case=router-post-panic'
// curl -i 'http://127.0.0.1:8080/interceptor-test/run?case=global-post-replace'
// curl -i 'http://127.0.0.1:8080/interceptor-test/run?case=global-post-error'
// curl -i 'http://127.0.0.1:8080/interceptor-test/run?case=global-post-panic'
// curl -i 'http://127.0.0.1:8080/interceptor-test/run?case=native-response'
type InterceptorTestRouter struct{}

func (i *InterceptorTestRouter) Info() *ginstarter.RouterInfo {
	return &ginstarter.RouterInfo{
		GroupPath: "interceptor-test",
		PreInterceptors: []ginstarter.PreInterceptor{
			interceptorTestRouterPreA,
			interceptorTestRouterPreB,
		},
		PostInterceptors: []ginstarter.PostInterceptor{
			interceptorTestRouterPostA,
			interceptorTestRouterPostB,
		},
	}
}

func (i *InterceptorTestRouter) Handlers(router *ginstarter.RouterWrapper) {
	router.GET("run", i.run())
}

// InterceptorTestGlobalPreInterceptors 返回拦截器测试使用的全局前置拦截器。
// 拦截器只处理 /interceptor-test/ 下的请求，不影响其他测试路由。
func InterceptorTestGlobalPreInterceptors() []ginstarter.PreInterceptor {
	return []ginstarter.PreInterceptor{
		interceptorTestGlobalPreA,
		interceptorTestGlobalPreB,
	}
}

// InterceptorTestGlobalPostInterceptors 返回拦截器测试使用的全局后置拦截器。
// 拦截器只处理 /interceptor-test/ 下的请求，不影响其他测试路由。
func InterceptorTestGlobalPostInterceptors() []ginstarter.PostInterceptor {
	return []ginstarter.PostInterceptor{
		interceptorTestGlobalPostA,
		interceptorTestGlobalPostB,
	}
}

func interceptorTestGlobalPreA(request *ginstarter.Request) (ginstarter.Response, error) {
	if !isInterceptorTestRequest(request) {
		return nil, nil
	}
	testCase := interceptorTestCase(request)
	logInterceptorStage(request, "01 GLOBAL_PRE_A", nil)
	switch testCase {
	case "global-pre-response":
		return testResponse("global-pre-response", http.StatusUnauthorized), nil
	case "global-pre-error":
		return nil, errors.New("global pre interceptor error")
	case "global-pre-panic":
		panic("global pre interceptor panic")
	default:
		return nil, nil
	}
}

func interceptorTestGlobalPreB(request *ginstarter.Request) (ginstarter.Response, error) {
	if isInterceptorTestRequest(request) {
		logInterceptorStage(request, "02 GLOBAL_PRE_B", nil)
	}
	return nil, nil
}

func interceptorTestRouterPreA(request *ginstarter.Request) (ginstarter.Response, error) {
	testCase := interceptorTestCase(request)
	logInterceptorStage(request, "03 ROUTER_PRE_A", nil)
	switch testCase {
	case "router-pre-response":
		return testResponse("router-pre-response", http.StatusForbidden), nil
	case "router-pre-error":
		return nil, errors.New("router pre interceptor error")
	case "router-pre-panic":
		panic("router pre interceptor panic")
	default:
		return nil, nil
	}
}

func interceptorTestRouterPreB(request *ginstarter.Request) (ginstarter.Response, error) {
	logInterceptorStage(request, "04 ROUTER_PRE_B", nil)
	return nil, nil
}

func (i *InterceptorTestRouter) run() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		testCase := interceptorTestCase(request)
		logInterceptorStage(request, "05 HANDLER", nil)
		switch testCase {
		case "handler-error":
			return nil, errors.New("handler error")
		case "handler-panic":
			panic("handler panic")
		case "bad-status":
			return testResponse("handler-bad-status", http.StatusTeapot), nil
		case "native-response":
			request.GinContext().String(http.StatusAccepted, "native-response")
			return testResponse("framework-response-must-be-ignored", http.StatusOK), nil
		default:
			return testResponse("handler-response", http.StatusOK), nil
		}
	}
}

func interceptorTestRouterPostA(request *ginstarter.Request, response ginstarter.Response) (ginstarter.Response, error) {
	testCase := interceptorTestCase(request)
	logInterceptorStage(request, "06 ROUTER_POST_A", response)
	switch testCase {
	case "router-post-replace":
		return testResponse("router-post-replaced", http.StatusOK), nil
	case "router-post-error":
		return nil, errors.New("router post interceptor error")
	case "router-post-panic":
		panic("router post interceptor panic")
	default:
		return nil, nil
	}
}

func interceptorTestRouterPostB(request *ginstarter.Request, response ginstarter.Response) (ginstarter.Response, error) {
	logInterceptorStage(request, "07 ROUTER_POST_B", response)
	return nil, nil
}

func interceptorTestGlobalPostA(request *ginstarter.Request, response ginstarter.Response) (ginstarter.Response, error) {
	if !isInterceptorTestRequest(request) {
		return nil, nil
	}
	testCase := interceptorTestCase(request)
	logInterceptorStage(request, "08 GLOBAL_POST_A", response)
	switch testCase {
	case "global-post-replace":
		return testResponse("global-post-replaced", http.StatusOK), nil
	case "global-post-error":
		return nil, errors.New("global post interceptor error")
	case "global-post-panic":
		panic("global post interceptor panic")
	default:
		return nil, nil
	}
}

func interceptorTestGlobalPostB(request *ginstarter.Request, response ginstarter.Response) (ginstarter.Response, error) {
	if isInterceptorTestRequest(request) {
		logInterceptorStage(request, "09 GLOBAL_POST_B", response)
	}
	return nil, nil
}

func isInterceptorTestRequest(request *ginstarter.Request) bool {
	return strings.HasPrefix(request.RequestPath(), interceptorTestPathPrefix)
}

func interceptorTestCase(request *ginstarter.Request) string {
	testCase, exists := request.GetQueryParam("case")
	if !exists || testCase == "" {
		return "normal"
	}
	return testCase
}

func logInterceptorStage(request *ginstarter.Request, stage string, response ginstarter.Response) {
	responseInfo := "nil"
	if response != nil && response.Data() != nil {
		responseInfo = response.Data().DebugString(512)
	}
	logger.Logrus().Infof(
		"[interceptor-test] case=%s stage=%s response=%s",
		interceptorTestCase(request),
		stage,
		responseInfo,
	)
}

func testResponse(body string, statusCode int) ginstarter.Response {
	return ginstarter.RespTextPlain([]byte(body), statusCode)
}
