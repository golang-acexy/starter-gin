package ginstarter

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/acexy/golang-toolkit/math/conversion"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type Request struct {
	ctx *gin.Context
}

// GinContext 获取原始 Gin 上下文。
func (r *Request) GinContext() *gin.Context {
	return r.ctx
}

// Method 获取 HTTP 请求方法。
func (r *Request) Method() string {
	return r.ctx.Request.Method
}

// RoutePattern 获取当前请求匹配的注册路由模式，例如 /users/:id。
func (r *Request) RoutePattern() string {
	return r.ctx.FullPath()
}

// RequestPath 获取请求路径
func (r *Request) RequestPath() string {
	return r.ctx.Request.URL.Path
}

// RequestURI 获取请求 URI，包含路径和查询参数。
func (r *Request) RequestURI() string {
	return r.ctx.Request.URL.RequestURI()
}

// Host 获取Host信息
func (r *Request) Host() string {
	return r.ctx.Request.Host
}

// Protocol 获取请求协议，例如 HTTP/1.1 或 HTTP/2.0。
func (r *Request) Protocol() string {
	return r.ctx.Request.Proto
}

// ClientIP 尝试获取客户端 IP。
func (r *Request) ClientIP() string {
	return r.ctx.ClientIP()
}

// --------------- path 路径参数

// GetPathParam 获取path路径参数 /:id
func (r *Request) GetPathParam(name string) string {
	return r.ctx.Param(name)
}

// GetPathParams 获取path路径参数 /:id 多个参数
func (r *Request) GetPathParams(names ...string) map[string]string {
	result := make(map[string]string, len(names))
	if len(names) > 0 {
		for _, name := range names {
			result[name] = r.GetPathParam(name)
		}
	}
	return result
}

// BindPathParams /:id 绑定结构体用于接收UriPath参数 结构体标签格式 `uri:""`
func (r *Request) BindPathParams(object any) error {
	return r.ctx.ShouldBindUri(object)
}

// MustBindPathParams /:id 绑定结构体用于接收UriPath参数 结构体标签格式 `uri:""`
// 任何错误将触发Panic流程中断
func (r *Request) MustBindPathParams(object any) {
	panicRequestError(r.BindPathParams(object), http.StatusBadRequest)
}

// --------------- query 参数

// GetQueryParam 获取 uri Query参数值 /?a=b&c=d
// return string: 参数值(可能是类型零值) bool: 请求方是否传递
func (r *Request) GetQueryParam(name string) (string, bool) {
	return r.ctx.GetQuery(name)
}

// MustGetQueryParam 获取 uri Query参数值 /?a=b&c=d 如果没有发送指定参数将触发异常中断流程
func (r *Request) MustGetQueryParam(name string) string {
	v, ok := r.GetQueryParam(name)
	if !ok {
		panicRequestError(paramNotSetError(name), http.StatusBadRequest)
	}
	return v
}

// GetQueryParams 获取 uri Query参数值 /?a=b&c=d 返回map类型数据
// 如果目标没有传递将, map中将不包含指定的参数名
func (r *Request) GetQueryParams(names ...string) map[string]string {
	result := make(map[string]string, len(names))
	if len(names) > 0 {
		for _, name := range names {
			v, ok := r.GetQueryParam(name)
			if ok {
				result[name] = v
			}
		}
	}
	return result
}

// MustGetQueryParams 获取 uri Query参数值 /?a=b&c=d 返回map类型数据
// 任何一个指定的参数没有传递将触发异常中断流程
func (r *Request) MustGetQueryParams(names ...string) map[string]string {
	result := make(map[string]string, len(names))
	if len(names) > 0 {
		for _, name := range names {
			v, ok := r.GetQueryParam(name)
			if ok {
				result[name] = v
			} else {
				panicRequestError(paramNotSetError(name), http.StatusBadRequest)
			}
		}
	}
	return result
}

// GetQueryParamArray 获取 uri Query参数值 /?a=b&a=d 返回切片数据
func (r *Request) GetQueryParamArray(name string) ([]string, bool) {
	return r.ctx.GetQueryArray(name)
}

