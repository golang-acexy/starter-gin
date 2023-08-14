package router

import (
	"github.com/golang-acexy/starter-gin/ginmodule"
)

type AbortRouter struct {
}

func (a *AbortRouter) RouterInfo() *ginmodule.RouterInfo {
	return &ginmodule.RouterInfo{
		GroupPath: "abort",
	}
}

func (a *AbortRouter) RegisterHandler(ginWrapper *ginmodule.GinWrapper) {

	// path /abort/exception 主动返回的异常触发系统错误
	ginWrapper.GET("code", a.code())

}

func (a *AbortRouter) code() func(request *ginmodule.Request) (*ginmodule.Response, error) {
	return func(request *ginmodule.Request) (*ginmodule.Response, error) {
		gCtx := request.RawGinContext()
		gCtx.AbortWithStatus(401)
		return nil, nil
	}
}
