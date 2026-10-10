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

// Request exposes request-scoped Gin operations; its zero value is not usable.
// Do not retain it after the handler returns or share binding/body operations across goroutines.
type Request struct {
	ctx *gin.Context
}

// GinContext exposes the borrowed native context; native writes own the final response.
// In pre-interceptors, return a Response/error to stop the framework pre-list.
func (r *Request) GinContext() *gin.Context {
	return r.ctx
}

// Method returns the HTTP method.
func (r *Request) Method() string {
	return r.ctx.Request.Method
}

// RoutePattern returns the matched route pattern, or an empty string for unmatched requests.
func (r *Request) RoutePattern() string {
	return r.ctx.FullPath()
}

// RequestPath returns the URL path without query parameters.
func (r *Request) RequestPath() string {
	return r.ctx.Request.URL.Path
}

// RequestURI returns the escaped path and query string.
func (r *Request) RequestURI() string {
	return r.ctx.Request.URL.RequestURI()
}

// Host returns the request Host value, which is supplied by the client.
func (r *Request) Host() string {
	return r.ctx.Request.Host
}

// Protocol returns the HTTP protocol string.
func (r *Request) Protocol() string {
	return r.ctx.Request.Proto
}

// ClientIP delegates to Gin and its proxy trust configuration; forwarded values require trusted proxies.
func (r *Request) ClientIP() string {
	return r.ctx.ClientIP()
}

// Path parameters.

// GetPathParam returns a path parameter, or an empty string when absent.
func (r *Request) GetPathParam(name string) string {
	return r.ctx.Param(name)
}

// GetPathParams returns requested path values; absent names have empty values.
func (r *Request) GetPathParams(names ...string) map[string]string {
	result := make(map[string]string, len(names))
	if len(names) > 0 {
		for _, name := range names {
			result[name] = r.GetPathParam(name)
		}
	}
	return result
}

// BindPathParams binds uri tags into a nonnil destination pointer and applies binding validation.
func (r *Request) BindPathParams(object any) error {
	return r.ctx.ShouldBindUri(object)
}

// MustBindPathParams binds uri tags or interrupts processing with a request error.
func (r *Request) MustBindPathParams(object any) {
	panicRequestError(r.BindPathParams(object), http.StatusBadRequest)
}

// Query parameters.

// GetQueryParam returns the first query value and whether the parameter was supplied, including empty values.
func (r *Request) GetQueryParam(name string) (string, bool) {
	return r.ctx.GetQuery(name)
}

// MustGetQueryParam returns the first supplied query value or interrupts with a missing-parameter error.
func (r *Request) MustGetQueryParam(name string) string {
	v, ok := r.GetQueryParam(name)
	if !ok {
		panicRequestError(paramNotSetError(name), http.StatusBadRequest)
	}
	return v
}

// GetQueryParams returns the first values of supplied names; missing names are omitted.
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

// MustGetQueryParams returns first values or interrupts if any requested name is missing.
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

// GetQueryParamArray returns all query values and presence; empty values are retained.
func (r *Request) GetQueryParamArray(name string) ([]string, bool) {
	return r.ctx.GetQueryArray(name)
}

// MustGetQueryParamArray returns query values or interrupts when absent.
func (r *Request) MustGetQueryParamArray(name string) []string {
	value, ok := r.GetQueryParamArray(name)
	if !ok {
		panicRequestError(paramNotSetError(name), http.StatusBadRequest)
	}
	return value
}

// GetQueryParamMap reads name[key] query values using Gin semantics.
func (r *Request) GetQueryParamMap(name string) (map[string]string, bool) {
	return r.ctx.GetQueryMap(name)
}

// MustGetQueryParamMap reads name[key] query values or interrupts when absent.
func (r *Request) MustGetQueryParamMap(name string) map[string]string {
	v, ok := r.GetQueryParamMap(name)
	if !ok {
		panicRequestError(paramNotSetError(name), http.StatusBadRequest)
	}
	return v
}

// BindQueryParams binds query values into a nonnil destination pointer; it does not read form bodies.
func (r *Request) BindQueryParams(object any) error {
	return r.ctx.ShouldBindQuery(object)
}

// MustBindQueryParams binds query values or interrupts on binding/validation failure.
func (r *Request) MustBindQueryParams(object any) {
	panicRequestError(r.BindQueryParams(object), http.StatusBadRequest)
}

// Body parameters.

// BindBodyJSON decodes cached body bytes with Gin JSON binding and validation.
// Repeated reads are supported; binding may partially mutate the caller-owned destination on error.
func (r *Request) BindBodyJSON(object any) error {
	body, err := r.GetRawBodyData()
	if err != nil {
		return err
	}
	return binding.JSON.BindBody(body, object)
}

// MustBindBodyJSON binds JSON or interrupts with a request error, including 413 for an oversized body.
func (r *Request) MustBindBodyJSON(object any) {
	panicRequestError(r.BindBodyJSON(object), http.StatusBadRequest)
}

// BindBodyForm binds urlencoded or multipart data into a nonnil pointer; multipart file fields are supported.
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

