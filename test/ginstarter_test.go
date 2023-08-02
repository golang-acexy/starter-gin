package test

import (
	"fmt"
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

	interceptor := func(instance interface{}) {
		engine := instance.(*gin.Engine)

		// 注册一个伪探活服务
		engine.GET("/ping", func(context *gin.Context) {
			context.String(http.StatusOK, "alive")
		})
	}

	moduleLoaders = []declaration.ModuleLoader{&ginmodule.GinModule{
		ListenAddress: ":8118",
		DebugModule:   true,
		Routers: []ginmodule.Router{
			&router.DemoRouter{},
			&router.ParamRouter{},
		},
		GinInterceptor: &interceptor,
	}}

	fmt.Println(moduleLoaders)
}

func TestGin(t *testing.T) {

	module := declaration.Module{
		ModuleLoaders: moduleLoaders,
	}

	err := module.Load()
	if err != nil {
		fmt.Printf("%+v\n", err)
		return
	}
	select {}
}
func TestGinUnload(t *testing.T) {

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
