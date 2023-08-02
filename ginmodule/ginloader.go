package ginmodule

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/golang-acexy/starter-parent/parentmodule/declaration"
	"net/http"
	"time"
)

type GinModule struct {
	server *http.Server

	// 自定义Module配置
	GinModuleConfig *declaration.ModuleConfig

	// * 注册业务路由
	Routers []Router

	// * 注册服务监听地址 :8080 (默认)
	ListenAddress string // ip:port

	// gin config
	DebugModule            bool
	MaxMultipartMemory     int64
	HandleMethodNotAllowed bool
	ForwardedByClientIP    bool
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
	}
}

func (g *GinModule) Interceptor() *func(instance interface{}) {
	return nil
}

func (g *GinModule) Register(interceptor *func(instance interface{})) error {
	var err error

	if g.DebugModule {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	ginEngin := gin.New()

	if interceptor != nil {
		(*interceptor)(ginEngin)
	}

	if g.MaxMultipartMemory > 0 {
		ginEngin.MaxMultipartMemory = g.MaxMultipartMemory
	}

	ginEngin.ForwardedByClientIP = g.ForwardedByClientIP
	ginEngin.HandleMethodNotAllowed = g.HandleMethodNotAllowed

	ginEngin.Use(BasicRecover())

	if len(g.Routers) > 0 {
		loadRouter(ginEngin, g.Routers)
	}

	if g.ListenAddress == "" {
		g.ListenAddress = ":8080"
	}

	g.server = &http.Server{
		Addr:    g.ListenAddress,
		Handler: ginEngin,
	}

	go func() {
		if err = g.server.ListenAndServe(); err != nil {
			return
		}
	}()

	return err
}

func (g *GinModule) Unregister(maxWaitSeconds uint) (gracefully bool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(maxWaitSeconds)*time.Second)
	defer cancel()
	if err = g.server.Shutdown(ctx); err != nil {
		gracefully = false
	} else {
		gracefully = true
	}
	return
}