// MustGetQueryParamArray 获取 uri Query参数值 /?a=b&a=d 返回切片数据
// 如果参数未设置将触发异常中断流程
func (r *Request) MustGetQueryParamArray(name string) []string {
	value, ok := r.GetQueryParamArray(name)
	if !ok {
		panicRequestError(paramNotSetError(name), http.StatusBadRequest)
	}
	return value
}

// GetQueryParamMap 获取 uri Query参数值 /?name[a]=1&name[b]=2 返回map类型数据
func (r *Request) GetQueryParamMap(name string) (map[string]string, bool) {
	return r.ctx.GetQueryMap(name)
}

// MustGetQueryParamMap 获取 uri Query参数值 /?name[a]=1&name[b]=2 返回map类型数据
// 如果参数未设置将触发异常中断流程
func (r *Request) MustGetQueryParamMap(name string) map[string]string {
	v, ok := r.GetQueryParamMap(name)
	if !ok {
		panicRequestError(paramNotSetError(name), http.StatusBadRequest)
	}
	return v
}

// BindQueryParams 绑定结构体用于接收Query参数
func (r *Request) BindQueryParams(object any) error {
	return r.ctx.ShouldBindQuery(object)
}

// MustBindQueryParams 绑定结构体用于接收Query参数以及POST表单符合条件的数据
// 任何错误将触发Panic流程中断
func (r *Request) MustBindQueryParams(object any) {
	panicRequestError(r.BindQueryParams(object), http.StatusBadRequest)
}

// --------------- body 参数

// BindBodyJSON 将请求正文绑定到 JSON 结构体中。
func (r *Request) BindBodyJSON(object any) error {
	body, err := r.GetRawBodyData()
	if err != nil {
		return err
	}
	return binding.JSON.BindBody(body, object)
}

// MustBindBodyJSON 将请求正文绑定到 JSON 结构体中。
// 任何错误将触发Panic流程中断
func (r *Request) MustBindBodyJSON(object any) {
	panicRequestError(r.BindBodyJSON(object), http.StatusBadRequest)
}

// BindBodyForm 将 urlencoded 或 multipart 请求正文绑定到结构体中。
// multipart 绑定支持 *multipart.FileHeader 文件字段。
func (r *Request) BindBodyForm(object any) error {
	mediaType, err := r.requestMediaType()
	if err != nil {
		return err
	}
	return r.bindBodyForm(mediaType, object)
}

func (r *Request) bindBodyForm(mediaType string, object any) error {
	switch mediaType {
	case gin.MIMEPOSTForm:
		if err := r.parseForm(); err != nil {
			return err
		}
		return r.ctx.ShouldBindWith(object, binding.FormPost)
	case gin.MIMEMultipartPOSTForm:
		if err := r.parseForm(); err != nil {
			return err
		}
		return r.ctx.ShouldBindWith(object, binding.FormMultipart)
	default:
		return unsupportedContentTypeError(r.GetHeader("Content-Type"))
	}
}

// MustBindBodyForm 将请求body表单数据绑定到from结构体中
// 任何错误将触发Panic流程中断
func (r *Request) MustBindBodyForm(object any) {
	panicRequestError(r.BindBodyForm(object), http.StatusBadRequest)
}

// GetRawBodyData 将请求body以字节数据返回
func (r *Request) GetRawBodyData() ([]byte, error) {
	if cachedBody, exists := r.ctx.Get(ginCtxKeyRequestBody); exists {
		if body, ok := cachedBody.([]byte); ok {
			return bytes.Clone(body), nil
		}
	}
	if r.ctx.Request.Body == nil {
		return []byte{}, nil
	}
	body, err := io.ReadAll(r.ctx.Request.Body)
	if err != nil {
		return nil, err
	}
	r.ctx.Set(ginCtxKeyRequestBody, bytes.Clone(body))
	r.ctx.Request.Body = io.NopCloser(bytes.NewReader(body))
	return bytes.Clone(body), nil
}

// BindBodyAuto 根据 Content-Type 自动选择 JSON、urlencoded 或 multipart 绑定器。
func (r *Request) BindBodyAuto(object any) error {
	mediaType, err := r.requestMediaType()
	if err != nil {
		return err
	}
	switch mediaType {
	case gin.MIMEJSON:
		return r.BindBodyJSON(object)
	case gin.MIMEPOSTForm, gin.MIMEMultipartPOSTForm:
		return r.bindBodyForm(mediaType, object)
	default:
		return unsupportedContentTypeError(r.GetHeader("Content-Type"))
	}
}

