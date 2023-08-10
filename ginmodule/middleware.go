package ginmodule

import (
	"github.com/acexy/golang-toolkit/log"
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
}

func BasicRecover() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			statusCode := ctx.Writer.Status()
			if r := recover(); r != nil {
				log.Logrus().WithField("error", r).Error(r)
				if http.StatusOK == statusCode {
					ctx.AbortWithStatusJSON(http.StatusOK, ResponseException())
					return
				}
			}
			if statusCode != http.StatusOK {
				log.Logrus().Warnln("not success response statusCode =", statusCode)
				v, ok := httpCodeWithStatus[statusCode]
				if !ok {
					ctx.AbortWithStatusJSON(http.StatusOK, ResponseException())
				} else {
					ctx.AbortWithStatusJSON(http.StatusOK, ResponseError(v))
				}
			}
		}()
		ctx.Next()
	}
}
