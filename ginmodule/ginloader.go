package ginmodule

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-acexy/starter-parent/parentmodule/declaration"
	"net/http"
	"time"
)

type GinModule struct {
	server *http.Server

	Routers []Router
	Address string // ip:port

	DebugModule            bool
	MaxMultipartMemory     int64
	HandleMethodNotAllowed bool
	ForwardedByClientIP    bool
}

func (g *GinModule) ModuleConfig() *declaration.ModuleConfig {
	return &declaration.ModuleConfig{
		ModuleName: "Gin",
	}
}

func (g *GinModule) Interceptor() *func(instance interface{}) {
	interceptor := func(ginClient interface{}) {
		if engine, ok := ginClient.(*gin.Engine); ok {
			fmt.Println(engine.BasePath())
		}
	}
	return &interceptor
}

func (g *GinModule) Register(interceptor *func(instance interface{})) error {
	var err error

	if g.DebugModule {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	ginEngin := gin.Default()

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

	g.server = &http.Server{
		Addr:    g.Address,
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
