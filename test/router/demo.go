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

	router.MATCH([]string{http.MethodGet, http.MethodPost}, "more", d.more())

	// path /demo/exception 主动返回的异常触发系统错误
	router.GET("error1", d.error1())
	router.GET("error2", d.error2())

	// path /demo/hold 5s的请求hold
	router.GET("hold", d.hold())

	router.GET("empty", d.empty())

	router.GET("redirect", d.redirect())
}

func (d *DemoRouter) more() ginmodule.HandlerWrapper {
	return func(request *ginmodule.Request) (ginmodule.Response, error) {
		fmt.Println("invoke")
		// 通过Builder来响应自定义Rest数据 并设置其他http属性
		return ginmodule.NewRespRest().DataBuilder(func(data *ginmodule.ResponseData) {
			data.SetStatusCode(http.StatusAccepted).SetData([]byte("success")).AddHeader(ginmodule.NewHeader("test", "test"))
		}), nil
	}
}

func (d *DemoRouter) error1() ginmodule.HandlerWrapper {
	return func(request *ginmodule.Request) (ginmodule.Response, error) {
		// 通过error响应异常请求
		return nil, errors.New("return error")
	}
}

func (d *DemoRouter) error2() ginmodule.HandlerWrapper {
	return func(request *ginmodule.Request) (ginmodule.Response, error) {
		// 通过未处理的崩溃触发异常
		panic("panic exception")
	}
}

func (d *DemoRouter) hold() ginmodule.HandlerWrapper {
	return func(request *ginmodule.Request) (ginmodule.Response, error) {
		fmt.Println("invoke")
		time.Sleep(time.Second * 5)
		return ginmodule.RespTextPlain("text"), nil
	}
}

func (d *DemoRouter) empty() ginmodule.HandlerWrapper {
	return func(request *ginmodule.Request) (ginmodule.Response, error) {
		return nil, nil
	}
}

func (d *DemoRouter) redirect() ginmodule.HandlerWrapper {
	return func(request *ginmodule.Request) (ginmodule.Response, error) {
		return ginmodule.RespRedirect("https://google.com"), nil
	}
}
