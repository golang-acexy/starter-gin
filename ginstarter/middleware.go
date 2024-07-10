package ginstarter

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/acexy/golang-toolkit/logger"
	"github.com/acexy/golang-toolkit/math/conversion"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
)

var defaultIgnoreHttpStatusCode = []int{
	http.StatusMultipleChoices,
	http.StatusMovedPermanently,
	http.StatusFound,
	http.StatusNoContent,
	http.StatusNotModified,
	http.StatusUseProxy,
	http.StatusTemporaryRedirect,
	http.StatusPermanentRedirect,
}

type ExceptionResolver interface {
	Resolve(httpStatusCode int, err error) Response
}

// 默认异常处理器
type panicResolver struct {
}

func (*panicResolver) Resolve(httpStatusCode int, err error) Response {
	if httpStatusCode != http.StatusOK { // 包裹异常状态码响应
		if ginStarter.DisableBadHttpCodeResolver { // 禁用httpCode包裹功能
			if err == nil {
				return RespAbortWithHttpStatusCode(httpStatusCode)
			}
			return RespTextPlain(err.Error(), httpStatusCode)
		} else {
			var statusMessage StatusMessage
			if err != nil {
				statusMessage = StatusMessage(err.Error())
			}
			if v, ok := httpCodeWithStatus[httpStatusCode]; ok {
				return RespRestStatusError(v, statusMessage)
			} else {
				return RespRestStatusError(StatusCodeException, statusMessage)
			}
		}
	}
	return nil
}

// 默认异常处理器
type badHttpCodeResolver struct {
}

func (*badHttpCodeResolver) Resolve(httpStatusCode int, err error) Response {
	if httpStatusCode != http.StatusOK { // 包裹异常状态码响应
		if ginStarter.DisableBadHttpCodeResolver { // 禁用httpCode包裹功能
			if err == nil {
				return RespAbortWithHttpStatusCode(httpStatusCode)
			}
			return RespTextPlain(err.Error(), httpStatusCode)
		} else {
			var statusMessage StatusMessage
			if err != nil {
				statusMessage = StatusMessage(err.Error())
			}
			if v, ok := httpCodeWithStatus[httpStatusCode]; ok {
				return RespRestStatusError(v, statusMessage)
			} else {
				return RespRestStatusError(StatusCodeException, statusMessage)
			}
		}
	}
	return nil
}

var (
	// 默认处理状态码handler的响应执行器
	defaultHttpStatusCodeHandlerResponse HttpStatusCodeCodeHandlerResponse = func(ctx *gin.Context, httpStatusCode int) Response {
		logger.Logrus().Warningln("Bad response path:", ctx.Request.URL, "status code:", httpStatusCode)
		v, ok := httpCodeWithStatus[httpStatusCode]
		if !ok {
			return RespRestStatusError(StatusCodeException)
		} else {
			return RespRestStatusError(v)
		}
	}

	httpCodeWithStatus map[int]StatusCode
)

type HttpStatusCodeCodeHandlerResponse func(ctx *gin.Context, httpStatusCode int) Response

func init() {
	httpCodeWithStatus = make(map[int]StatusCode, 7)
	httpCodeWithStatus[http.StatusBadRequest] = StatusCodeBadRequestParameters
	httpCodeWithStatus[http.StatusForbidden] = StatusCodeForbidden
	httpCodeWithStatus[http.StatusNotFound] = StatusCodeNotFound
	httpCodeWithStatus[http.StatusMethodNotAllowed] = StatusCodeMethodNotAllowed
	httpCodeWithStatus[http.StatusUnsupportedMediaType] = StatusCodeMediaTypeNotAllowed
	httpCodeWithStatus[http.StatusRequestEntityTooLarge] = StatusCodeUploadLimitExceeded
	httpCodeWithStatus[http.StatusUnauthorized] = StatusCodeForbidden
}

