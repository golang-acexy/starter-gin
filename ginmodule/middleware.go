package ginmodule

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"runtime/debug"
)

func BasicRecover() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				debug.PrintStack()
				ctx.AbortWithStatusJSON(http.StatusOK, NewException())
			} else {
				statusCode := ctx.Writer.Status()
				if statusCode != 200 {
					switch statusCode {
					case 404:
						ctx.AbortWithStatusJSON(http.StatusOK, NewError(statusCodeNotFound))
					case 403:
						ctx.AbortWithStatusJSON(http.StatusOK, NewError(statusCodeForbidden))
					case 405:
						ctx.AbortWithStatusJSON(http.StatusOK, NewError(statusCodeMethodNotAllowed))
					default:
						ctx.AbortWithStatusJSON(http.StatusOK, NewException())
					}
				}
			}
		}()
		ctx.Next()
	}
}
