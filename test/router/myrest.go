package router

import (
	"github.com/acexy/golang-toolkit/util/json"
	"github.com/golang-acexy/starter-gin/ginmodule"
)

// RestStruct 自定义的Rest结构体
type RestStruct struct {
	Code int
	Msg  string
	Data any
}

type MyRestRouter struct {
}

func (m *MyRestRouter) Info() *ginmodule.RouterInfo {
	return &ginmodule.RouterInfo{
		GroupPath: "my-rest",
	}
}

func (m *MyRestRouter) Handlers(router *ginmodule.RouterWrapper) {
	router.GET("m1", m.m1())
	router.GET("m2", m.m2())
	router.GET("m3", m.m3())
}

// 使用框架自带的Rest响应默认Rest结构体
func (m *MyRestRouter) m1() ginmodule.HandlerWrapper {
	return func(request *ginmodule.Request) (ginmodule.Response, error) {
		return ginmodule.RespRestSuccess("data part"), nil
	}
}

// 使用框架自带的Rest响应自定义结构体
func (m *MyRestRouter) m2() ginmodule.HandlerWrapper {
	return func(request *ginmodule.Request) (ginmodule.Response, error) {
		return ginmodule.NewRespRest().RestDataResponse(&RestStruct{
			Code: 200,
			Msg:  "success",
			Data: "invoke",
		}), nil
	}
}

// 自实现Response响应数据
func (m *MyRestRouter) m3() ginmodule.HandlerWrapper {
	return func(request *ginmodule.Request) (ginmodule.Response, error) {
		response := &MyRestResponse{}
		response.setData(&RestStruct{
			Code: 200,
			Msg:  "success",
			Data: "my rest impl",
		})
		return response, nil
	}
}

type MyRestResponse struct {
	responseData *ginmodule.ResponseData
}

func (m *MyRestResponse) Data() *ginmodule.ResponseData {
	return m.responseData
}

func (m *MyRestResponse) setData(data *RestStruct) {
	m.responseData = ginmodule.NewResponseData()
	m.responseData.SetData(json.ToJsonBytes(data))
}
