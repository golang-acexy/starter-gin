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
	sys.EnableLocalTraceId(nil)
	starterLoader = parent.NewStarterLoader([]parent.Starter{
		&ginstarter.GinStarter{
			Config: ginstarter.GinConfig{ListenAddress: ":8080",
				DebugModule: true,
				Routers: []ginstarter.Router{
					&router.DemoRouter{},
					&router.ParamRouter{},
					&router.AbortRouter{},
					&router.BasicAuthRouter{},
					&router.MyRestRouter{},
				},
				EnableGoroutineTraceIdResponse: true,
				InitFunc: func(instance *gin.Engine) {
					instance.GET("/ping", func(context *gin.Context) {
						context.String(http.StatusOK, "alive")
					})
					instance.GET("/err", func(context *gin.Context) {
						context.Status(500)
					})
				},
				//DisableDefaultIgnoreHttpCode: true,
			},
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
		Config: ginstarter.GinConfig{ListenAddress: ":8080",
			DebugModule: true,
			Routers: []ginstarter.Router{
				&router.DemoRouter{},
				&router.ParamRouter{},
				&router.AbortRouter{},
				&router.BasicAuthRouter{},
				&router.MyRestRouter{},
			},
			HidePanicErrorDetails: false,
			InitFunc: func(instance *gin.Engine) {
				instance.GET("/ping", func(context *gin.Context) {
					context.String(http.StatusOK, "alive")
				})
				instance.GET("/err", func(context *gin.Context) {
					context.Status(500)
				})
			},
			//DisableBadHttpCodeResolver: true,
			//DisableDefaultIgnoreHttpCode: true,
			DisableMethodNotAllowedError: false,
			//PanicResolver: func(ctx *gin.Context, err error) ginstarter.Response {
			//	logger.Logrus().Errorln("Request catch exception", err)
			//	return ginstarter.RespTextPlain("something error", http.StatusOK)
			//},
			GlobalPreInterceptors: []ginstarter.PreInterceptor{
				func(request *ginstarter.Request) (ginstarter.Response, bool) {
					t, _ := request.GetQueryParam("t")
					if t == "" {
						logger.Logrus().Infoln("前置 不继续执行 忽略其他中间件")
						return ginstarter.RespTextPlain("interceptor", http.StatusOK), false
					} else {
						logger.Logrus().Infoln("前置 继续执行 不忽略其他中间件")
						return nil, true
					}
				},
				func(request *ginstarter.Request) (ginstarter.Response, bool) {
					logger.Logrus().Infoln("前置 interceptor 2 执行")
					return nil, true
				},
			},
			GlobalPostInterceptors: []ginstarter.PostInterceptor{
				func(request *ginstarter.Request, response ginstarter.Response) bool {
					t, _ := request.GetQueryParam("t")
					if t == "" {
						logger.Logrus().Infoln("后置 不继续执行 忽略其他中间件")
						return false
					} else {
						logger.Logrus().Infoln("后置 继续执行 不忽略其他中间件")
						return false
					}
				},
				func(request *ginstarter.Request, response ginstarter.Response) bool {
					logger.Logrus().Infoln("后置 interceptor 2 执行")
					return true
				},
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
			Config: ginstarter.GinConfig{
				ListenAddress: ":8080",
				DebugModule:   true,
			},
		},
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
