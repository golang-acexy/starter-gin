package ginstarter

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/acexy/golang-toolkit/util/json"
	"github.com/gin-gonic/gin"
)

// Response supplies a request-owned entity for framework writing.
// Data must return a stable value without panicking; final writing is outside downstream recovery.
type Response interface {
	// Data returns the response entity; nil is written as an empty HTTP 200.
	Data() *ResponseEntity
}

// ResponseBodyEncoder encodes structured bodies; shared implementations must support concurrent requests.
type ResponseBodyEncoder interface {
	// Encode returns encoded bytes or an error; response helpers panic on encoding errors.
	Encode(body any) ([]byte, error)
}

// jsonResponseBodyEncoder is the default structured-body encoder.
type jsonResponseBodyEncoder struct {
}

// Encode serializes a structured body as JSON and returns serialization errors.
func (r jsonResponseBodyEncoder) Encode(body any) ([]byte, error) {
	return json.ToBytesError(body)
}

func currentResponseBodyEncoder() ResponseBodyEncoder {
	if config := currentGinConfig(); config != nil && config.ResponseBodyEncoder != nil {
		return config.ResponseBodyEncoder
	}
	return jsonResponseBodyEncoder{}
}

// writeResponse is the single commit point for framework-managed responses.
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

	// Attach the configured trace identifier before writing output.
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

// ResponseEntity holds mutable response data and is not safe for concurrent mutation.
// Body inputs and cookies are borrowed; Body copies bytes while UnsafeRawBody exposes them.
type ResponseEntity struct {
	// body holds borrowed payload bytes.
	body []byte
	// contentType specifies the response media type.
	contentType string
	// statusCode defaults to HTTP 200 when zero.
	statusCode int
	// headers holds additional response headers.
	headers http.Header
	// cookies holds borrowed response cookies.
	cookies []*http.Cookie
}

// NewEmptyResponseEntity creates an empty entity; an unset status defaults to HTTP 200 when written.
func NewEmptyResponseEntity() *ResponseEntity {
	return &ResponseEntity{}
}

// NewResponseEntity borrows body bytes and sets the media type; status is unset.
func NewResponseEntity(contentType string, body []byte) *ResponseEntity {
	return &ResponseEntity{
		contentType: contentType,
		body:        body,
	}
}

// NewResponseEntityWithStatusCode borrows body bytes and sets status/media type; resolver policy may replace it.
func NewResponseEntityWithStatusCode(contentType string, body []byte, statusCode int) *ResponseEntity {
	return &ResponseEntity{
		contentType: contentType,
		body:        body,
		statusCode:  statusCode,
	}
}

// NewCookie creates a cookie using net/http MaxAge semantics: zero omits expiry, negative requests deletion.
func NewCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) *http.Cookie {
	return &http.Cookie{Name: name, Value: value, MaxAge: maxAge, Path: path, Domain: domain, Secure: secure, HttpOnly: httpOnly}
}

// SetBody borrows body bytes without copying or encoding.
func (r *ResponseEntity) SetBody(body []byte) *ResponseEntity {
	r.body = body
	return r
}

// SetContentType sets the response media type; empty allows body-based detection.
func (r *ResponseEntity) SetContentType(contentType string) *ResponseEntity {
	r.contentType = contentType
	return r
}

// SetStatusCode sets the transport status; zero defaults to 200 and resolver policy may replace other statuses.
func (r *ResponseEntity) SetStatusCode(statusCode int) *ResponseEntity {
	r.statusCode = statusCode
	return r
}

// AddHeaders copies header values into this entity, appending to existing values.
func (r *ResponseEntity) AddHeaders(headers http.Header) *ResponseEntity {
	for name, values := range headers {
		for _, value := range values {
			r.AddHeader(name, value)
		}
	}
	return r
}

// SetHeader replaces values for a header; an empty name is ignored.
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

// AddHeader appends a header value; an empty name is ignored.
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

// AddCookies borrows nonnil cookie pointers; callers must keep them stable until writing completes.
func (r *ResponseEntity) AddCookies(cookies []*http.Cookie) *ResponseEntity {
	for _, cookie := range cookies {
		if cookie != nil {
			r.cookies = append(r.cookies, cookie)
		}
	}
	return r
}

// AddCookie borrows a cookie pointer; nil is ignored.
func (r *ResponseEntity) AddCookie(cookie *http.Cookie) *ResponseEntity {
	if cookie != nil {
		r.cookies = append(r.cookies, cookie)
	}
	return r
}

// DebugString formats diagnostic data; a negative limit includes all body bytes. Avoid logging sensitive responses.
func (r *ResponseEntity) DebugString(maxBodyBytes int) string {
	body := r.body
	truncated := false
	if maxBodyBytes >= 0 && len(body) > maxBodyBytes {
		body = body[:maxBodyBytes]
		truncated = true
	}
	return fmt.Sprintf("status-code: %d body: %s body-length: %d truncated: %t headers: %v content-type: %s", r.statusCode, string(body), len(r.body), truncated, r.headers, r.contentType)
}