// MustBindBodyForm binds form data or interrupts on parsing, binding, validation, or size-limit failure.
func (r *Request) MustBindBodyForm(object any) {
	panicRequestError(r.BindBodyForm(object), http.StatusBadRequest)
}

// GetRawBodyData returns a copy of cached body bytes and resets the reader after a successful first read.
// Read errors are returned without caching partial data; an absent body returns empty bytes.
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

// BindBodyAuto selects JSON, urlencoded, or multipart from Content-Type; malformed/unsupported media types return an error.
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

// MustBindBodyAuto binds by Content-Type or interrupts; unsupported content maps to 415 and body limits to 413.
func (r *Request) MustBindBodyAuto(object any) {
	panicRequestError(r.BindBodyAuto(object), http.StatusBadRequest)
}

// MustGetRawBodyData returns copied body bytes or interrupts on a read error.
func (r *Request) MustGetRawBodyData() []byte {
	v, err := r.GetRawBodyData()
	panicRequestError(err, http.StatusBadRequest)
	return v
}

// MustGetRawBodyString returns cached body content as a string or interrupts on a read error.
func (r *Request) MustGetRawBodyString() string {
	return conversion.FromBytes(r.MustGetRawBodyData())
}

// GetFormValue returns the first body form value, presence, and any cached form parsing error.
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

// MustGetFormValue returns a body form value or interrupts on parsing failure or absence.
func (r *Request) MustGetFormValue(name string) string {
	v, ok, err := r.GetFormValue(name)
	panicRequestError(err, http.StatusBadRequest)
	if !ok {
		panicRequestError(paramNotSetError(name), http.StatusBadRequest)
	}
	return v
}

// GetFormArray returns body form values without copying; do not mutate them during further request processing.
func (r *Request) GetFormArray(name string) ([]string, bool, error) {
	if err := r.parseForm(); err != nil {
		return nil, false, err
	}
	values, ok := r.ctx.Request.PostForm[name]
	return values, ok, nil
}

// MustGetFormArray returns body form values or interrupts on parsing failure or absence.
func (r *Request) MustGetFormArray(name string) []string {
	v, ok, err := r.GetFormArray(name)
	panicRequestError(err, http.StatusBadRequest)
	if !ok {
		panicRequestError(paramNotSetError(name), http.StatusBadRequest)
	}
	return v
}

// GetFormMap reads name[key] body form values and preserves form parsing errors.
func (r *Request) GetFormMap(name string) (map[string]string, bool, error) {
	if err := r.parseForm(); err != nil {
		return nil, false, err
	}
	value, ok := r.ctx.GetPostFormMap(name)
	return value, ok, nil
}

// MustGetFormMap reads name[key] body form values or interrupts on parsing failure or absence.
func (r *Request) MustGetFormMap(name string) map[string]string {
	v, ok, err := r.GetFormMap(name)
	panicRequestError(err, http.StatusBadRequest)
	if !ok {
		panicRequestError(paramNotSetError(name), http.StatusBadRequest)
	}
	return v
}

// GetFormFile returns an upload header after parsing; callers must close files they open from it.
func (r *Request) GetFormFile(name string) (*multipart.FileHeader, error) {
	if err := r.parseForm(); err != nil {
		return nil, err
	}
	return r.ctx.FormFile(name)
}

// MustGetFormFile returns an upload header or interrupts on parsing failure or absence.
func (r *Request) MustGetFormFile(name string) *multipart.FileHeader {
	v, err := r.GetFormFile(name)
	panicRequestError(err, http.StatusBadRequest)
	return v
}

// SaveUploadedFile copies the named upload to a caller-selected full path, creating parent directories.
// Existing files are overwritten; callers own path validation. Storage errors wrap ErrSaveUploadedFile.
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

// MustSaveUploadedFile saves an upload or interrupts with 400 for invalid input, 413 for body limits,
// or 500 for storage failures. It has the same overwrite behavior as SaveUploadedFile.
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

// parseMediaType normalizes a media type and ignores Content-Type parameters.
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

// parseForm caches parsing results and errors instead of suppressing malformed-form failures.
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

// Panic interrupts processing with an explicit HTTP status and a nonnil error; resolver policy controls final output.
func (r *Request) Panic(statusCode int, err error) {
	panic(&internalPanic{
		statusCode: statusCode,
		rawError:   err,
	})
}

// GetHeader returns the first request header value or an empty string.
func (r *Request) GetHeader(name string) string {
	return r.ctx.GetHeader(name)
}

// GetCookie returns a decoded cookie value or an error when absent or invalid.
func (r *Request) GetCookie(name string) (string, error) {
	return r.ctx.Cookie(name)
}

// MustGetCookie returns a cookie or interrupts on lookup failure.
func (r *Request) MustGetCookie(name string) string {
	v, err := r.ctx.Cookie(name)
	panicRequestError(err, http.StatusBadRequest)
	return v
}

// SetValue stores request-local data; synchronization of mutable stored values remains caller-owned.
func (r *Request) SetValue(key string, value any) {
	r.ctx.Set(key, value)
}

// GetValue returns request-local data and presence; the returned value is not copied.
func (r *Request) GetValue(key string) (any, bool) {
	return r.ctx.Get(key)
}
