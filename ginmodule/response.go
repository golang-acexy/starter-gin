package ginmodule

import (
	"fmt"
	"github.com/acexy/golang-toolkit/util/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// 默认结构体数据解码为[]byte的处理器 json
var defaultResponseDataDecoder ResponseDataStructDecoder = responseJsonDataStructDecoder{}

// Response 标准响应 用户可以通过自定义实现该接口定义自己的响应结构体
type Response interface {

	// Data 响应的Body数据
	Data() any

	// ContentType 响应的ContentType
	ContentType() string

	// Headers 响应的Header 当值为""时，执行移除指定头名称
	Headers() map[string]string

	// HttpStatusCode 响应的StatusCode
	HttpStatusCode() int
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
	fmt.Println(json.ToJson(data))
	return json.ToJsonBytesError(data)
}

// restResp 默认的Rest响应结构体
type restResp struct {
	dataRest       *RestRespStruct
	dataString     string
	contentType    string
	headers        map[string]string
	httpStatusCode int
}

func (r *restResp) Data() any {
	if r.dataRest != nil {
		return json.ToJson(r.dataRest)
	}
	return r.dataString
}

func (r *restResp) ContentType() string {
	if r.contentType == "" {
		return gin.MIMEJSON
	}
	return r.contentType
}

func (r *restResp) Headers() map[string]string {
	return r.headers
}
func (r *restResp) HttpStatusCode() int {
	return r.httpStatusCode
}

// RespRestSuccess 响应标准格式的Rest成功数据
func RespRestSuccess(data ...any) Response {
	rest := &restResp{
		dataRest: &RestRespStruct{
			Status: &RestRespStatusStruct{
				StatusCode:    StatusCodeSuccess,
				StatusMessage: statusMessageSuccess,
				Timestamp:     time.Now().UnixMilli(),
			},
		},
		httpStatusCode: StatusCodeSuccess,
	}
	if len(data) > 0 {
		rest.dataRest.Data = data[0]
	}
	return rest
}

// RespRestException 响应标准格式的Rest异常错误
func RespRestException() Response {
	rest := &restResp{
		dataRest: &RestRespStruct{
			Status: &RestRespStatusStruct{
				StatusCode:    StatusCodeException,
				StatusMessage: statusMessageException,
				Timestamp:     time.Now().UnixMilli(),
			},
		},
		httpStatusCode: StatusCodeSuccess,
	}
	return rest
}

// RespRestStatusError 响应标准格式的Rest状态错误
func RespRestStatusError(statusCode StatusCode, statusMessage ...StatusMessage) Response {
	rest := &restResp{
		dataRest: &RestRespStruct{
			Status: &RestRespStatusStruct{
				StatusCode: statusCode,
				Timestamp:  time.Now().UnixMilli(),
			},
		},
		httpStatusCode: StatusCodeSuccess,
	}
	if len(statusMessage) > 0 {
		rest.dataRest.Status.StatusMessage = statusMessage[0]
	} else {
		rest.dataRest.Status.StatusMessage = GetStatusMessage(statusCode)
	}
	return rest
}

// RespRestBizError 响应标准格式的Rest业务错误
func RespRestBizError(bizErrorCode BizErrorCode, bizErrorMessage BizErrorMessage) Response {
	rest := &restResp{
		dataRest: &RestRespStruct{
			Status: &RestRespStatusStruct{
				StatusCode:      StatusCodeSuccess,
				StatusMessage:   statusMessageSuccess,
				BizErrorCode:    bizErrorCode,
				BizErrorMessage: bizErrorMessage,
				Timestamp:       time.Now().UnixMilli(),
			},
		},
		httpStatusCode: StatusCodeSuccess,
	}
	return rest
}

// ginRawResp 通过Gin原始上下文响应
type ginRawResp struct {
	ginFn func(context *gin.Context)
}

func (g ginRawResp) Data() any {
	return nil
}

func (g ginRawResp) ContentType() string {
	return ""
}

func (g ginRawResp) Headers() map[string]string {
	return nil
}

func (g ginRawResp) HttpStatusCode() int {
	return 200
}

// RespGinRaw 操作Gin原始上下文响应
func RespGinRaw(fn func(context *gin.Context)) Response {
	return ginRawResp{ginFn: fn}
}

// RespJson 响应Json数据
func RespJson(data any, httpStatusCode ...int) Response {
	return ginRawResp{ginFn: func(context *gin.Context) {
		statusCode := http.StatusOK
		if len(httpStatusCode) > 0 {
			statusCode = httpStatusCode[0]
		}
		context.JSON(statusCode, data)
	}}
}

// RespXml 响应Xml数据
func RespXml(data any, httpStatusCode ...int) Response {
	return ginRawResp{ginFn: func(context *gin.Context) {
		statusCode := http.StatusOK
		if len(httpStatusCode) > 0 {
			statusCode = httpStatusCode[0]
		}
		context.XML(statusCode, data)
	}}
}

// RespYaml 响应Yaml数据
func RespYaml(data any, httpStatusCode ...int) Response {
	return ginRawResp{ginFn: func(context *gin.Context) {
		statusCode := http.StatusOK
		if len(httpStatusCode) > 0 {
			statusCode = httpStatusCode[0]
		}
		context.YAML(statusCode, data)
	}}
}

// RespToml 响应Toml数据
func RespToml(data any, httpStatusCode ...int) Response {
	return ginRawResp{ginFn: func(context *gin.Context) {
		statusCode := http.StatusOK
		if len(httpStatusCode) > 0 {
			statusCode = httpStatusCode[0]
		}
		context.TOML(statusCode, data)
	}}
}

// RespTextPlain 响应Json数据
func RespTextPlain(data string, httpStatusCode ...int) Response {
	return ginRawResp{ginFn: func(context *gin.Context) {
		statusCode := http.StatusOK
		if len(httpStatusCode) > 0 {
			statusCode = httpStatusCode[0]
		}
		context.Data(statusCode, gin.MIMEPlain, []byte(data))
	}}
}
