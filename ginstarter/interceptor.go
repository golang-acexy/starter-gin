package ginstarter

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/acexy/golang-toolkit/logger"
	"github.com/acexy/golang-toolkit/math/conversion"
	"github.com/acexy/golang-toolkit/util/coll"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var (
	httpCodeWithStatus          map[int]StatusCode
	defaultIgnoreHTTPStatusCode = []int{
		http.StatusMultipleChoices,
		http.StatusMovedPermanently,
		http.StatusFound,
		http.StatusNoContent,
		http.StatusNotModified,
		http.StatusUseProxy,
		http.StatusTemporaryRedirect,
		http.StatusPermanentRedirect,
	}

	panicResolver PanicResolver = func(err error) string {
		return err.Error()
	}
	badHTTPCodeResolver BadHTTPCodeResolver = func(httpStatusCode int, errMsg string) Response {
		var statusMessage StatusMessage
		if errMsg != "" {
			statusMessage = StatusMessage(errMsg)
		}
		body := RestRespStruct{
			Status: &RestRespStatusStruct{
				Timestamp: time.Now().UnixMilli(),
			},
		}
		var statusCode StatusCode
		if v, ok := httpCodeWithStatus[httpStatusCode]; ok {
			statusCode = v
		} else {
			statusCode = StatusCodeException
		}
		if statusMessage == "" {
			body.Status.StatusMessage = GetStatusMessage(statusCode)
		} else {
			body.Status.StatusMessage = statusMessage
		}
		body.Status.StatusCode = statusCode
		return NewRespRest().DataBuilder(func() *ResponseEntity {
			bodyBytes, err := currentResponseBodyEncoder().Encode(body)
			if err != nil {
				return NewResponseEntityWithStatusCode(gin.MIMEPlain, []byte(statusMessageException), http.StatusOK)
			}
			return NewResponseEntityWithStatusCode(gin.MIMEJSON, bodyBytes, http.StatusOK)
		})
	}
)

type BasicAuthAccount struct {
	Username string
	Password string
}

// 定义内部panic 用于特殊处理 中断请求流程
type internalPanic struct {
	statusCode int
	rawError   error
}

// PreInterceptor 前置拦截器。
// 返回非空 Response 或 error 时终止后续前置拦截器及 Handler，并进入后置拦截器链。
type PreInterceptor func(request *Request) (Response, error)

// PostInterceptor 后置拦截器。
// 返回非空 Response 时替换当前响应；返回 error 时转换为异常响应并停止当前后置拦截器链。
type PostInterceptor func(request *Request, response Response) (Response, error)

type PanicResolver func(err error) string
type BadHTTPCodeResolver func(httpStatusCode int, errMsg string) Response

func init() {
	httpCodeWithStatus = make(map[int]StatusCode, 7)
	httpCodeWithStatus[http.StatusBadRequest] = StatusCodeBadRequestParameters
	httpCodeWithStatus[http.StatusForbidden] = StatusCodeForbidden
	httpCodeWithStatus[http.StatusNotFound] = StatusCodeNotFound
	httpCodeWithStatus[http.StatusMethodNotAllowed] = StatusCodeMethodNotAllowed
	httpCodeWithStatus[http.StatusUnsupportedMediaType] = StatusCodeMediaTypeNotAllowed
	httpCodeWithStatus[http.StatusRequestEntityTooLarge] = StatusCodeUploadLimitExceeded
	httpCodeWithStatus[http.StatusUnauthorized] = StatusCodeUnauthorized
}

func isIgnoreHTTPStatusCode(httpCode int) bool {
	config := currentGinConfig()
	if config == nil {
		return false
	}
	if !config.DisableDefaultIgnoreHTTPCode {
		for _, v := range defaultIgnoreHTTPStatusCode {
			if httpCode == v {
				return true
			}
		}
	}
	if len(config.IgnoreHTTPCode) > 0 {
		for _, v := range config.IgnoreHTTPCode {
			if httpCode == v {
				return true
			}
		}
	}
	return false
}

