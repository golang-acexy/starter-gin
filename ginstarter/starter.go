package ginstarter

import (
	"context"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/acexy/golang-toolkit/util/coll"
	"github.com/gin-gonic/gin"
	"github.com/golang-acexy/starter-parent/parent"
	"github.com/libp2p/go-reuseport"
	"github.com/sirupsen/logrus"
)

var ginRuntimeState atomic.Pointer[ginRuntime]
var ginLifecycleLock sync.Mutex

type ginRuntime struct {
	server *http.Server
	engine *gin.Engine
	config *GinConfig
	stopping atomic.Bool
}

type GinConfig struct {

	// 模块组件在启动时执行初始化
	InitFunc func(instance *gin.Engine)

	// * 注册业务路由
	Routers []Router

	// * 注册服务监听地址 :8080 (默认)
	ListenAddress string // ip:port

	// 启用ReusePortModel 使用 SO_REUSEPORT 实现 多个进程监听同一端口，基于操作系统内核实现负载均衡
	// 该功能受操作系统影响 win可能不支持， mac 只支持端口复用不能负载 linux支持负载加复用
	UseReusePortModel bool

	// 默认情况系统会将捕获的异常详细发给PanicResolver处理，如果不想将细节暴露向外
	// 方案 1. 启用隐藏异常细节功能，系统不会向客户端暴露未知异常信息，但会保留明确指定的HTTP状态码
	// 方案 2. 如果不想禁用异常时调用PanicResolver, 可以在初始化时手动设置自定义PanicResolver处理器
	// * panic 将被分为框架内部错误和框架未知错误 框架内部错误是非敏感错误，不受该参数控制，每次都会触发PanicResolver，例如验证框架错误
	HidePanicErrorDetails bool
	// 全局异常响应处理器 如果不指定则使用默认方式
	PanicResolver PanicResolver

	// 禁用异常http响应码Resolver
	DisableBadHTTPCodeResolver bool
	// 禁用系统内置的忽略异常响应码
	DisableDefaultIgnoreHTTPCode bool
	// 启用异常http响应码Resolver 指定不处理特定的异常响应码
	IgnoreHTTPCode []int
	// 启用异常http响应码Resolver 如果不指定则使用默认方式
	BadHTTPCodeResolver BadHTTPCodeResolver

	// GlobalPreInterceptors 全局前置拦截器。
	// 在路由前置拦截器和业务 Handler 之前，按照切片注册顺序执行。
	// 任一拦截器返回非空 Response 或 error 时，将终止后续前置拦截器和 Handler；
	// 此时不会进入路由拦截器链，但仍会执行全局后置拦截器，并由框架统一写入最终 Response。
	GlobalPreInterceptors []PreInterceptor

	// GlobalPostInterceptors 全局后置拦截器。
	// 在路由后置拦截器执行完成后，按照切片注册顺序执行。
	// 每个拦截器接收当前 Response：返回非空 Response 时替换当前响应，返回 nil 时保留当前响应；
	// 返回 error 时将其转换为异常 Response，并终止剩余的全局后置拦截器。
	// 全部执行完成后，由框架统一写入最终 Response。
	GlobalPostInterceptors []PostInterceptor

	// ResponseBodyEncoder 响应正文编码器，默认使用 JSON 编码。
	// NewRespRest 使用该编码器将结构化正文编码为字节数据。
	// 如果自实现 Response 接口并直接提供字节正文，则不会使用该编码器。
	ResponseBodyEncoder ResponseBodyEncoder

	// 启用TraceId响应
	TraceIDResponse func() string

	// ========== gin config
	DebugModule bool
	// MaxMultipartMemory multipart 表单解析时允许保存在内存中的最大字节数，超出部分写入临时文件。
	MaxMultipartMemory int64
	// MaxRequestBodyBytes 请求正文最大字节数，零值表示不限制。
	MaxRequestBodyBytes int64

	// 关闭包裹405错误展示，使用404代替
	DisableMethodNotAllowedError bool

	// 禁用尝试获取转发真实IP
	DisableForwardedByClientIP bool
}

type GinStarter struct {
	// GinConfig 配置
	Config GinConfig
	// 懒加载函数，用于在实际执行时动态获取配置 该权重高于Config的直接配置
	LazyConfig func() GinConfig
	config     *GinConfig
	configOnce sync.Once
	// 自定义Gin模块的组件属性
	GinSetting *parent.Setting
}

// 获取配置信息
func (g *GinStarter) getConfig() *GinConfig {
	g.configOnce.Do(func() {
		if g.LazyConfig != nil {
			config := g.LazyConfig()
			g.config = &config
		} else {
			config := g.Config
			g.config = &config
		}
	})
	return g.config
}

func (g *GinStarter) Setting() *parent.Setting {
	if g.GinSetting != nil {
		return g.GinSetting
	}
	config := g.getConfig()
	return parent.NewSetting(
		"Gin-Starter",
		false,
		1,
		false,
		time.Second*30,
		func(instance any) {
			if config.InitFunc != nil {
				config.InitFunc(instance.(*gin.Engine))
			}
		})
}

