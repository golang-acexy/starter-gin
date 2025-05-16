package router

import (
	"errors"
	"fmt"
	"github.com/golang-acexy/starter-gin/ginstarter"
	"net/http"
	"time"
)

type DemoRouter struct {
}

func (d *DemoRouter) Info() *ginstarter.RouterInfo {
	return &ginstarter.RouterInfo{
		GroupPath: "demo",
	}
}

func (d *DemoRouter) Handlers(router *ginstarter.RouterWrapper) {

	router.MATCH([]string{http.MethodGet, http.MethodPost}, "more", d.more())

	// path /demo/exception 主动返回的异常触发系统错误
	router.GET("error1", d.error1())
	router.GET("error2", d.error2())
	router.GET("error3", d.error3())

	// path /demo/hold 5s的请求hold
	router.GET("hold", d.hold())

	router.GET("empty", d.empty())

	router.GET("redirect", d.redirect())
	router.GET("common", d.common())
}

func (d *DemoRouter) more() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		fmt.Println("invoke")
		// 通过Builder来响应自定义Rest数据 并设置其他http属性
		return ginstarter.NewRespRest().DataBuilder(func() *ginstarter.ResponseData {
			return ginstarter.NewEmptyResponseData().SetStatusCode(http.StatusAccepted).SetData([]byte("success")).AddHeader("test", "test")
		}), nil
	}
}

func (d *DemoRouter) error1() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		// 通过error响应异常请求
		return nil, errors.New("return error")
	}
}

func (d *DemoRouter) error2() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		// 通过未处理的崩溃触发异常
		panic("panic exception")
	}
}

func (d *DemoRouter) error3() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		// 通过未处理的崩溃触发异常
		ii := 1
		ii -= 1
		i := 1 / ii
		fmt.Println(i)
		return nil, nil
	}
}

func (d *DemoRouter) hold() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		fmt.Println("invoke")
		time.Sleep(time.Second * 5)
		return ginstarter.RespTextPlain([]byte("text")), nil
	}
}

func (d *DemoRouter) empty() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		return nil, nil
	}
}

func (d *DemoRouter) redirect() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		request.RawGinContext().Redirect(http.StatusMovedPermanently, "https://google.com")
		return nil, nil
	}
}

func (d *DemoRouter) common() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		return ginstarter.NewCommonResp().DataBuilder(func() *ginstarter.ResponseData {
			return ginstarter.NewEmptyResponseData().SetData([]byte("success"))
		}), nil
	}
}
