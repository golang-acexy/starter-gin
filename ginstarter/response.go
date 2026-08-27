package ginstarter

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/acexy/golang-toolkit/util/json"
	"github.com/gin-gonic/gin"
)

// Response 标准响应，用户可以通过实现该接口定义自己的响应类型。
// 也可以使用 NewRespRest 创建响应。
type Response interface {
	// Data 返回完整响应实体，包含状态码、Content-Type、Header、Cookie 和 Body。
	Data() *ResponseEntity
}

// ResponseBodyEncoder 将结构化响应正文编码为字节数据。
// 默认使用 JSON 编码器，用户可以通过实现该接口提供自定义编码方式。
type ResponseBodyEncoder interface {
	// Encode 编码响应正文。
	Encode(body any) ([]byte, error)
}

// 默认 JSON 响应正文编码器。
type jsonResponseBodyEncoder struct {
}

func (r jsonResponseBodyEncoder) Encode(body any) ([]byte, error) {
	return json.ToBytesError(body)
}

func currentResponseBodyEncoder() ResponseBodyEncoder {
	if config := currentGinConfig(); config != nil && config.ResponseBodyEncoder != nil {
		return config.ResponseBodyEncoder
	}
	return jsonResponseBodyEncoder{}
}

// writeResponse 是框架托管响应的唯一写入点。
func writeResponse(context *gin.Context, response Response) {
	if response == nil {
		context.Status(http.StatusOK)
		return
	}
	responseEntity := response.Data()
	if responseEntity == nil {
		context.Status(http.StatusOK)
		return
	}

	// 是否启用traceId响应
	if config := currentGinConfig(); config != nil && config.TraceIDResponse != nil {
		context.Header("Trace-Id", config.TraceIDResponse())
	}

	contentType := responseEntity.contentType
	if contentType == "" && len(responseEntity.body) > 0 {
		contentType = http.DetectContentType(responseEntity.body)
	}

	httpStatusCode := responseEntity.statusCode
	if httpStatusCode == 0 {
		httpStatusCode = http.StatusOK
	}

	cookies := responseEntity.cookies
	if len(cookies) > 0 {
		for _, cookie := range cookies {
			if cookie != nil {
				http.SetCookie(context.Writer, cookie)
			}
		}
	}

	headers := responseEntity.headers
	if len(headers) > 0 {
		for name, values := range headers {
			for _, value := range values {
				context.Writer.Header().Add(name, value)
			}
		}
	}

	context.Data(httpStatusCode, contentType, responseEntity.body)
}

// ResponseEntity 完整响应实体，描述框架最终写入客户端的全部响应信息。
type ResponseEntity struct {
	// body响应体负载数据
	body []byte
	// ContentType 响应的ContentType
	contentType string
	// 响应状态码
	statusCode int
	// 响应头
	headers http.Header
	// 响应Cookie
	cookies []*http.Cookie
}

func NewEmptyResponseEntity() *ResponseEntity {
	return &ResponseEntity{}
}

func NewResponseEntity(contentType string, body []byte) *ResponseEntity {
	return &ResponseEntity{
		contentType: contentType,
		body:        body,
	}
}

func NewResponseEntityWithStatusCode(contentType string, body []byte, statusCode int) *ResponseEntity {
	return &ResponseEntity{
		contentType: contentType,
		body:        body,
		statusCode:  statusCode,
	}
}

func NewCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) *http.Cookie {
	return &http.Cookie{Name: name, Value: value, MaxAge: maxAge, Path: path, Domain: domain, Secure: secure, HttpOnly: httpOnly}
}

func (r *ResponseEntity) SetBody(body []byte) *ResponseEntity {
	r.body = body
	return r
}

func (r *ResponseEntity) SetContentType(contentType string) *ResponseEntity {
	r.contentType = contentType
	return r
}

func (r *ResponseEntity) SetStatusCode(statusCode int) *ResponseEntity {
	r.statusCode = statusCode
	return r
}

func (r *ResponseEntity) AddHeaders(headers http.Header) *ResponseEntity {
	for name, values := range headers {
		for _, value := range values {
			r.AddHeader(name, value)
		}
	}
	return r
}

// SetHeader 设置响应头，同名响应头的旧值会被覆盖。
func (r *ResponseEntity) SetHeader(name, value string) *ResponseEntity {
	if name == "" {
		return r
	}
	if r.headers == nil {
		r.headers = make(http.Header)
	}
	r.headers.Set(name, value)
	return r
}

// AddHeader 添加响应头，同名响应头可以保留多个值。
func (r *ResponseEntity) AddHeader(name, value string) *ResponseEntity {
	if name == "" {
		return r
	}
	if r.headers == nil {
		r.headers = make(http.Header)
	}
	r.headers.Add(name, value)
	return r
}

func (r *ResponseEntity) AddCookies(cookies []*http.Cookie) *ResponseEntity {
	for _, cookie := range cookies {
		if cookie != nil {
			r.cookies = append(r.cookies, cookie)
		}
	}
	return r
}

func (r *ResponseEntity) AddCookie(cookie *http.Cookie) *ResponseEntity {
	if cookie != nil {
		r.cookies = append(r.cookies, cookie)
	}
	return r
}

func (r *ResponseEntity) DebugString(maxBodyBytes int) string {
	body := r.body
	truncated := false
	if maxBodyBytes >= 0 && len(body) > maxBodyBytes {
		body = body[:maxBodyBytes]
		truncated = true
	}
	return fmt.Sprintf("status-code: %d body: %s body-length: %d truncated: %t headers: %v content-type: %s", r.statusCode, string(body), len(r.body), truncated, r.headers, r.contentType)
}

