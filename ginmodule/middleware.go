package ginmodule

import (
	"github.com/acexy/golang-toolkit/log"
	"github.com/gin-gonic/gin"
	"net/http"
)

var httpCodeWithStatus map[int]StatusCode

func init() {
	httpCodeWithStatus = make(map[int]StatusCode, 6)
	httpCodeWithStatus[http.StatusBadRequest] = statusCodeBadRequestParameters
	httpCodeWithStatus[http.StatusForbidden] = statusCodeForbidden
	httpCodeWithStatus[http.StatusNotFound] = statusCodeNotFound
	httpCodeWithStatus[http.StatusMethodNotAllowed] = statusCodeMethodNotAllowed
	httpCodeWithStatus[http.StatusUnsupportedMediaType] = statusCodeMediaTypeNotAllowed
	httpCodeWithStatus[http.StatusRequestEntityTooLarge] = statusCodeUploadLimitExceeded
}

func BasicRecover() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			statusCode := ctx.Writer.Status()
			if r := recover(); r != nil {
				log.Logrus().WithField("error", r).Errorln("catch exception")
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
