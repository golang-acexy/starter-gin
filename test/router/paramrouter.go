package router

import (
	"fmt"
	"github.com/golang-acexy/starter-gin/ginmodule"
)

type ParamRouter struct {
}

func (d *ParamRouter) RouterInfo() *ginmodule.RouterInfo {
	return &ginmodule.RouterInfo{
		GroupPath: "param",
	}
}

func (d *ParamRouter) RegisterHandler(ginWrapper *ginmodule.GinWrapper) {

	// /url/101/acexy/query
	ginWrapper.GET("uri/:id/:name/query", d.get())
}

func (d *ParamRouter) get() func(request *ginmodule.Request) (*ginmodule.Response, error) {
	return func(request *ginmodule.Request) (*ginmodule.Response, error) {

		fmt.Println("Request Ip", request.RequestIP())

		// 获取url路径参数
		uriParams := request.UriPathParams("id", "name", "unknown")
		fmt.Printf("uriPath %+v\n", uriParams)

		return ginmodule.NewSuccess(), nil
	}
}
