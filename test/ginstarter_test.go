package test

import (
	"fmt"
	"github.com/golang-acexy/starter-gin/ginmodule"
	"github.com/golang-acexy/starter-gin/test/router"
	"github.com/golang-acexy/starter-parent/parentmodule/declaration"
	"os"
	"os/signal"
	"syscall"
	"testing"
)

var moduleLoaders []declaration.ModuleLoader

func init() {

	moduleLoaders = []declaration.ModuleLoader{&ginmodule.GinModule{
		Address:     ":8118",
		DebugModule: true,
		Routers: []ginmodule.Router{
			&router.DemoRouter{},
		},
	}}

	fmt.Println(moduleLoaders)

}

func TestGin(t *testing.T) {
	module := declaration.Module{}
	err := module.Load(moduleLoaders)
	if err != nil {
		return
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	module.Unload(20)
	//var gracefully bool
	//gracefully, err = moduleLoader.Unregister(10) // 最大等待后台任务自动结束时间 秒
	//fmt.Println("gracefully shutdown?", gracefully)
}
