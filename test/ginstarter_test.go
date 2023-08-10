package test

import (
	"fmt"
	"github.com/acexy/golang-toolkit/log"
	"github.com/gin-gonic/gin"
	"github.com/golang-acexy/starter-gin/ginmodule"
	"github.com/golang-acexy/starter-gin/test/router"
	"github.com/golang-acexy/starter-parent/parentmodule/declaration"
	"github.com/sirupsen/logrus"
	"net/http"
	"testing"
	"time"
)

var moduleLoaders []declaration.ModuleLoader

func init() {
	config := log.LogrusConfig{}
	config.EnableConsole(logrus.TraceLevel, false)

	interceptor := func(instance interface{}) {
		engine := instance.(*gin.Engine)

		// 使用interceptor的形式，获取原始gin实例 注册一个伪探活服务
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
