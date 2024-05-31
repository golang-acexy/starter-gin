package ginmodule

import (
	"github.com/acexy/golang-toolkit/logger"
	"github.com/gin-gonic/gin"
	"net/http"
)

var httpCodeWithStatus map[int]StatusCode

func init() {
	httpCodeWithStatus = make(map[int]StatusCode, 6)
	httpCodeWithStatus[http.StatusBadRequest] = StatusCodeBadRequestParameters
	httpCodeWithStatus[http.StatusForbidden] = StatusCodeForbidden
	httpCodeWithStatus[http.StatusNotFound] = StatusCodeNotFound
	httpCodeWithStatus[http.StatusMethodNotAllowed] = StatusCodeMethodNotAllowed
	httpCodeWithStatus[http.StatusUnsupportedMediaType] = StatusCodeMediaTypeNotAllowed
	httpCodeWithStatus[http.StatusRequestEntityTooLarge] = StatusCodeUploadLimitExceeded
	httpCodeWithStatus[http.StatusUnauthorized] = StatusCodeForbidden
}

func ErrorCodeHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
		statusCode := ctx.Writer.Status()
		if statusCode != http.StatusOK {
			ctx.Status(200)
			logger.Logrus().Warnln("not success response statusCode =", statusCode)
			v, ok := httpCodeWithStatus[statusCode]
			if !ok {
				ctx.AbortWithStatusJSON(http.StatusOK, ResponseException())
			} else {
				ctx.AbortWithStatusJSON(http.StatusOK, ResponseError(v))
			}
			ctx.Abort()
		}
	}
}

func Recover() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Logrus().WithField("error", r).Error(r)
			}
		}()
		ctx.Next()
	}
}
