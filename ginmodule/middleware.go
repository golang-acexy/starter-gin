package ginmodule

import (
	"bytes"
	"github.com/acexy/golang-toolkit/logger"
	"github.com/gin-gonic/gin"
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

	defaultRecoverHandlerResponse RecoverHandlerResponse = func(ctx *gin.Context, err any) Response {
		if v, ok := err.(error); ok {
			logger.Logrus().WithError(v).Errorln("Request catch exception path:", ctx.Request.URL, "panic:", err)
		} else {
			logger.Logrus().Errorln("Request catch exception path:", ctx.Request.URL, "panic:", err)
		}
		return RespRestException()
	}

	httpCodeWithStatus map[int]StatusCode
)

type HttpStatusCodeCodeHandlerResponse func(ctx *gin.Context, httpStatusCode int) Response
type RecoverHandlerResponse func(ctx *gin.Context, err any) Response

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
		defer func() {
			var err any
			if r := recover(); r != nil {
				err = r
			} else {
				if ctx.Err() != nil {
					err = ctx.Err()
				}
			}
			if err != nil {
				response := defaultRecoverHandlerResponse(ctx, err)
				if response != nil {
					httpResponse(ctx, response)
					writer := ctx.Writer
					// 如果使用了可覆写中间件
					if v, ok := writer.(*responseRewriter); ok {
						v.ResponseWriter.WriteHeader(v.statusCode)
						_, _ = v.ResponseWriter.Write(v.body.Bytes())
					}
				}
			}
		}()
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
	if !disabledDefaultIgnoreHttpStatusCode {
		for _, v := range defaultIgnoreHttpStatusCode {
			if httpCode == v {
				return true
			}
		}
	}
	if len(ignoreHttpStatusCode) > 0 {
		for _, v := range ignoreHttpStatusCode {
			if httpCode == v {
				return true
			}
		}
	}
	return false
}
