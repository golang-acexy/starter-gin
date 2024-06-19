package router

import (
	"fmt"
	"github.com/golang-acexy/starter-gin/ginmodule"
)

type ParamRouter struct {
}

func (d *ParamRouter) Info() *ginmodule.RouterInfo {
	return &ginmodule.RouterInfo{
		GroupPath: "param",
	}
}

func (d *ParamRouter) Handlers(router *ginmodule.RouterWrapper) {
	// demo path /param/uri-path/101/acexy
	router.GET("uri-path/:id/:name", d.path())
	// demo path /param/uri-path/query?id=1&name=acexy
	router.GET("uri-path/query", d.query())
	// demo path /param/body/json    body > {"id":1,"name":"acexy"}
	router.POST("body/json", d.json())
	// demo path /param/body/form    body > id=1&name=acexy
	router.POST("body/form", d.form())
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

func (d *ParamRouter) path() ginmodule.HandlerWrapper {
	return func(request *ginmodule.Request) (ginmodule.Response, error) {
		fmt.Println("Request Ip", request.RequestIP())
		// 获取url路径参数
		uriParams := request.UriPathParams("id", "name", "unknown")
		fmt.Printf("uriPath %+v\n", uriParams)
		// demo path /param/uri-path/a/acexy 触发参数错误
		user := new(UriPathUser)
		request.BindUriPathParams(user)
		fmt.Printf("%+v\n", user)
		return ginmodule.RespRestSuccess(), nil
	}
}

func (d *ParamRouter) query() ginmodule.HandlerWrapper {
	return func(request *ginmodule.Request) (ginmodule.Response, error) {
		user := new(UriQueryUser)
		// demo path /param/uri-path/query?name=acexy 触发参数错误
		request.BindUriQueryParams(user)
		fmt.Printf("%+v\n", user)
		return ginmodule.RespRestSuccess(), nil
	}
}

func (d *ParamRouter) json() ginmodule.HandlerWrapper {
	return func(request *ginmodule.Request) (ginmodule.Response, error) {
		user := BodyJsonUser{}
		request.BindBodyJson(&user)
		fmt.Printf("%+v\n", user)
		return ginmodule.RespRestSuccess(), nil
	}
}

func (d *ParamRouter) form() ginmodule.HandlerWrapper {
	return func(request *ginmodule.Request) (ginmodule.Response, error) {
		user := BodyFormUser{}
		request.BindBodyForm(&user)
		fmt.Printf("%+v\n", user)
		return ginmodule.RespRestSuccess(), nil
	}
}
