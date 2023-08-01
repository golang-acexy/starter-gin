package router

import (
	"fmt"
	"github.com/golang-acexy/starter-gin/ginmodule"
	"net/http"
	"time"
)

type DemoRouter struct {
}

func (d *DemoRouter) RouterInfo() *ginmodule.RouterInfo {
	return &ginmodule.RouterInfo{
		GroupPath: "demo",
	}
}

func (d *DemoRouter) RegisterHandler(ginWrapper *ginmodule.GinWrapper) {
	ginWrapper.MATCH([]string{http.MethodGet, http.MethodPost}, "error", d.error())
	ginWrapper.GET("hold", d.hold())
}

func (d *DemoRouter) error() func(request *ginmodule.Request) (*ginmodule.Response, error) {
	return func(request *ginmodule.Request) (response *ginmodule.Response, err error) {
		fmt.Println("invoke")
		i := 0
		_ = 1 / i
		return ginmodule.NewSuccess(), nil
	}
}

func (d *DemoRouter) hold() func(request *ginmodule.Request) (*ginmodule.Response, error) {
	return func(request *ginmodule.Request) (*ginmodule.Response, error) {
		time.Sleep(time.Hour)
		return nil, nil
	}
}
