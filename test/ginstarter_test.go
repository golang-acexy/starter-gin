package test

import (
	"fmt"
	"github.com/acexy/golang-toolkit/sys"
	"github.com/gin-gonic/gin"
	"github.com/golang-acexy/starter-gin/ginmodule"
	"github.com/golang-acexy/starter-gin/test/router"
	"github.com/golang-acexy/starter-parent/parentmodule/declaration"
	"net/http"
	"testing"
	"time"
)

var moduleLoaders []declaration.ModuleLoader

func init() {
	interceptor := func(instance *gin.Engine) {
		// 使用interceptor的形式，获取原始gin实例 注册一个伪探活服务
		instance.GET("/ping", func(context *gin.Context) {
			context.String(http.StatusOK, "alive")
		})
	}

	moduleLoaders = []declaration.ModuleLoader{&ginmodule.GinModule{
		ListenAddress: ":8080",
		DebugModule:   true,
		Routers: []ginmodule.Router{
			&router.DemoRouter{},
			&router.ParamRouter{},
			&router.AbortRouter{},
			&router.BasicAuthRouter{},
		},
		GinInterceptor: interceptor,
	}}

}

// 默认Gin表现行为
func TestGinDefault(t *testing.T) {
	module := declaration.Module{
		ModuleLoaders: moduleLoaders,
	}

	err := module.Load()
	if err != nil {
		fmt.Printf("%+v\n", err)
		return
	}

	sys.ShutdownHolding()
}

// 自定义Gin的表现
// 将在和默认行为相同的路由功能代码下表现不同的响应
func TestGinCustomer(t *testing.T) {

	ginConfig := []declaration.ModuleLoader{&ginmodule.GinModule{
		ListenAddress:                ":8080",
		DebugModule:                  true,
		DisableHttpStatusCodeHandler: true,
		Routers: []ginmodule.Router{
			&router.DemoRouter{},
			&router.ParamRouter{},
			&router.AbortRouter{},
			&router.BasicAuthRouter{},
			&router.MyRestRouter{},
		},
	}}

	module := declaration.Module{
		ModuleLoaders: ginConfig,
	}

	err := module.Load()
	if err != nil {
		fmt.Printf("%+v\n", err)
		return
	}

	sys.ShutdownHolding()
}

func TestGinLoadAndUnload(t *testing.T) {
	module := declaration.Module{
		ModuleLoaders: moduleLoaders,
	}

	err := module.Load()
	if err != nil {
		fmt.Printf("%+v\n", err)
		return
	}

	time.Sleep(time.Second * 5)

	shutdownResult := module.UnloadByConfig()
	fmt.Printf("%+v\n", shutdownResult)
}
