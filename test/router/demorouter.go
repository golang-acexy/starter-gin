package router

import (
	"errors"
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

	// path /demo/error 未捕获的异常触发系统错误
	ginWrapper.MATCH([]string{http.MethodGet, http.MethodPost}, "error", d.error())

	// path /demo/exception 主动返回的异常触发系统错误
	ginWrapper.GET("exception", d.exception())

	// path /demo/hold 5s的请求hold
	ginWrapper.GET("hold", d.hold())

	ginWrapper.GET("empty", d.empty())
}

func (d *DemoRouter) error() func(request *ginmodule.Request) (*ginmodule.Response, error) {
	return func(request *ginmodule.Request) (response *ginmodule.Response, err error) {
		fmt.Println("invoke")
		panic("error")
		return ginmodule.ResponseSuccess(), nil
	}
}

func (d *DemoRouter) exception() func(request *ginmodule.Request) (*ginmodule.Response, error) {
	return func(request *ginmodule.Request) (response *ginmodule.Response, err error) {
		return nil, errors.New("biz exception")
	}
}

func (d *DemoRouter) hold() func(request *ginmodule.Request) (*ginmodule.Response, error) {
	return func(request *ginmodule.Request) (*ginmodule.Response, error) {
		fmt.Println("invoke")
		time.Sleep(time.Second * 5)
		return ginmodule.ResponseSuccess(), nil
	}
}

func (d *DemoRouter) empty() func(request *ginmodule.Request) (*ginmodule.Response, error) {
	return func(request *ginmodule.Request) (*ginmodule.Response, error) {
		return nil, nil
	}
}
