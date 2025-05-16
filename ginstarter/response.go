package ginstarter

import (
	"github.com/acexy/golang-toolkit/util/json"
	"github.com/gin-gonic/gin"
	"net/http"
)

// Response 标准响应 用户可以通过自定义实现该接口定义自己的响应结构体
// 也可以使用NewRespRest来创建自定义响应结构体
type Response interface {

	// Data 响应的Body数据
	Data() *ResponseData
}

// ResponseDataStructDecoder 针对Response.Data() 响应的时结构体数据时的解码为[]byte功能
// 默认为JSON解码器 用户可以自定义实现该接口 实现自定义解码器
type ResponseDataStructDecoder interface {
	// Decode 解析响应数据
	Decode(response any) ([]byte, error)
}

// 默认解码器
type responseJsonDataStructDecoder struct {
}

func (r responseJsonDataStructDecoder) Decode(data any) ([]byte, error) {
	return json.ToJsonBytesError(data)
}

// restResp 默认的Rest响应结构体
type restResp struct {
	responseData *ResponseData
}

func (r *restResp) Data() *ResponseData {
	return r.responseData
}

// NewRespRest 创建一个Rest响应体
func NewRespRest() *restResp {
	resp := new(restResp)
	resp.responseData = &ResponseData{}
	resp.responseData.contentType = gin.MIMEJSON
	return resp
}

// DataBuilder 响应数据构造器
func (r *restResp) DataBuilder(fn func() *ResponseData) Response {
	r.responseData = fn()
	return r
}

// SetData 设置Rest标准的响应结构
func (r *restResp) SetData(data any) *ResponseData {
	bytes, err := ginConfig.ResponseDataStructDecoder.Decode(data)
	if err != nil {
		panic(err)
	}
	r.responseData.data = bytes
	return r.responseData
}

// SetDataResponse 设置Rest标准的响应结构 并返回响应体数据
func (r *restResp) SetDataResponse(data any) Response {
	bytes, err := ginConfig.ResponseDataStructDecoder.Decode(data)
	if err != nil {
		panic(err)
	}
	r.responseData.data = bytes
	return r
}

// ToResponse 转换为响应体数据
func (r *restResp) ToResponse() Response {
	return r
}

// RespRestRaw 响应标准格式的Rest原始数据
func RespRestRaw(dataRest *RestRespStruct) Response {
	return NewRespRest().SetDataResponse(dataRest)
}

// RespRestSuccess 响应标准格式的Rest成功数据
func RespRestSuccess(data ...any) Response {
	return NewRespRest().SetDataResponse(NewRestSuccess(data...))
}

// RespRestException 响应标准格式的Rest系统异常错误
func RespRestException(statusMessage ...string) Response {
	return NewRespRest().SetDataResponse(NewRestException(statusMessage...))
}

// RespRestBadParameters 响应标准格式的Rest参数错误
func RespRestBadParameters(statusMessage ...string) Response {
	return NewRespRest().SetDataResponse(NewRestBadParameters(statusMessage...))
}

// RespRestUnAuthorized 响应标准格式的Rest未授权错误
func RespRestUnAuthorized(statusMessage ...string) Response {
	return NewRespRest().SetDataResponse(NewRestUnauthorized(statusMessage...))
}

// RespRestStatusError 响应标准格式的Rest状态错误
func RespRestStatusError(statusCode StatusCode, statusMessage ...StatusMessage) Response {
	return NewRespRest().SetDataResponse(NewRestStatusError(statusCode, statusMessage...))
}

// RespRestBizError 响应标准格式的Rest业务错误
func RespRestBizError(bizErrorCode BizErrorCode, bizErrorMessage BizErrorMessage) Response {
	return NewRespRest().SetDataResponse(NewRestBizError(bizErrorCode, bizErrorMessage))
}

// commonResp 普通响应
type commonResp struct {
	responseData *ResponseData
}

func (c *commonResp) Data() *ResponseData {
	return c.responseData
}

// NewCommonResp 创建一个普通响应
func NewCommonResp() *commonResp {
	resp := new(commonResp)
	resp.responseData = &ResponseData{}
	return resp
}

// DataBuilder 响应数据构造器
func (c *commonResp) DataBuilder(fn func() *ResponseData) Response {
	c.responseData = fn()
	return c
}

// SetData 响应数据
func (c *commonResp) SetData(data *ResponseData) *ResponseData {
	c.responseData = data
	return c.responseData
}

// SetDataToResponse 响应数据
func (c *commonResp) SetDataToResponse(data *ResponseData) Response {
	c.responseData = data
	return c
}

// ToResponse 转换为响应体数据
func (c *commonResp) ToResponse() Response {
	return c
}

// RespHttpStatusCode 设置响应状态码
func RespHttpStatusCode(statusCode int) Response {
	return &commonResp{NewEmptyResponseData().SetStatusCode(statusCode)}
}

// RespJson 响应Json数据
func RespJson(data []byte, httpStatusCode ...int) Response {
	respData := NewEmptyResponseData()
	respData.SetData(data)
	statusCode := http.StatusOK
	respData.SetContentType(gin.MIMEJSON)
	if len(httpStatusCode) > 0 {
		statusCode = httpStatusCode[0]
	}
	respData.SetStatusCode(statusCode)
	return NewCommonResp().SetDataToResponse(respData)
}

func RespTextPlain(data []byte, httpStatusCode ...int) Response {
	respData := NewEmptyResponseData()
	respData.SetData(data)
	statusCode := http.StatusOK
	respData.SetContentType(gin.MIMEPlain)
	if len(httpStatusCode) > 0 {
		statusCode = httpStatusCode[0]
	}
	respData.SetStatusCode(statusCode)
	return NewCommonResp().SetDataToResponse(respData)
}
