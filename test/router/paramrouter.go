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

	// demo path /url/101/acexy/query
	ginWrapper.GET("uri/:id/:name/query", d.get())
}

type User struct {
	Id uint `uri:"id" validate:"number"`
}

func (d *ParamRouter) get() func(request *ginmodule.Request) (*ginmodule.Response, error) {
	return func(request *ginmodule.Request) (*ginmodule.Response, error) {

		fmt.Println("Request Ip", request.RequestIP())

		// 获取url路径参数
		uriParams := request.UriPathParams("id", "name", "unknown")
		fmt.Printf("uriPath %+v\n", uriParams)

		// demo path /url/a/acexy/query 触发错误
		user := new(User)
		request.BindUriPathParams(user)

		return ginmodule.ResponseSuccess(), nil
	}
}
