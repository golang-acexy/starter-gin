package router

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-acexy/starter-gin/ginmodule"
	"net/http"
)

// RestStruct 自定义的Rest结构体
type RestStruct struct {
	Code int
	Msg  string
	Data any
}

// MyRestResponse 实现Gin Response定义 替换默认RespRest结构体风格
type MyRestResponse struct {
	result *RestStruct
	head   map[string]string
}

func (m MyRestResponse) Data() any {
	return m.result
}

func (m MyRestResponse) ContentType() string {
	return gin.MIMEJSON
}

func (m MyRestResponse) Headers() map[string]string {
	return m.head
}

func (m MyRestResponse) HttpStatusCode() int {
	return http.StatusOK
}

type MyRestRouter struct {
}

func (m *MyRestRouter) Info() *ginmodule.RouterInfo {
	return &ginmodule.RouterInfo{
		GroupPath: "my-rest",
	}
}

func (m *MyRestRouter) Handlers(router *ginmodule.RouterWrapper) {
	router.GET("invoke", m.invoke())
}

func (m *MyRestRouter) invoke() ginmodule.HandlerWrapper {
	return func(request *ginmodule.Request) (ginmodule.Response, error) {
		return MyRestResponse{
			result: &RestStruct{
				Code: 200,
				Msg:  "success",
				Data: "invoke",
			},
		}, nil
	}
}
