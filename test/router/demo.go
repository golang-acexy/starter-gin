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

func (d *DemoRouter) Info() *ginmodule.RouterInfo {
	return &ginmodule.RouterInfo{
		GroupPath: "demo",
	}
}

func (d *DemoRouter) Handlers(router *ginmodule.RouterWrapper) {

	// path /demo/error 未捕获的异常触发系统错误
	router.MATCH([]string{http.MethodGet, http.MethodPost}, "error", d.error())

	// path /demo/exception 主动返回的异常触发系统错误
	router.GET("exception", d.exception())

	// path /demo/hold 5s的请求hold
	router.GET("hold", d.hold())

	router.GET("empty", d.empty())
}

func (d *DemoRouter) error() ginmodule.HandlerWrapper {
	return func(request *ginmodule.Request) (ginmodule.Response, error) {
		fmt.Println("invoke")
		panic("error")
		return ginmodule.RespRestSuccess(), nil
	}
}

func (d *DemoRouter) exception() ginmodule.HandlerWrapper {
	return func(request *ginmodule.Request) (ginmodule.Response, error) {
		return ginmodule.RespRestSuccess(), errors.New("my exception")
	}
}

func (d *DemoRouter) hold() ginmodule.HandlerWrapper {
	return func(request *ginmodule.Request) (ginmodule.Response, error) {
		fmt.Println("invoke")
		time.Sleep(time.Second * 5)
		return ginmodule.RespTextPlain("文本响应"), nil
	}
}

func (d *DemoRouter) empty() ginmodule.HandlerWrapper {
	return func(request *ginmodule.Request) (ginmodule.Response, error) {
		return nil, nil
	}
}
