package test

import (
	"fmt"
	"github.com/golang-acexy/starter-gin/ginmodule"
	"github.com/golang-acexy/starter-gin/test/router"
	"github.com/golang-acexy/starter-parent/parentmodule/declaration"
	"testing"
	"time"
)

var moduleLoaders []declaration.ModuleLoader

func init() {

	moduleLoaders = []declaration.ModuleLoader{&ginmodule.GinModule{
		ListenAddress: ":8118",
		DebugModule:   true,
		Routers: []ginmodule.Router{
			&router.DemoRouter{},
			&router.ParamRouter{},
		},
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