// Body 返回响应正文副本，避免调用方绕过 SetBody 修改内部数据。
func (r *ResponseEntity) Body() []byte {
	return bytes.Clone(r.body)
}

// UnsafeRawBody 返回内部响应正文，仅用于确实需要原地修改的场景。
func (r *ResponseEntity) UnsafeRawBody() []byte {
	return r.body
}

// responseImpl 是 REST 响应和普通响应共用的标准实现。
type responseImpl struct {
	responseEntity *ResponseEntity
}

func (r *responseImpl) Data() *ResponseEntity {
	return r.responseEntity
}

// NewRespRest 创建一个Rest响应体
func NewRespRest() *responseImpl {
	return &responseImpl{responseEntity: &ResponseEntity{contentType: gin.MIMEJSON}}
}

// NewCommonResp 创建一个普通响应。
func NewCommonResp() *responseImpl {
	return &responseImpl{responseEntity: &ResponseEntity{}}
}

// DataBuilder 使用构造函数替换完整响应实体。
func (r *responseImpl) DataBuilder(fn func() *ResponseEntity) Response {
	if fn == nil {
		panic(ErrResponseEntityBuilderNil)
	}
	entity := fn()
	if entity == nil {
		panic(ErrResponseEntityNil)
	}
	r.responseEntity = entity
	return r
}

// SetBody 编码结构化正文并返回响应实体。
func (r *responseImpl) SetBody(body any) *ResponseEntity {
	bodyBytes, err := currentResponseBodyEncoder().Encode(body)
	if err != nil {
		panic(err)
	}
	r.responseEntity.body = bodyBytes
	return r.responseEntity
}

// SetBodyResponse 编码结构化正文并返回标准响应。
func (r *responseImpl) SetBodyResponse(body any) Response {
	bodyBytes, err := currentResponseBodyEncoder().Encode(body)
	if err != nil {
		panic(err)
	}
	r.responseEntity.body = bodyBytes
	return r
}

// SetEntity 设置完整响应实体并返回该实体。
func (r *responseImpl) SetEntity(entity *ResponseEntity) *ResponseEntity {
	if entity == nil {
		panic(ErrResponseEntityNil)
	}
	r.responseEntity = entity
	return r.responseEntity
}

// SetEntityResponse 设置完整响应实体并返回标准响应。
func (r *responseImpl) SetEntityResponse(entity *ResponseEntity) Response {
	r.SetEntity(entity)
	return r
}

// RespRestRaw 响应标准格式的Rest原始数据
func RespRestRaw(dataRest *RestRespStruct) Response {
	return NewRespRest().SetBodyResponse(dataRest)
}

// RespRestSuccess 响应标准格式的Rest成功数据
func RespRestSuccess(data ...any) Response {
	return NewRespRest().SetBodyResponse(NewRestSuccess(data...))
}

// RespRestException 响应标准格式的Rest系统异常错误
func RespRestException(statusMessage ...string) Response {
	return NewRespRest().SetBodyResponse(NewRestException(statusMessage...))
}

// RespRestBadParameters 响应标准格式的Rest参数错误
func RespRestBadParameters(statusMessage ...string) Response {
	return NewRespRest().SetBodyResponse(NewRestBadParameters(statusMessage...))
}

// RespRestUnauthorized 响应标准格式的 REST 未授权错误。
func RespRestUnauthorized(statusMessage ...string) Response {
	return NewRespRest().SetBodyResponse(NewRestUnauthorized(statusMessage...))
}

// RespRestStatusError 响应标准格式的Rest状态错误
func RespRestStatusError(statusCode StatusCode, statusMessage ...StatusMessage) Response {
	return NewRespRest().SetBodyResponse(NewRestStatusError(statusCode, statusMessage...))
}

// RespRestBizError 响应标准格式的Rest业务错误
func RespRestBizError(bizErrorCode BizErrorCode, bizErrorMessage BizErrorMessage) Response {
	return NewRespRest().SetBodyResponse(NewRestBizError(bizErrorCode, bizErrorMessage))
}

// RespHTTPStatusCode 设置响应状态码。
func RespHTTPStatusCode(statusCode int) Response {
	return NewCommonResp().SetEntityResponse(NewEmptyResponseEntity().SetStatusCode(statusCode))
}

// RespJSON 将结构化正文编码为 JSON 响应。
func RespJSON(body any, httpStatusCode ...int) Response {
	bodyBytes, err := currentResponseBodyEncoder().Encode(body)
	if err != nil {
		panic(err)
	}
	return RespJSONRaw(bodyBytes, httpStatusCode...)
}

// RespJSONRaw 响应已经编码完成的 JSON 数据。
func RespJSONRaw(body []byte, httpStatusCode ...int) Response {
	respData := NewEmptyResponseEntity()
	respData.SetBody(body)
	statusCode := http.StatusOK
	respData.SetContentType(gin.MIMEJSON)
	if len(httpStatusCode) > 0 {
		statusCode = httpStatusCode[0]
	}
	respData.SetStatusCode(statusCode)
	return NewCommonResp().SetEntityResponse(respData)
}

// RespTextPlain 响应Text数据
func RespTextPlain(data []byte, httpStatusCode ...int) Response {
	respData := NewEmptyResponseEntity()
	respData.SetBody(data)
	statusCode := http.StatusOK
	respData.SetContentType(gin.MIMEPlain)
	if len(httpStatusCode) > 0 {
		statusCode = httpStatusCode[0]
	}
	respData.SetStatusCode(statusCode)
	return NewCommonResp().SetEntityResponse(respData)
}