// Body returns copied body bytes, leaving internal storage unchanged.
func (r *ResponseEntity) Body() []byte {
	return bytes.Clone(r.body)
}

// UnsafeRawBody borrows internal bytes for deliberate mutation; callers must provide synchronization.
func (r *ResponseEntity) UnsafeRawBody() []byte {
	return r.body
}

// responseImpl is the shared implementation of common and REST responses.
type responseImpl struct {
	responseEntity *ResponseEntity
}

// Data returns the mutable entity owned by this response; no copy is made.
func (r *responseImpl) Data() *ResponseEntity {
	return r.responseEntity
}

// NewRespRest creates a mutable JSON-marked response; structured bodies use the configured encoder.
func NewRespRest() *responseImpl {
	return &responseImpl{responseEntity: &ResponseEntity{contentType: gin.MIMEJSON}}
}

// NewCommonResp creates a mutable response with no preset media type.
func NewCommonResp() *responseImpl {
	return &responseImpl{responseEntity: &ResponseEntity{}}
}

// DataBuilder replaces the entity; a nil function or nil result panics with a package error.
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

// SetBody encodes structured data and returns the entity; encoding errors panic.
func (r *responseImpl) SetBody(body any) *ResponseEntity {
	bodyBytes, err := currentResponseBodyEncoder().Encode(body)
	if err != nil {
		panic(err)
	}
	r.responseEntity.body = bodyBytes
	return r.responseEntity
}

// SetBodyResponse encodes a structured body and returns the response; encoding errors panic.
func (r *responseImpl) SetBodyResponse(body any) Response {
	bodyBytes, err := currentResponseBodyEncoder().Encode(body)
	if err != nil {
		panic(err)
	}
	r.responseEntity.body = bodyBytes
	return r
}

// SetEntity borrows and replaces the entity, returning it; nil panics with ErrResponseEntityNil.
func (r *responseImpl) SetEntity(entity *ResponseEntity) *ResponseEntity {
	if entity == nil {
		panic(ErrResponseEntityNil)
	}
	r.responseEntity = entity
	return r.responseEntity
}

// SetEntityResponse borrows and replaces the entity, returning the response; nil panics.
func (r *responseImpl) SetEntityResponse(entity *ResponseEntity) Response {
	r.SetEntity(entity)
	return r
}

// RespRestRaw encodes the supplied REST envelope; a nil envelope encodes as null.
func RespRestRaw(dataRest *RestRespStruct) Response {
	return NewRespRest().SetBodyResponse(dataRest)
}

// RespRestSuccess encodes a successful REST envelope using only the first optional data value.
func RespRestSuccess(data ...any) Response {
	return NewRespRest().SetBodyResponse(NewRestSuccess(data...))
}

// RespRestException encodes an exception envelope using only the first optional message.
func RespRestException(statusMessage ...string) Response {
	return NewRespRest().SetBodyResponse(NewRestException(statusMessage...))
}

// RespRestBadParameters encodes a bad-parameters envelope using only the first optional message.
func RespRestBadParameters(statusMessage ...string) Response {
	return NewRespRest().SetBodyResponse(NewRestBadParameters(statusMessage...))
}

// RespRestUnauthorized encodes an unauthorized envelope using only the first optional message.
func RespRestUnauthorized(statusMessage ...string) Response {
	return NewRespRest().SetBodyResponse(NewRestUnauthorized(statusMessage...))
}

// RespRestStatusError encodes a body status, independently of the HTTP transport status.
func RespRestStatusError(statusCode StatusCode, statusMessage ...StatusMessage) Response {
	return NewRespRest().SetBodyResponse(NewRestStatusError(statusCode, statusMessage...))
}

// RespRestBizError encodes a business error under a successful framework status.
func RespRestBizError(bizErrorCode BizErrorCode, bizErrorMessage BizErrorMessage) Response {
	return NewRespRest().SetBodyResponse(NewRestBizError(bizErrorCode, bizErrorMessage))
}

// RespHTTPStatusCode creates an empty response with the given status; automatic wrapping may replace it.
func RespHTTPStatusCode(statusCode int) Response {
	return NewCommonResp().SetEntityResponse(NewEmptyResponseEntity().SetStatusCode(statusCode))
}

// RespJSON encodes a structured body and marks it as JSON; the first optional status defaults to 200.
func RespJSON(body any, httpStatusCode ...int) Response {
	bodyBytes, err := currentResponseBodyEncoder().Encode(body)
	if err != nil {
		panic(err)
	}
	return RespJSONRaw(bodyBytes, httpStatusCode...)
}

// RespJSONRaw borrows encoded bytes without validation; only the first optional HTTP status is used.
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

// RespTextPlain borrows text bytes; only the first optional HTTP status is used, defaulting to 200.
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