func panicToError(panicError any) (statusCode int, err error, internalError bool) {
	switch t := panicError.(type) {
	case string:
		err = errors.New(t)
	case error:
		err = t
	default:
		// 内部特殊错误
		if v, ok := t.(*internalPanic); ok {
			rawError := v.rawError
			statusCode = v.statusCode
			if validationErrs, ok := rawError.(validator.ValidationErrors); ok {
				internalError = true
				err = errors.New(friendlyValidatorMessage(validationErrs))
			} else if jsonErr, ok := rawError.(*json.UnmarshalTypeError); ok {
				err = fmt.Errorf("%w: %s", ErrJSONTypeMismatch, jsonErr.Field)
			} else if _, ok := rawError.(*json.SyntaxError); ok {
				err = ErrBadJSONPayload
			} else {
				err = rawError
			}
		} else {
			err = fmt.Errorf("%v", t)
		}
	}
	// 明确的 4xx 表示请求本身不符合约束，属于可预期的客户端错误，不输出异常调用栈。
	if statusCode >= http.StatusBadRequest && statusCode < http.StatusInternalServerError {
		logger.Logrus().Warnf("request rejected: %v", err)
	} else if !internalError {
		stack := string(debug.Stack())
		lines := strings.Split(stack, "\n")
		index := coll.SliceIndexBy(lines, func(line string) bool {
			return strings.Contains(line, "runtime/panic.go")
		})
		filter := lines
		if index != -1 {
			filter = lines[index:]
		}
		index = coll.SliceIndexBy(filter, func(line string) bool {
			return strings.Contains(line, "ginstarter/wrapper.go")
		})
		if index != -1 {
			index = coll.SliceIndexBy(filter, func(line string) bool {
				return strings.Contains(line, "ginstarter/interceptor.go")
			})
			if index != -1 {
				stack = strings.Join(filter[:index], "\n")
			}
		}
		logger.Logrus().Errorf("panic: %v %s", err, stack)
	} else {
		logger.Logrus().Errorf("panic: %v", err)
	}
	return
}

// responseTracker 记录业务代码是否直接操作了 Gin 原生响应。
// 仅设置 Header 不视为提交响应，框架最终响应仍会保留这些 Header。
type responseTracker struct {
	gin.ResponseWriter
	nativeWritten bool
}

func (r *responseTracker) WriteHeader(code int) {
	r.nativeWritten = true
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseTracker) Write(data []byte) (int, error) {
	r.nativeWritten = true
	return r.ResponseWriter.Write(data)
}

func (r *responseTracker) WriteString(data string) (int, error) {
	r.nativeWritten = true
	return r.ResponseWriter.WriteString(data)
}

func nativeResponseWritten(context *gin.Context) bool {
	if tracker, ok := context.Writer.(*responseTracker); ok && tracker.nativeWritten {
		return true
	}
	return context.Writer.Written()
}

// resolvePanic 将 panic 转换为框架 Response，不直接写入客户端。
func resolvePanic(panicValue any) Response {
	config := currentGinConfig()
	status, err, safeToExpose := panicToError(panicValue)
	if status == 0 {
		status = http.StatusInternalServerError
	}
	errMsg := ""
	if config == nil {
		return RespTextPlain([]byte(err.Error()), status)
	}
	if !config.HidePanicErrorDetails || safeToExpose {
		errMsg = config.PanicResolver(err)
	}
	if !config.DisableBadHTTPCodeResolver && !isIgnoreHTTPStatusCode(status) {
		return config.BadHTTPCodeResolver(status, errMsg)
	}
	return RespTextPlain([]byte(errMsg), status)
}

func recoverResponse(context *gin.Context) {
	if panicValue := recover(); panicValue != nil {
		setResponse(context, resolvePanic(panicValue))
		getRequestState(context).stopped = true
		context.Abort()
	}
}

func nextWithRecovery(context *gin.Context) {
	defer recoverResponse(context)
	context.Next()
}

func runPreInterceptors(context *gin.Context, interceptors []PreInterceptor) {
	state := getRequestState(context)
	for _, interceptor := range interceptors {
		var response Response
		var interceptorErr error
		func() {
			defer func() {
				if panicValue := recover(); panicValue != nil {
					response = resolvePanic(panicValue)
				}
			}()
			response, interceptorErr = interceptor(&Request{ctx: context})
		}()
		if interceptorErr != nil {
			response = resolvePanic(interceptorErr)
		}
		if response != nil {
			state.response = response
			state.stopped = true
			context.Abort()
			return
		}
	}
}

