package ginmodule

import (
	"context"
	"github.com/acexy/golang-toolkit/logger"
	"github.com/gin-gonic/gin"
	"github.com/golang-acexy/starter-parent/parentmodule/declaration"
	"net/http"
	"time"
)

var server *http.Server

const (
	defaultListenAddress = ":8080"
)

type GinModule struct {

	// 自定义Module配置
	GinModuleConfig *declaration.ModuleConfig
	GinInterceptor  func(instance interface{})

	ginEngine *gin.Engine

	// * 注册业务路由
	Routers []Router

	// * 注册服务监听地址 :8080 (默认)
	ListenAddress string // ip:port

	UseErrorCodeHandler bool // 使用错误包装处理器 在出现非200响应码或者异常时，将自动进行转化

	// gin config
	DebugModule                  bool
	MaxMultipartMemory           int64
	DisableMethodNotAllowedError bool
	ForwardedByClientIP          bool
}

func (g *GinModule) ModuleConfig() *declaration.ModuleConfig {
	if g.GinModuleConfig != nil {
		return g.GinModuleConfig
	}
	return &declaration.ModuleConfig{
		ModuleName:               "Gin",
		UnregisterPriority:       0,
		UnregisterAllowAsync:     true,
		UnregisterMaxWaitSeconds: 30,
		LoadInterceptor:          g.GinInterceptor,
	}
}

func (g *GinModule) RawInstance() interface{} {
	g.ginEngine = gin.New()
	return g.ginEngine
}

func (g *GinModule) Register() error {
	var err error
	ginEngin := g.ginEngine
	if g.DebugModule {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	if g.MaxMultipartMemory > 0 {
		ginEngin.MaxMultipartMemory = g.MaxMultipartMemory
	}

	ginEngin.ForwardedByClientIP = g.ForwardedByClientIP
	if !g.DisableMethodNotAllowedError {
		ginEngin.HandleMethodNotAllowed = true
	}

	if g.UseErrorCodeHandler {
		ginEngin.Use(ErrorCodeHandler())
	}

	ginEngin.Use(Recover())

	if len(g.Routers) > 0 {
		loadRouter(ginEngin, g.Routers)
	}

	if g.ListenAddress == "" {
		g.ListenAddress = defaultListenAddress
	}

	server = &http.Server{
		Addr:    g.ListenAddress,
		Handler: ginEngin,
	}

	go func() {
		logger.Logrus().Traceln(g.ModuleConfig().ModuleName, "started")
		if err = server.ListenAndServe(); err != nil {
		}
	}()

	return err
}

func (g *GinModule) Unregister(maxWaitSeconds uint) (gracefully bool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(maxWaitSeconds)*time.Second)
	defer cancel()
	if err = server.Shutdown(ctx); err != nil {
		gracefully = false
	} else {
		gracefully = true
	}
	return
}
