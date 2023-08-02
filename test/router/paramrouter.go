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
	// demo path /param/uri-path/101/acexy
	ginWrapper.GET("uri-path/:id/:name", d.path())
	// demo path /param/uri-path/query?id=1&name=acexy
	ginWrapper.GET("uri-path/query", d.query())
	// demo path /param/body/json    body > {"id":1,"name":"acexy"}
	ginWrapper.POST("body/json", d.json())
	// demo path /param/body/form    body > id=1&name=acexy
	ginWrapper.POST("body/form", d.form())
}

type UriPathUser struct {
	Id   uint   `uri:"id" validate:"number"`
	Name string `uri:"name"`
}

type UriQueryUser struct {
	Id   uint   `form:"id" binding:"required"`
	Name string `form:"name"`
}

type BodyJsonUser struct {
	Id   uint   `json:"id"`
	Name string `json:"name" binding:"required"`
}

type BodyFormUser struct {
	Id   uint   `form:"id"`
	Name string `form:"name" binding:"required"`
}

func (d *ParamRouter) path() func(request *ginmodule.Request) (*ginmodule.Response, error) {
	return func(request *ginmodule.Request) (*ginmodule.Response, error) {
		fmt.Println("Request Ip", request.RequestIP())
		// 获取url路径参数
		uriParams := request.UriPathParams("id", "name", "unknown")
		fmt.Printf("uriPath %+v\n", uriParams)
		// demo path /param/uri-path/a/acexy 触发参数错误
		user := new(UriPathUser)
		request.BindUriPathParams(user)
		fmt.Printf("%+v\n", user)
		return ginmodule.ResponseSuccess(), nil
	}
}

func (d *ParamRouter) query() func(request *ginmodule.Request) (*ginmodule.Response, error) {
	return func(request *ginmodule.Request) (*ginmodule.Response, error) {
		user := new(UriQueryUser)
		// demo path /param/uri-path/query?name=acexy 触发参数错误
		request.BindUriQueryParams(user)
		fmt.Printf("%+v\n", user)
		return ginmodule.ResponseSuccess(), nil
	}
}

func (d *ParamRouter) json() func(request *ginmodule.Request) (*ginmodule.Response, error) {
	return func(request *ginmodule.Request) (*ginmodule.Response, error) {
		user := BodyJsonUser{}
		request.BindBodyJson(&user)
		fmt.Printf("%+v\n", user)
		return ginmodule.ResponseSuccess(), nil
	}
}

func (d *ParamRouter) form() func(request *ginmodule.Request) (*ginmodule.Response, error) {
	return func(request *ginmodule.Request) (*ginmodule.Response, error) {
		user := BodyFormUser{}
		request.BindBodyForm(&user)
		fmt.Printf("%+v\n", user)
		return ginmodule.ResponseSuccess(), nil
	}
}