// MustBindBodyAuto 根据 Content-Type 自动绑定请求正文，任何错误都会中断请求流程。
func (r *Request) MustBindBodyAuto(object any) {
	panicRequestError(r.BindBodyAuto(object), http.StatusBadRequest)
}

// MustGetRawBodyData 将请求body以字节数据返回
// 任何错误将触发Panic流程中断
func (r *Request) MustGetRawBodyData() []byte {
	v, err := r.GetRawBodyData()
	panicRequestError(err, http.StatusBadRequest)
	return v
}

// MustGetRawBodyString 将请求body以字符串返回
// 任何错误将触发Panic流程中断
func (r *Request) MustGetRawBodyString() string {
	return conversion.FromBytes(r.MustGetRawBodyData())
}

// GetFormValue 获取 Form 表单值，并返回表单解析错误。
func (r *Request) GetFormValue(name string) (string, bool, error) {
	values, ok, err := r.GetFormArray(name)
	if err != nil || !ok {
		return "", false, err
	}
	if len(values) == 0 {
		return "", false, nil
	}
	return values[0], true, nil
}

// MustGetFormValue 获取Form表单的值
// 任何错误将触发Panic流程中断
func (r *Request) MustGetFormValue(name string) string {
	v, ok, err := r.GetFormValue(name)
	panicRequestError(err, http.StatusBadRequest)
	if !ok {
		panicRequestError(paramNotSetError(name), http.StatusBadRequest)
	}
	return v
}

// GetFormArray 获取 Form 表单的多个值，并返回表单解析错误。
func (r *Request) GetFormArray(name string) ([]string, bool, error) {
	if err := r.parseForm(); err != nil {
		return nil, false, err
	}
	values, ok := r.ctx.Request.PostForm[name]
	return values, ok, nil
}

// MustGetFormArray 获取Form表单的值
// 任何错误将触发Panic流程中断
func (r *Request) MustGetFormArray(name string) []string {
	v, ok, err := r.GetFormArray(name)
	panicRequestError(err, http.StatusBadRequest)
	if !ok {
		panicRequestError(paramNotSetError(name), http.StatusBadRequest)
	}
	return v
}

// GetFormMap 获取使用 name[key] 形式提交的 Form 表单值，并返回表单解析错误。
func (r *Request) GetFormMap(name string) (map[string]string, bool, error) {
	if err := r.parseForm(); err != nil {
		return nil, false, err
	}
	value, ok := r.ctx.GetPostFormMap(name)
	return value, ok, nil
}

// MustGetFormMap 获取Form表单的值
// 任何错误将触发Panic流程中断
func (r *Request) MustGetFormMap(name string) map[string]string {
	v, ok, err := r.GetFormMap(name)
	panicRequestError(err, http.StatusBadRequest)
	if !ok {
		panicRequestError(paramNotSetError(name), http.StatusBadRequest)
	}
	return v
}

// GetFormFile 获取上传文件内容
func (r *Request) GetFormFile(name string) (*multipart.FileHeader, error) {
	if err := r.parseForm(); err != nil {
		return nil, err
	}
	return r.ctx.FormFile(name)
}

// MustGetFormFile 获取上传文件内容
// 任何错误将触发Panic流程中断
func (r *Request) MustGetFormFile(name string) *multipart.FileHeader {
	v, err := r.GetFormFile(name)
	panicRequestError(err, http.StatusBadRequest)
	return v
}

// SaveUploadedFile 将 multipart 中 fieldName 对应的上传文件保存到 targetPath。
// targetPath 是包含文件名的完整目标路径，由调用方负责生成和校验。
// 父目录不存在时会自动创建；目标文件已存在时会被覆盖。
// 文件解析错误直接返回，文件打开、目录创建或写入错误包装为 ErrSaveUploadedFile。
func (r *Request) SaveUploadedFile(fieldName, targetPath string) error {
	file, err := r.GetFormFile(fieldName)
	if err != nil {
		return err
	}
	if err = r.ctx.SaveUploadedFile(file, targetPath); err != nil {
		return fmt.Errorf("%w: %w", ErrSaveUploadedFile, err)
	}
	return nil
}

