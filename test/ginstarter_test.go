package test

import (
	"fmt"
	"github.com/acexy/golang-toolkit/logger"
	"github.com/acexy/golang-toolkit/sys"
	"github.com/acexy/golang-toolkit/util/json"
	"github.com/gin-gonic/gin"
	"github.com/golang-acexy/starter-gin/ginstarter"
	"github.com/golang-acexy/starter-gin/test/router"
	"github.com/golang-acexy/starter-parent/parent"
	"net/http"
	"testing"
	"time"
)

var starterLoader *parent.StarterLoader

// 默认Gin表现行为
// 启用了非200状态码自动包裹响应
func TestGinDefault(t *testing.T) {
	starterLoader = parent.NewStarterLoader([]parent.Starter{
		&ginstarter.GinStarter{
			ListenAddress: ":8080",
			DebugModule:   true,
			Routers: []ginstarter.Router{
				&router.DemoRouter{},
				&router.ParamRouter{},
				&router.AbortRouter{},
				&router.BasicAuthRouter{},
				&router.MyRestRouter{},
			},
			InitFunc: func(instance *gin.Engine) {
				instance.GET("/ping", func(context *gin.Context) {
					context.String(http.StatusOK, "alive")
				})
				instance.GET("/err", func(context *gin.Context) {
					context.Status(500)
				})
			},
			DisabledDefaultIgnoreHttpStatusCode: true,
		},
	})
	err := starterLoader.Start()
	if err != nil {
		fmt.Printf("%+v\n", err)
		return
	}

	sys.ShutdownHolding()
}

// 自定义Gin的表现 将在和默认行为相同的路由功能代码下表现不同的响应
// 禁用了http异常响应码自动包裹
// 自定义panic异常响应
func TestGinCustomer(t *testing.T) {
	starter := &ginstarter.GinStarter{
		ListenAddress: ":8080",
		DebugModule:   true,
		Routers: []ginstarter.Router{
			&router.DemoRouter{},
			&router.ParamRouter{},
			&router.AbortRouter{},
			&router.BasicAuthRouter{},
			&router.MyRestRouter{},
		},
		InitFunc: func(instance *gin.Engine) {
			instance.GET("/ping", func(context *gin.Context) {
				context.String(http.StatusOK, "alive")
			})
			instance.GET("/err", func(context *gin.Context) {
				context.Status(500)
			})
		},
		DisabledDefaultIgnoreHttpStatusCode: true,
		DisableMethodNotAllowedError:        true,
		RecoverHandlerResponse: func(ctx *gin.Context, err any) ginstarter.Response {
			logger.Logrus().Errorln("Request catch exception", err)
			return ginstarter.RespTextPlain("something error", http.StatusOK)
		},
		DisableHttpStatusCodeHandler: true,
		GlobalMiddlewares: []ginstarter.Middleware{
			func(request *ginstarter.Request) (ginstarter.Response, bool) {
				if request.RequestPath() == "/mdw" {
					return ginstarter.RespTextPlain("middleware", http.StatusOK), false
				}
				return ginstarter.RespTextPlain("hello world", http.StatusOK), true
			},
		},
	}
	loader := parent.NewStarterLoader([]parent.Starter{starter})

	err := loader.Start()
	if err != nil {
		fmt.Printf("%+v\n", err)
		return
	}

	sys.ShutdownHolding()
}

func TestGinLoadAndUnload(t *testing.T) {
	starterLoader = parent.NewStarterLoader([]parent.Starter{
		&ginstarter.GinStarter{
			ListenAddress: ":8080",
			DebugModule:   true},
	})
	err := starterLoader.Start()
	if err != nil {
		fmt.Printf("%+v\n", err)
		return
	}
	time.Sleep(time.Second * 5)
	stopResult, err := starterLoader.StopBySetting()
	if err != nil {
		fmt.Printf("%+v\n", err)
		return
	}
	fmt.Println(json.ToJsonFormat(stopResult))
}
