package ginstarter

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/acexy/golang-toolkit/logger"
	"github.com/acexy/golang-toolkit/math/conversion"
	"github.com/acexy/golang-toolkit/util/coll"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
	"strings"
	"time"
)

var (
	httpCodeWithStatus          map[int]StatusCode
	defaultIgnoreHttpStatusCode = []int{
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

	badHttpCodeResolver BadHttpCodeResolver = func(httpStatusCode int, errMsg string) Response {

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

		return NewRespRest().DataBuilder(func() *ResponseData {
			bodyBytes, _ := ginConfig.ResponseDataStructDecoder.Decode(body)
			return NewResponseDataWithStatusCode(gin.MIMEJSON, bodyBytes, http.StatusOK)
		})
	}
)

type PanicResolver func(err error) string
type BadHttpCodeResolver func(httpStatusCode int, errMsg string) Response

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

func isIgnoreHttpStatusCode(httpCode int) bool {
	if !ginConfig.DisableDefaultIgnoreHttpCode {
		for _, v := range defaultIgnoreHttpStatusCode {
			if httpCode == v {
				return true
			}
		}
	}
	if len(ginConfig.IgnoreHttpCode) > 0 {
		for _, v := range ginConfig.IgnoreHttpCode {
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
				err = errors.New(jsonErr.Field + " type mismatch")
			} else if _, ok := rawError.(*json.SyntaxError); ok {
				err = errors.New("bad json payload")
			} else {
				err = rawError
			}
		} else {
			err = fmt.Errorf("%v", t)
		}
	}
	logger.Logrus().Errorf("panic: %v", err)
	return
}

// recoverHandler 全局Panic处理中间件
func recoverHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// panic异常处理
		defer func() {
			if panicError := recover(); panicError != nil {

				var errMsg string
				// 将panic异常进行转换
				status, err, internalError := panicToError(panicError)
				if ginConfig.HidePanicErrorDetails { // 禁用异常信息显示
					if !internalError {
						errMsg = ""
						status = 500
					} else {
						errMsg = err.Error()
					}
				} else {
					errMsg = ginConfig.PanicResolver(err)
				}

				if status != 0 {
					ctx.Status(status)
				}

				writer := ctx.Writer
				var statusCode int
				var rewriter *responseRewriter
				// 如果使用了可覆写中间件
				if w, ok := writer.(*responseRewriter); ok {
					rewriter = w
					statusCode = w.statusCode
				} else {
					statusCode = ctx.Writer.Status()
				}

				var response Response
				if !ginConfig.DisableBadHttpCodeResolver {
					response = ginConfig.BadHttpCodeResolver(statusCode, errMsg)
				} else {
					response = RespTextPlain(errMsg, statusCode)
				}

				httpResponse(ctx, response)
				if rewriter != nil {
					rewriter.ResponseWriter.WriteHeader(rewriter.statusCode)
					_, _ = rewriter.ResponseWriter.Write(rewriter.body.Bytes())
				}
			}
		}()

		ctx.Next()
		if !ctx.IsAborted() {
			// 异常响应码处理
			if !ginConfig.DisableBadHttpCodeResolver {
				var statusCode int
				var rewriter *responseRewriter
				if v, ok := ctx.Writer.(*responseRewriter); ok {
					rewriter = v
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
					response := ginConfig.BadHttpCodeResolver(statusCode, "")
					httpResponse(ctx, response)
					if rewriter != nil {
						rewriter.ResponseWriter.WriteHeader(rewriter.statusCode)
						_, _ = rewriter.ResponseWriter.Write(rewriter.body.Bytes())
					}
				}
			}
		}
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

// MediaTypeMiddleware 类型校验中间件
func MediaTypeMiddleware(contentType []string, match ...func(request *Request) bool) Middleware {
	return func(request *Request) (Response, bool) {
		if len(match) > 0 {
			if !match[0](request) {
				return nil, true
			}
		}
		if len(contentType) > 0 {
			if !isMatchMediaType(contentType, request.GetHeader("Content-Type")) {
				return RespAbortWithHttpStatusCode(http.StatusUnsupportedMediaType), false
			}
		} else {
			logger.Logrus().Warningln("valid Content-Type restriction not set")
		}
		return nil, true
	}
}

func isMatchMediaType(allowContentType []string, requestContentType string) bool {
	return coll.SliceContains(allowContentType, strings.TrimSpace(strings.Split(requestContentType, ";")[0]), func(s1 string, s2 string) bool {
		return strings.ToLower(s1) == strings.ToLower(s2)
	})
}