func (g *GinStarter) Start() (any, error) {
	var err error
	ginLifecycleLock.Lock()
	if runtime := ginRuntimeState.Load(); runtime != nil {
		ginLifecycleLock.Unlock()
		return runtime.engine, ErrGinStarterAlreadyStarted
	}
	defer ginLifecycleLock.Unlock()
	config := g.getConfig()
	if config.DebugModule {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	gin.DefaultWriter = &logrusLogger{level: logrus.DebugLevel}
	gin.DefaultErrorWriter = &logrusLogger{level: logrus.ErrorLevel}

	ginEngine := gin.New()
	registerValidators()
	if config.PanicResolver == nil {
		config.PanicResolver = panicResolver
	}

	if config.MaxMultipartMemory > 0 {
		ginEngine.MaxMultipartMemory = config.MaxMultipartMemory
	}

	ginEngine.ForwardedByClientIP = !config.DisableForwardedByClientIP

	if !config.DisableMethodNotAllowedError {
		ginEngine.HandleMethodNotAllowed = true
	}

	if config.BadHTTPCodeResolver == nil {
		config.BadHTTPCodeResolver = badHTTPCodeResolver
	}

	if config.ResponseBodyEncoder == nil {
		config.ResponseBodyEncoder = jsonResponseBodyEncoder{}
	}
	config.GlobalPreInterceptors = coll.SliceFilter(config.GlobalPreInterceptors, func(p PreInterceptor) bool {
		return p != nil
	})
	config.GlobalPostInterceptors = coll.SliceFilter(config.GlobalPostInterceptors, func(p PostInterceptor) bool {
		return p != nil
	})
	// 响应管道必须位于最外层，负责一次性提交最终 Response。
	ginEngine.Use(responsePipelineHandler())
	// 全局前置与后置拦截器使用同一中间件，确保前置短路时仍会执行对应后置拦截器。
	ginEngine.Use(func(ctx *gin.Context) {
		defer recoverResponse(ctx)
		state := getRequestState(ctx)
		runPreInterceptors(ctx, config.GlobalPreInterceptors)
		if !state.stopped {
			nextWithRecovery(ctx)
		}
		normalizeResponse(ctx)
		runPostInterceptors(ctx, config.GlobalPostInterceptors)
	})

	ginEngine.NoRoute(func(ctx *gin.Context) {
		setResponse(ctx, RespHTTPStatusCode(http.StatusNotFound))
	})
	ginEngine.NoMethod(func(ctx *gin.Context) {
		setResponse(ctx, RespHTTPStatusCode(http.StatusMethodNotAllowed))
	})

	config.Routers = coll.SliceFilter(config.Routers, func(r Router) bool {
		return r != nil
	})
	if len(config.Routers) > 0 {
		if err = registerRouter(ginEngine, config.Routers); err != nil {
			return nil, err
		}
	}

	if config.ListenAddress == "" {
		config.ListenAddress = ":8080"
	}

	var listener net.Listener
	if config.UseReusePortModel {
		listener, err = reuseport.Listen("tcp", config.ListenAddress)
	} else {
		listener, err = net.Listen("tcp", config.ListenAddress)
	}
	if err != nil {
		return nil, err
	}

	currentServer := &http.Server{
		Addr:    config.ListenAddress,
		Handler: ginEngine,
	}
	runtime := &ginRuntime{server: currentServer, engine: ginEngine, config: config}
	ginRuntimeState.Store(runtime)

	go func() {
		if serveErr := currentServer.Serve(listener); serveErr != nil && serveErr != http.ErrServerClosed {
			logrus.Errorln("gin server stopped unexpectedly:", serveErr)
			ginRuntimeState.CompareAndSwap(runtime, nil)
		}
	}()
	return ginEngine, nil
}

func (g *GinStarter) Stop(maxWaitTime time.Duration) (gracefully, stopped bool, err error) {
	ginLifecycleLock.Lock()
	runtime := ginRuntimeState.Load()
	if runtime == nil {
		ginLifecycleLock.Unlock()
		return false, true, ErrGinServerNotStarted
	}
	currentServer := runtime.server
	ginLifecycleLock.Unlock()
	if !runtime.stopping.CompareAndSwap(false, true) {
		return false, true, ErrGinServerNotStarted
	}
	ctx, cancel := context.WithTimeout(context.Background(), maxWaitTime)
	defer cancel()
	if err = currentServer.Shutdown(ctx); err != nil {
		runtime.stopping.Store(false)
		gracefully = false
		stopped = false
	} else {
		gracefully = true
		stopped = true
		ginLifecycleLock.Lock()
		ginRuntimeState.CompareAndSwap(runtime, nil)
		ginLifecycleLock.Unlock()
	}
	return
}

// RawGinEngine 获取原始的gin引擎实例
func RawGinEngine() *gin.Engine {
	runtime := ginRuntimeState.Load()
	if runtime == nil {
		return nil
	}
	return runtime.engine
}

func currentGinConfig() *GinConfig {
	runtime := ginRuntimeState.Load()
	if runtime == nil {
		return nil
	}
	return runtime.config
}
