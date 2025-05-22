package ginstarter

import (
	"bytes"
	"fmt"
	"github.com/acexy/golang-toolkit/logger"
	"github.com/acexy/golang-toolkit/sys"
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

func httpResponse(context *gin.Context, response Response) {
	if response == nil {
		return
	}
	context.Set(ginCtxKeyCurrentResponse, response)

	// 是否启用traceId响应
	if ginConfig.EnableGoroutineTraceIdResponse && sys.IsEnabledLocalTraceId() {
		context.Header("Trace-Id", sys.GetLocalTraceId())
	}

	responseData := response.Data()
	if responseData == nil {
		return
	}

	contentType := responseData.contentType
	if contentType == "" {
		contentType = gin.MIMEJSON
	}

	httpStatusCode := responseData.statusCode
	if httpStatusCode == 0 {
		httpStatusCode = http.StatusOK
	}

	cookies := responseData.cookies
	if len(cookies) > 0 {
		for _, v := range cookies {
			context.SetCookie(v.name, v.value, v.maxAge, v.path, v.domain, v.secure, v.httpOnly)
		}
	}

	headers := responseData.headers
	if len(headers) > 0 {
		for _, v := range headers {
			context.Header(v.name, v.value)
		}
	}

	data := responseData.data
	if len(data) > 0 {
		writer := context.Writer
		if w, ok := writer.(*responseRewriter); ok {
			w.Rest() // 重置响应体
		}
		context.Data(httpStatusCode, contentType, data)
		if context.ContentType() != contentType {
			context.Header("Content-Type", contentType)
		}
	}
}

// 支持将gin statusCode重写的响应处理器
type responseRewriter struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (r *responseRewriter) WriteHeader(code int) {
	r.statusCode = code
}

func (r *responseRewriter) Write(data []byte) (int, error) {
	return r.body.Write(data)
}

func (r *responseRewriter) WriteHeaderNow() {
	if !r.Written() {
		r.ResponseWriter.WriteHeader(r.statusCode)
	}
}

func (r *responseRewriter) Status() int {
	return r.statusCode
}

func (r *responseRewriter) Rest() {
	if r.body.Len() > 0 {
		r.body.Reset()
	}
}

// ResponseData 标准响应数据内容
type ResponseData struct {
	// body响应体负载数据
	data []byte
	// ContentType 响应的ContentType
	contentType string
	// 响应状态码
	statusCode int
	// 响应头
	headers []*ResponseHeader
	// 响应Cookie
	cookies []*ResponseCookie
}

// ResponseHeader 响应头
type ResponseHeader struct {
	name string
	// 设置零值可以清除该Name响应头
	value string
}

// ResponseCookie 响应Cookie
type ResponseCookie struct {
	name     string
	value    string
	maxAge   int
	path     string
	domain   string
	secure   bool
	httpOnly bool
}

func NewEmptyResponseData() *ResponseData {
	return &ResponseData{}
}

func NewResponseData(contentType string, body []byte) *ResponseData {
	return &ResponseData{
		contentType: contentType,
		data:        body,
	}
}

func NewResponseDataWithStatusCode(contentType string, body []byte, statusCode int) *ResponseData {
	return &ResponseData{
		contentType: contentType,
		data:        body,
		statusCode:  statusCode,
	}
}

func NewHeader(name, value string) *ResponseHeader {
	return &ResponseHeader{name: name, value: value}
}

func NewCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) *ResponseCookie {
	return &ResponseCookie{name: name, value: value, maxAge: maxAge, path: path, domain: domain, secure: secure, httpOnly: httpOnly}
}

func (r *ResponseData) SetData(data []byte) *ResponseData {
	r.data = data
	return r
}

func (r *ResponseData) SetContentType(contentType string) *ResponseData {
	if r.contentType != "" {
		logger.Logrus().Traceln("rewrite rest response content-type current =", r.contentType, "target =", contentType)
	}
	r.contentType = contentType
	return r
}

func (r *ResponseData) SetStatusCode(statusCode int) *ResponseData {
	r.statusCode = statusCode
	return r
}

func (r *ResponseData) AddHeaders(headers []*ResponseHeader) *ResponseData {
	if len(r.headers) != 0 {
		r.headers = append(r.headers, headers...)
	} else {
		r.headers = headers
	}
	return r
}

func (r *ResponseData) AddHeader(name, value string) *ResponseData {
	if len(r.headers) == 0 {
		r.headers = []*ResponseHeader{{
			name:  name,
			value: value,
		}}
	} else {
		r.headers = append(r.headers, &ResponseHeader{
			name:  name,
			value: value,
		})
	}
	return r
}

func (r *ResponseData) AddCookies(cookies []*ResponseCookie) *ResponseData {
	if len(r.cookies) != 0 {
		r.cookies = append(r.cookies, cookies...)
	} else {
		r.cookies = cookies
	}
	return r
}

func (r *ResponseData) AddCookie(cookie *ResponseCookie) *ResponseData {
	if len(r.cookies) == 0 {
		r.cookies = []*ResponseCookie{cookie}
	} else {
		r.cookies = append(r.cookies, cookie)
	}
	return r
}

func (r *ResponseData) ToDebugString() string {
	return fmt.Sprintf("body: %s head: %v content-type: %s", string(r.data), r.headers, r.contentType)
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
