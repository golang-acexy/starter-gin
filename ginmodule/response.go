package ginmodule

import (
	"github.com/acexy/golang-toolkit/logger"
	"github.com/acexy/golang-toolkit/util/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// 默认结构体数据解码为[]byte的处理器 json
var defaultResponseDataDecoder ResponseDataStructDecoder = responseJsonDataStructDecoder{}

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
func (r *restResp) DataBuilder(fn func(data *ResponseData)) Response {
	fn(r.responseData)
	return r
}

// RestData 设置Rest标准的响应结构
func (r *restResp) RestData(data any) *ResponseData {
	bytes, err := defaultResponseDataDecoder.Decode(data)
	if err != nil {
		panic(err)
	}
	r.responseData.data = bytes
	return r.responseData
}

// RestDataResponse 设置Rest标准的响应结构 并返回响应体数据
func (r *restResp) RestDataResponse(data any) Response {
	bytes, err := defaultResponseDataDecoder.Decode(data)
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

// RespRestSuccess 响应标准格式的Rest成功数据
func RespRestSuccess(data ...any) Response {
	dataRest := &RestRespStruct{
		Status: &RestRespStatusStruct{
			StatusCode:    StatusCodeSuccess,
			StatusMessage: statusMessageSuccess,
			Timestamp:     time.Now().UnixMilli(),
		},
	}
	if len(data) > 0 {
		dataRest.Data = data[0]
	}
	return NewRespRest().RestDataResponse(dataRest)
}

// RespRestException 响应标准格式的Rest系统异常错误
func RespRestException() Response {
	dataRest := &RestRespStruct{
		Status: &RestRespStatusStruct{
			StatusCode:    StatusCodeException,
			StatusMessage: statusMessageException,
			Timestamp:     time.Now().UnixMilli(),
		},
	}
	return NewRespRest().RestDataResponse(dataRest)
}

// RespRestStatusError 响应标准格式的Rest状态错误
func RespRestStatusError(statusCode StatusCode, statusMessage ...StatusMessage) Response {
	dataRest := &RestRespStruct{
		Status: &RestRespStatusStruct{
			StatusCode: statusCode,
			Timestamp:  time.Now().UnixMilli(),
		},
	}
	if len(statusMessage) > 0 {
		dataRest.Status.StatusMessage = statusMessage[0]
	} else {
		dataRest.Status.StatusMessage = GetStatusMessage(statusCode)
	}
	return NewRespRest().RestDataResponse(dataRest)
}

// RespRestBizError 响应标准格式的Rest业务错误
func RespRestBizError(bizErrorCode BizErrorCode, bizErrorMessage BizErrorMessage) Response {
	dataRest := &RestRespStruct{
		Status: &RestRespStatusStruct{
			StatusCode:      StatusCodeSuccess,
			StatusMessage:   statusMessageSuccess,
			BizErrorCode:    bizErrorCode,
			BizErrorMessage: bizErrorMessage,
			Timestamp:       time.Now().UnixMilli(),
		},
	}
	return NewRespRest().RestDataResponse(dataRest)
}

// commonResp 普通响应
type commonResp struct {
	ginFn        func(context *gin.Context)
	responseData *ResponseData
}

func (c *commonResp) Data() *ResponseData {
	return nil
}

// NewCommonResp 创建一个普通响应
func NewCommonResp() *commonResp {
	resp := new(commonResp)
	resp.responseData = &ResponseData{}
	return resp
}

// DataBuilder 响应数据构造器
func (r *commonResp) DataBuilder(fn func(data *ResponseData)) Response {
	fn(r.responseData)
	return r
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

// RespStatusCode 设置响应状态码
func RespStatusCode(statusCode int) Response {
	return &commonResp{ginFn: func(context *gin.Context) {
		context.Status(statusCode)
	}}
}

// RespAbortWithStatus 设置响应状态码并设置忽略执行后续handler
func RespAbortWithStatus(statusCode int) Response {
	return &commonResp{ginFn: func(context *gin.Context) {
		context.AbortWithStatus(statusCode)
	}}
}

// RespJson 响应Json数据
func RespJson(data any, httpStatusCode ...int) Response {
	return &commonResp{ginFn: func(context *gin.Context) {
		statusCode := http.StatusOK
		if len(httpStatusCode) > 0 {
			statusCode = httpStatusCode[0]
		}
		context.JSON(statusCode, data)
	}}
}

// RespXml 响应Xml数据
func RespXml(data any, httpStatusCode ...int) Response {
	return &commonResp{ginFn: func(context *gin.Context) {
		statusCode := http.StatusOK
		if len(httpStatusCode) > 0 {
			statusCode = httpStatusCode[0]
		}
		context.XML(statusCode, data)
	}}
}

// RespYaml 响应Yaml数据
func RespYaml(data any, httpStatusCode ...int) Response {
	return &commonResp{ginFn: func(context *gin.Context) {
		statusCode := http.StatusOK
		if len(httpStatusCode) > 0 {
			statusCode = httpStatusCode[0]
		}
		context.YAML(statusCode, data)
	}}
}

// RespToml 响应Toml数据
func RespToml(data any, httpStatusCode ...int) Response {
	return &commonResp{ginFn: func(context *gin.Context) {
		statusCode := http.StatusOK
		if len(httpStatusCode) > 0 {
			statusCode = httpStatusCode[0]
		}
		context.TOML(statusCode, data)
	}}
}

// RespTextPlain 响应Json数据
func RespTextPlain(data string, httpStatusCode ...int) Response {
	return &commonResp{ginFn: func(context *gin.Context) {
		statusCode := http.StatusOK
		if len(httpStatusCode) > 0 {
			statusCode = httpStatusCode[0]
		}
		context.Data(statusCode, gin.MIMEPlain, []byte(data))
	}}
}

// RespRedirect 响应重定向
func RespRedirect(url string, httpStatusCode ...int) Response {
	return &commonResp{ginFn: func(context *gin.Context) {
		statusCode := http.StatusMovedPermanently
		if len(httpStatusCode) > 0 {
			statusCode = httpStatusCode[0]
			if statusCode < http.StatusMultipleChoices && statusCode > http.StatusPermanentRedirect {
				logger.Logrus().Warningln("Bad redirect status code", statusCode, "maybe not work")
			}
		}
		context.Redirect(statusCode, url)
	}}
}