func runPostInterceptors(context *gin.Context, interceptors []PostInterceptor) {
	state := getRequestState(context)
	for _, interceptor := range interceptors {
		var response Response
		var interceptorErr error
		failed := false
		func() {
			defer func() {
				if panicValue := recover(); panicValue != nil {
					response = resolvePanic(panicValue)
					failed = true
				}
			}()
			response, interceptorErr = interceptor(&Request{ctx: context}, state.response)
		}()
		if interceptorErr != nil {
			response = resolvePanic(interceptorErr)
			failed = true
		}
		if response != nil {
			state.response = response
		}
		if failed {
			return
		}
	}
}

func normalizeResponse(context *gin.Context) Response {
	response := currentResponse(context)
	config := currentGinConfig()
	if response == nil || response.Data() == nil || config == nil || config.DisableBadHTTPCodeResolver {
		return response
	}
	statusCode := response.Data().statusCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	if statusCode == http.StatusOK || isIgnoreHTTPStatusCode(statusCode) {
		return response
	}
	logger.Logrus().Warningln("Bad response path:", context.Request.URL, "status code:", statusCode)
	response = config.BadHTTPCodeResolver(statusCode, "")
	setResponse(context, response)
	return response
}

// responsePipelineHandler 负责捕获遗漏异常、规范化状态并且只提交一次框架响应。
func responsePipelineHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		config := currentGinConfig()
		tracker := &responseTracker{ResponseWriter: ctx.Writer}
		ctx.Writer = tracker
		if config != nil && config.MaxRequestBodyBytes > 0 && ctx.Request.Body != nil {
			ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, config.MaxRequestBodyBytes)
		}
		getRequestState(ctx)
		defer func() {
			if panicValue := recover(); panicValue != nil {
				setResponse(ctx, resolvePanic(panicValue))
			}
			// 原生 Gin 响应一旦写入，框架不再尝试覆盖。
			if nativeResponseWritten(ctx) || tracker.ResponseWriter.Written() {
				return
			}
			writeResponse(ctx, normalizeResponse(ctx))
		}()
		ctx.Next()
	}
}

// 常用的一些中间件

// BasicAuthInterceptor 基础权限校验中间件
// match 满足指定条件才执行
func BasicAuthInterceptor(account *BasicAuthAccount, match ...func(request *Request) bool) PreInterceptor {
	return func(request *Request) (Response, error) {
		if len(match) > 0 {
			if !match[0](request) {
				return nil, nil
			}
		}
		if request.GetHeader("Authorization") == "" {
			return RespHTTPStatusCode(http.StatusUnauthorized), nil
		}
		enc := "Basic " + base64.StdEncoding.EncodeToString(conversion.ParseBytes(account.Username+":"+account.Password))
		if request.GetHeader("Authorization") != enc {
			return RespHTTPStatusCode(http.StatusUnauthorized), nil
		}
		return nil, nil
	}
}

// MediaTypeInterceptor ContentType校验中间件
func MediaTypeInterceptor(contentType []string, match ...func(request *Request) bool) PreInterceptor {
	return func(request *Request) (Response, error) {
		if len(match) > 0 {
			if !match[0](request) {
				return nil, nil
			}
		}
		if len(contentType) > 0 {
			if !isMatchMediaType(contentType, request.GetHeader("Content-Type")) {
				return RespHTTPStatusCode(http.StatusUnsupportedMediaType), nil
			}
		} else {
			logger.Logrus().Warningln("valid Content-Type restriction not set")
		}
		return nil, nil
	}
}

func isMatchMediaType(allowContentType []string, requestContentType string) bool {
	requestMediaType, err := parseMediaType(requestContentType)
	if err != nil {
		return false
	}
	for _, allowedContentType := range allowContentType {
		allowedMediaType, parseErr := parseMediaType(allowedContentType)
		if parseErr == nil && allowedMediaType == requestMediaType {
			return true
		}
	}
	return false
}
