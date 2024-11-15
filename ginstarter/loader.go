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
var ginStarter *GinStarter

type GinStarter struct {

	// 自定义Gin模块的组件属性
	GinSetting *parent.Setting

	// 模块组件在启动时执行初始化
	InitFunc func(instance *gin.Engine)

	// * 注册业务路由
	Routers []Router

	// * 注册服务监听地址 :8080 (默认)
	ListenAddress string // ip:port

	// 默认情况系统会将捕获的异常详细发给PanicResolver处理，如果不想将细节暴露向外
	// 方案 1. 启用隐藏异常细节功能，系统将在触发panic重要错误时不再调用PanicResolver处理，并统一响应500错误
	// 方案 2. 如果不想禁用异常时调度PanicResolver, 可以在初始化时手动设置自定义PanicResolver处理器
	// * panic 将被分为框架内部错误和框架未知错误 框架内部错误是非敏感错误，不受该参数控制，每次触发PanicResolver，例如验证框架错误
	HidePanicErrorDetails bool
	// 全局异常响应处理器 如果不指定则使用默认方式
	PanicResolver PanicResolver

	// 禁用异常http响应码Resolver
	DisableBadHttpCodeResolver bool
	// 启用异常http响应码Resolver 系统已内置常见的不处理的非正常响应码 可以禁用
	DisableDefaultIgnoreHttpCode bool
	// 启用异常http响应码Resolver 指定不处理特定的异常响应码
	IgnoreHttpCode []int
	// 启用异常http响应码Resolver 如果不指定则使用默认方式
	BadHttpCodeResolver BadHttpCodeResolver

	// 自定义全局中间件 作用于所有请求 按照顺序执行
	GlobalMiddlewares []Middleware

	// 响应数据的结构体解码器 默认为JSON方式解码
	// 在使用NewRespRest响应结构体数据时解码为[]byte数据的解码器
	// 如果自实现Response接口将不使用解码器
	ResponseDataStructDecoder ResponseDataStructDecoder

	// ========== gin config
	DebugModule        bool
	MaxMultipartMemory int64

	// 关闭包裹405错误展示，使用404代替
	DisableMethodNotAllowedError bool

	// 禁用尝试获取转发真实IP
	DisableForwardedByClientIP bool
}

func (g *GinStarter) Setting() *parent.Setting {
	if g.GinSetting != nil {
		return g.GinSetting
	}
	return parent.NewSetting(
		"Gin-Starter",
		0,
		false,
		time.Second*30,
		func(instance interface{}) {
			if g.InitFunc != nil {
				g.InitFunc(instance.(*gin.Engine))
			}
		})
}

func (g *GinStarter) Start() (interface{}, error) {
	ginStarter = g

	var err error
	if g.DebugModule {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	gin.DefaultWriter = &logrusLogger{log: logger.Logrus(), level: logrus.DebugLevel}
	gin.DefaultErrorWriter = &logrusLogger{log: logger.Logrus(), level: logrus.ErrorLevel}
	ginEngine = gin.New()

	registerValidators()

	ginEngine.Use(recoverHandler())

	if g.PanicResolver == nil {
		g.PanicResolver = panicResolver
	}

	if g.MaxMultipartMemory > 0 {
		ginEngine.MaxMultipartMemory = g.MaxMultipartMemory
	}

	ginEngine.ForwardedByClientIP = !g.DisableForwardedByClientIP

	if !g.DisableMethodNotAllowedError {
		ginEngine.HandleMethodNotAllowed = true
	}

	if !g.DisableBadHttpCodeResolver {
		ginEngine.Use(responseRewriteHandler())
		if g.BadHttpCodeResolver == nil {
			g.BadHttpCodeResolver = badHttpCodeResolver
		}
	}

	if g.ResponseDataStructDecoder == nil {
		g.ResponseDataStructDecoder = responseJsonDataStructDecoder{}
	}

	if len(g.GlobalMiddlewares) > 0 {
		for i := range g.GlobalMiddlewares {
			middleware := g.GlobalMiddlewares[i]
			if middleware != nil {
				ginEngine.Use(func(ctx *gin.Context) {
					response, continued := middleware(&Request{ctx: ctx})
					if !continued {
						httpResponse(ctx, response)
						ctx.Abort()
					} else {
						ctx.Next()
					}
				})
			}
		}
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
