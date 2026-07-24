package test

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/acexy/golang-toolkit/logger"
	"github.com/acexy/golang-toolkit/sys"
	"github.com/acexy/golang-toolkit/util/json"
	"github.com/gin-gonic/gin"
	"github.com/golang-acexy/starter-gin/ginstarter"
	"github.com/golang-acexy/starter-gin/test/router"
	"github.com/golang-acexy/starter-parent/parent"
	"github.com/sirupsen/logrus"
)

var starterLoader *parent.StarterLoader

func init() {
	logger.SetTraceIdSupplier(&traceID{})
	logger.EnableConsoleWithFormatter(logger.TraceLevel, logger.NewFormatter(func(traceSupplier logger.TraceIdSupplier, entry *logrus.Entry) ([]byte, error) {
		// 格式化时间戳，保留毫秒部分
		timestamp := entry.Time.Format("2006-01-02 15:04:05.000")
		// 格式化日志等级，大写右对齐
		level := strings.ToUpper(entry.Level.String())
		if len(level) > 5 {
			level = level[:5]
		}
		// 获取文件名与行号
		file := "unknown:0"
		if entry.HasCaller() {
			file = fmt.Sprintf("%s:%d", filepath.Base(entry.Caller.File), entry.Caller.Line)
		}
		log := fmt.Sprintf("%s %s %-5s [%s] - %s\n", traceSupplier.GetTraceId(), timestamp, level, file, entry.Message)
		return []byte(log), nil
	}))
	logger.EnableFileWithJson(logger.TraceLevel)
}

type traceID struct {
}

func (t *traceID) SetTraceId(s string) {

}

func (t *traceID) GetTraceId() string {
	return "traceId"
}

// 默认Gin表现行为
// 启用了非200状态码自动包裹响应
func TestGinDefault(t *testing.T) {
	starterLoader = parent.InitStarterLoader([]parent.Starter{
		&ginstarter.GinStarter{
			Config: ginstarter.GinConfig{
				ListenAddress:       ":8080",
				UseReusePortModel:   true,
				DebugModule:         true,
				MaxRequestBodyBytes: 4 << 10,
				Routers: []ginstarter.Router{
					&router.DemoRouter{},
					&router.ParamRouter{},
					&router.AbortRouter{},
					&router.BasicAuthRouter{},
					&router.MyRestRouter{},
					&router.InterceptorTestRouter{},
				},
				GlobalPreInterceptors:  router.InterceptorTestGlobalPreInterceptors(),
				GlobalPostInterceptors: router.InterceptorTestGlobalPostInterceptors(),
				InitFunc: func(instance *gin.Engine) {
					instance.GET("/ping", func(context *gin.Context) {
						context.String(http.StatusOK, "alive")
					})
					instance.GET("/err", func(context *gin.Context) {
						context.Status(500)
					})
				},
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
		Config: ginstarter.GinConfig{
			ListenAddress:       ":8080",
			UseReusePortModel:   true,
			DebugModule:         true,
			MaxRequestBodyBytes: 4 << 10,
			Routers: []ginstarter.Router{
				&router.DemoRouter{},
				&router.ParamRouter{},
				&router.AbortRouter{},
				&router.BasicAuthRouter{},
				&router.MyRestRouter{},
				&router.InterceptorTestRouter{},
			},
			GlobalPreInterceptors:  router.InterceptorTestGlobalPreInterceptors(),
			GlobalPostInterceptors: router.InterceptorTestGlobalPostInterceptors(),
			HidePanicErrorDetails: false,
			InitFunc: func(instance *gin.Engine) {
				instance.GET("/ping", func(context *gin.Context) {
					context.String(http.StatusOK, "alive")
				})
				instance.GET("/err", func(context *gin.Context) {
					context.Status(500)
				})
			},
			DisableBadHTTPCodeResolver:   true,
			DisableDefaultIgnoreHTTPCode: true,
			DisableMethodNotAllowedError: false,
		},
	}
	loader := parent.InitStarterLoader([]parent.Starter{starter})

	err := loader.Start()
	if err != nil {
		fmt.Printf("%+v\n", err)
		return
	}

	sys.ShutdownHolding()
}

func TestGinLoadAndUnload(t *testing.T) {
	starterLoader = parent.InitStarterLoader([]parent.Starter{
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
	stopResult, err := starterLoader.StopAllBySetting()
	if err != nil {
		fmt.Printf("%+v\n", err)
		return
	}
	fmt.Println(json.ToStringFormat(stopResult))
}