// MustSaveUploadedFile 将 multipart 中 fieldName 对应的上传文件保存到 targetPath。
// 保存行为与 SaveUploadedFile 相同。文件参数或 multipart 格式错误时以 400 中断请求，
// 请求正文超限时以 413 中断请求，文件打开、目录创建或写入失败时以 500 中断请求。
func (r *Request) MustSaveUploadedFile(fieldName, targetPath string) {
	panicRequestError(r.SaveUploadedFile(fieldName, targetPath), http.StatusBadRequest)
}

func requestErrorStatus(err error, defaultStatus int) int {
	var maxBytesError *http.MaxBytesError
	if errors.As(err, &maxBytesError) {
		return http.StatusRequestEntityTooLarge
	}
	if errors.Is(err, ErrUnsupportedContent) {
		return http.StatusUnsupportedMediaType
	}
	if errors.Is(err, ErrSaveUploadedFile) {
		return http.StatusInternalServerError
	}
	return defaultStatus
}

func panicRequestError(err error, defaultStatus int) {
	if err == nil {
		return
	}
	panic(&internalPanic{
		statusCode: requestErrorStatus(err, defaultStatus),
		rawError:   err,
	})
}

func paramNotSetError(name string) error {
	return fmt.Errorf("%w: %s", ErrParamNotSet, name)
}

func unsupportedContentTypeError(contentType string) error {
	return fmt.Errorf("%w: %s", ErrUnsupportedContent, contentType)
}

// parseMediaType 使用标准库解析并规范化媒体类型，忽略 Content-Type 参数。
func parseMediaType(contentType string) (string, error) {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", err
	}
	return strings.ToLower(mediaType), nil
}

func (r *Request) requestMediaType() (string, error) {
	contentType := r.GetHeader("Content-Type")
	mediaType, err := parseMediaType(contentType)
	if err != nil {
		return "", unsupportedContentTypeError(contentType)
	}
	return mediaType, nil
}

// parseForm 只解析一次表单并保留错误，避免 Gin 的 GetPostForm 系列方法吞掉解析异常。
func (r *Request) parseForm() error {
	state := getRequestState(r.ctx)
	if state.formParsed {
		return state.formErr
	}
	state.formParsed = true
	contentType := r.GetHeader("Content-Type")
	mediaType, mediaTypeErr := parseMediaType(contentType)
	var err error
	if mediaType == gin.MIMEMultipartPOSTForm {
		maxMemory := int64(32 << 20)
		if engine := RawGinEngine(); engine != nil {
			maxMemory = engine.MaxMultipartMemory
		}
		err = r.ctx.Request.ParseMultipartForm(maxMemory)
	} else if mediaTypeErr != nil && strings.HasPrefix(strings.ToLower(strings.TrimSpace(contentType)), "multipart/") {
		err = mediaTypeErr
	} else {
		err = r.ctx.Request.ParseForm()
	}
	state.formErr = err
	return err
}

// Panic 抛出异常，中断业务
func (r *Request) Panic(statusCode int, err error) {
	panic(&internalPanic{
		statusCode: statusCode,
		rawError:   err,
	})
}

// GetHeader 获取Head name对应的参数值
func (r *Request) GetHeader(name string) string {
	return r.ctx.GetHeader(name)
}

// GetCookie 获取Cookie name对应的参数值
func (r *Request) GetCookie(name string) (string, error) {
	return r.ctx.Cookie(name)
}

// MustGetCookie 获取Cookie name对应的参数值
// 任何错误将触发Panic流程中断
func (r *Request) MustGetCookie(name string) string {
	v, err := r.ctx.Cookie(name)
	panicRequestError(err, http.StatusBadRequest)
	return v
}

// SetValue 向gin上下文绑定数据
func (r *Request) SetValue(key string, value any) {
	r.ctx.Set(key, value)
}

// GetValue 从gin上下文获取数据
func (r *Request) GetValue(key string) (any, bool) {
	return r.ctx.Get(key)
}