// recoverHandler 全局Panic处理中间件
func recoverHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		// panic异常处理
		defer func() {
			var err any
			if r := recover(); r != nil {
				err = r
			} else {
				if ctx.Err() != nil {
					err = ctx.Err()
				}
			}
			// 内部特殊错误
			if v, ok := err.(*internalPanic); ok {
				err = v.rawError
				if validationErrs, ok := err.(validator.ValidationErrors); ok {
					for _, vErr := range validationErrs {
						field := vErr.Field()
						fmt.Println(field + " " + vErr.Tag())
					}
				}
			}

			writer := ctx.Writer
			//ctx.Status(v.statusCode)
			// 如果使用了可覆写中间件
			if w, ok := writer.(*responseRewriter); ok {
				w.ResponseWriter.WriteHeader(w.statusCode)
				_, _ = w.ResponseWriter.Write(w.body.Bytes())
				return
			}
		}()

		// 异常响应码处理
		if !ginStarter.DisableBadHttpCodeResolver {
			writer := ctx.Writer
			var statusCode int
			if v, ok := writer.(*responseRewriter); ok {
				if v.statusCode != 0 && v.statusCode != http.StatusOK {
					statusCode = v.statusCode
				} else {
					statusCode = v.ResponseWriter.Status()
				}
			} else {
				statusCode = ctx.Writer.Status()
			}

			if statusCode != http.StatusOK {
				if isIgnoreHttpStatusCode(statusCode) {
					return
				}
				logger.Logrus().Warningln("Bad response path:", ctx.Request.URL, "status code:", statusCode)

				ginStarter.BadHttpCodeResolver.Resolve(statusCode, nil)

				response := defaultHttpStatusCodeHandlerResponse(ctx, statusCode)
				httpResponse(ctx, response)
			}
		}

		ctx.Next()
	}
}

// responseRewriteHandler 可重写Http状态码中间件
func responseRewriteHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		writer := &responseRewriter{
			ResponseWriter: ctx.Writer,
			body:           bytes.NewBufferString(""),
		}
		ctx.Writer = writer
		ctx.Next()
		if writer.statusCode == 0 { // 未设置自定义状态码
			writer.statusCode = writer.ResponseWriter.Status()
		}
		writer.ResponseWriter.WriteHeader(writer.statusCode)
		if writer.body.Len() > 0 {
			_, err := writer.ResponseWriter.Write(writer.body.Bytes())
			if err != nil {
				panic(err)
			}
		}
	}
}

// httpStatusCodeHandler 异常状态码自动转换响应中间件
func httpStatusCodeHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
		writer := ctx.Writer
		var statusCode int
		if v, ok := writer.(*responseRewriter); ok {
			if v.statusCode != 0 && v.statusCode != http.StatusOK {
				statusCode = v.statusCode
			} else {
				statusCode = v.ResponseWriter.Status()
			}
		}
		if statusCode != http.StatusOK {
			if isIgnoreHttpStatusCode(statusCode) {
				return
			}
			if !isIgnoreHttpStatusCode(statusCode) {
				logger.Logrus().Warningln("Bad response path:", ctx.Request.URL, "status code:", statusCode)
			}
			response := defaultHttpStatusCodeHandlerResponse(ctx, statusCode)
			httpResponse(ctx, response)
		}
	}
}

func isIgnoreHttpStatusCode(httpCode int) bool {
	if !ginStarter.DisableDefaultIgnoreHttpCode {
		for _, v := range defaultIgnoreHttpStatusCode {
			if httpCode == v {
				return true
			}
		}
	}
	if len(ginStarter.IgnoreHttpCode) > 0 {
		for _, v := range ginStarter.IgnoreHttpCode {
			if httpCode == v {
				return true
			}
		}
	}
	return false
}

// 常用的一些中间件

// BasicAuthMiddleware 基础权限校验中间件
// match 满足指定条件才执行
func BasicAuthMiddleware(account *BasicAuthAccount, match ...func(request *Request) bool) Middleware {
	return func(request *Request) (Response, bool) {
		if len(match) > 0 {
			if !match[0](request) {
				return nil, true
			}
		}
		if request.GetHeader("Authorization") == "" {
			return RespAbortWithHttpStatusCode(http.StatusUnauthorized), false
		}
		enc := "Basic " + base64.StdEncoding.EncodeToString(conversion.ParseBytes(account.Username+":"+account.Password))
		if request.GetHeader("Authorization") != enc {
			return RespAbortWithHttpStatusCode(http.StatusUnauthorized), false
		}
		return nil, true
	}
}
