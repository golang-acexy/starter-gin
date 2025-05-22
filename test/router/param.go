package router

import (
	"fmt"
	"github.com/acexy/golang-toolkit/logger"
	"github.com/golang-acexy/starter-gin/ginstarter"
)

type ParamRouter struct {
}

func (d *ParamRouter) Info() *ginstarter.RouterInfo {
	return &ginstarter.RouterInfo{
		GroupPath: "param",
		PreInterceptors: []ginstarter.PreInterceptor{func(request *ginstarter.Request) (response ginstarter.Response, continuePreInterceptor bool, continueHandler bool) {
			logger.Logrus().Infoln("group interceptor invoke")
			return ginstarter.RespTextPlain([]byte("hello world"), 200), true, false
		}},
		PostInterceptors: []ginstarter.PostInterceptor{
			func(request *ginstarter.Request, response ginstarter.Response) (newResponse ginstarter.Response, continuePostInterceptor bool) {
				if response != nil {
					fmt.Println(response.Data().ToDebugString())
				}
				return ginstarter.NewRespRest().SetDataResponse("ok"), true
			},
		},
	}
}

func (d *ParamRouter) Handlers(router *ginstarter.RouterWrapper) {
	router.POST1("json", []string{"application-json"}, d.json())
	// demo path /param/uri-path/101/acexy
	router.GET("uri-path/:id/:name", d.path())
	// demo path /param/uri-path/query?id=1&name=acexy
	router.GET("uri-path/query", d.query())
	// demo path /param/body/json    body > {"id":1,"name":"acexy"}
	router.POST("body/json", d.json())
	// demo path /param/body/form    body > id=1&name=acexy
	router.POST("body/form", d.form())
	router.GET("bind-query", d.bindQuery())
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
	Id   uint   `json:"id" binding:"required,numeric"`
	Name string `json:"name" binding:"required"`
	Age  uint   `json:"age"`
	Fat  bool   `json:"fat"`
}

type BodyFormUser struct {
	Id     uint   `form:"id" binding:"required,min=10"`
	Name   string `form:"name" binding:"required"`
	Email  string `form:"email" binding:"required,email"`
	Domain string `form:"domain" binding:"domain"`
}

func (d *ParamRouter) path() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		fmt.Println("Request Ip", request.RequestIP())
		// 获取url路径参数
		uriParams := request.GetPathParams("id", "name", "unknown")
		fmt.Printf("uriPath %+v\n", uriParams)
		// demo path /param/uri-path/a/acexy 触发参数错误
		user := new(UriPathUser)
		request.MustBindPathParams(user)
		fmt.Printf("%+v\n", user)
		return ginstarter.RespRestSuccess(), nil
	}
}

func (d *ParamRouter) query() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		user := new(UriQueryUser)
		// demo path /param/uri-path/query?name=acexy 触发参数错误
		request.BindQueryParams(user)
		fmt.Printf("%+v\n", user)
		return ginstarter.RespRestSuccess(), nil
	}
}

func (d *ParamRouter) json() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		user := BodyJsonUser{}
		request.MustBindBodyJson(&user)
		fmt.Printf("%+v\n", user)
		return ginstarter.RespRestSuccess(user), nil
	}
}

func (d *ParamRouter) form() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		user := BodyFormUser{}
		request.MustBindBodyForm(&user)
		fmt.Printf("%+v\n", user)
		return ginstarter.RespRestSuccess(), nil
	}
}

func (d *ParamRouter) bindQuery() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		user := BodyFormUser{}
		request.MustBindQueryParams(&user)
		fmt.Printf("%+v\n", user)
		return ginstarter.RespRestSuccess(), nil
	}
}
