package ginmodule

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"runtime/debug"
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
			if r := recover(); r != nil {
				debug.PrintStack()
				ctx.AbortWithStatusJSON(http.StatusOK, NewException())
			} else {
				statusCode := ctx.Writer.Status()
				if statusCode != http.StatusOK {
					v, ok := httpCodeWithStatus[statusCode]
					if !ok {
						ctx.AbortWithStatusJSON(http.StatusOK, NewException())
					} else {
						ctx.AbortWithStatusJSON(http.StatusOK, v)
					}
				}
			}
		}()
		ctx.Next()
	}
}
