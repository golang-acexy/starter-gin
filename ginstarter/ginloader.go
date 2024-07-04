package ginstarter

import (
	"context"
	"github.com/acexy/golang-toolkit/logger"
	"github.com/acexy/golang-toolkit/util/net"
	"github.com/gin-gonic/gin"
	"github.com/golang-acexy/starter-parent/parent"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

var server *http.Server
var ginEngine *gin.Engine

var (
	disabledDefaultIgnoreHttpStatusCode bool
	ignoreHttpStatusCode                []int
)

type GinStarter struct {

	// 自定义Gin模块的组件属性
	GinSetting *parent.Setting

	// 模块组件在启动时执行初始化
	InitFunc func(instance *gin.Engine)

	// * 注册业务路由
	Routers []Router

	// * 注册服务监听地址 :8080 (默认)
	ListenAddress string // ip:port

	// 自定义异常响应处理 如果不指定则使用默认方式
	RecoverHandlerResponse RecoverHandlerResponse

	// 禁用错误包装处理器 在出现非200响应码或者异常时，将自动进行转化
	DisableHttpStatusCodeHandler bool
	// 在启用非200响应码自动处理后，指定忽略需要自动包裹响应码
	IgnoreHttpStatusCode []int
	// 关闭系统内置的忽略的http状态码
	DisabledDefaultIgnoreHttpStatusCode bool
	// 在出现非200响应码或者异常时具体响应策略 如果不指定则使用默认处理器 仅在UseHttpStatusCodeHandler = true 生效
	HttpStatusCodeCodeHandlerResponse HttpStatusCodeCodeHandlerResponse

	// 响应数据的结构体解码器 默认为JSON方式解码
	// 在使用NewRespRest响应结构体数据时解码为[]byte数据的解码器
	// 如果自实现Response接口将不使用解码器
	ResponseDataStructDecoder ResponseDataStructDecoder

	// gin config
	DebugModule        bool
	MaxMultipartMemory int64

	// 关闭包裹405错误展示，使用404代替
	DisableMethodNotAllowedError bool

	// 禁用尝试获取真实IP
	DisableForwardedByClientIP bool
}

func (g *GinStarter) Setting() *parent.Setting {
	if g.GinSetting != nil {
		return g.GinSetting
	}
	return parent.NewSetting(
		"Gin-Starter",
		0,
		true,
		time.Second*30,
		func(instance interface{}) {
			if g.InitFunc != nil {
				g.InitFunc(instance.(*gin.Engine))
			}
		})
}

func (g *GinStarter) Start() (interface{}, error) {
	var err error
	if g.DebugModule {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	gin.DefaultWriter = &logrusLogger{log: logger.Logrus(), level: logrus.DebugLevel}
	gin.DefaultErrorWriter = &logrusLogger{log: logger.Logrus(), level: logrus.ErrorLevel}
	ginEngine = gin.New()

	ginEngine.Use(recoverHandler())
	if g.RecoverHandlerResponse != nil {
		defaultRecoverHandlerResponse = g.RecoverHandlerResponse
	}

	if g.MaxMultipartMemory > 0 {
		ginEngine.MaxMultipartMemory = g.MaxMultipartMemory
	}

	ginEngine.ForwardedByClientIP = !g.DisableForwardedByClientIP

	if !g.DisableMethodNotAllowedError {
		ginEngine.HandleMethodNotAllowed = true
	}

	if !g.DisableHttpStatusCodeHandler {
		ginEngine.Use(responseRewriteHandler())
		ginEngine.Use(httpStatusCodeHandler())
		disabledDefaultIgnoreHttpStatusCode = g.DisabledDefaultIgnoreHttpStatusCode
		ignoreHttpStatusCode = g.IgnoreHttpStatusCode
		if g.HttpStatusCodeCodeHandlerResponse != nil {
			defaultHttpStatusCodeHandlerResponse = g.HttpStatusCodeCodeHandlerResponse
		}
	}

	if g.ResponseDataStructDecoder != nil {
		defaultResponseDataDecoder = g.ResponseDataStructDecoder
	}

	if len(g.Routers) > 0 {
		registerRouter(ginEngine, g.Routers)
	}

	if g.ListenAddress == "" {
		g.ListenAddress = ":8080"
	}

	server = &http.Server{
		Addr:    g.ListenAddress,
		Handler: ginEngine,
	}

	errChn := make(chan error)
	go func() {
		if err = server.ListenAndServe(); err != nil {
			errChn <- err
		}
	}()

	select {
	case <-time.After(time.Second):
		return ginEngine, nil
	case err = <-errChn:
		return ginEngine, err
	}
}

func (g *GinStarter) Stop(maxWaitTime time.Duration) (gracefully, stopped bool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), maxWaitTime)
	defer cancel()
	if err = server.Shutdown(ctx); err != nil {
		gracefully = false
	} else {
		gracefully = true
	}
	stopped = !net.Telnet(g.ListenAddress, time.Second)
	return
}

// RawGinEngine 获取原始的gin引擎实例
func RawGinEngine() *gin.Engine {
	return ginEngine
}
