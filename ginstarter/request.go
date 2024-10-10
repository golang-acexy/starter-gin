package ginstarter

import (
	"errors"
	"github.com/acexy/golang-toolkit/math/conversion"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"mime/multipart"
	"net/http"
	"path/filepath"
)

type Request struct {
	ctx *gin.Context
}

// RawGinContext 获取原始Gin上下文
func (r *Request) RawGinContext() *gin.Context {
	return r.ctx
}

// HttpMethod 获取请求方法
func (r *Request) HttpMethod() string {
	return r.ctx.Request.Method
}

// RouterFullPath 当前请求的注册路由路径
func (r *Request) RouterFullPath() string {
	return r.ctx.FullPath()
}

// RequestPath 获取请求路径
func (r *Request) RequestPath() string {
	return r.ctx.Request.URL.Path
}

// RequestFullPath 获取请求完整路径
func (r *Request) RequestFullPath() string {
	return r.ctx.Request.URL.RequestURI()
}

// Host 获取Host信息
func (r *Request) Host() string {
	return r.ctx.Request.Host
}

// Proto 获取请求协议
func (r *Request) Proto() string {
	return r.ctx.Request.Proto
}

// RequestIP 尝试获取请求方客户端IP
func (r *Request) RequestIP() string {
	return r.ctx.ClientIP()
}

// --------------- path 路径参数

// GetPathParam 获取path路径参数 /:id/
func (r *Request) GetPathParam(name string) string {
	return r.ctx.Param(name)
}

// GetPathParams 获取path路径参数 /:id/ 多个参数
func (r *Request) GetPathParams(names ...string) map[string]string {
	result := make(map[string]string, len(names))
	if len(names) > 0 {
		for _, name := range names {
			result[name] = r.GetPathParam(name)
		}
	}
	return result
}

// BindPathParams /:id/ 绑定结构体用于接收UriPath参数 结构体标签格式 `uri:""`
func (r *Request) BindPathParams(object any) error {
	return r.ctx.ShouldBindUri(object)
}

// MustBindPathParams /:id/ 绑定结构体用于接收UriPath参数 结构体标签格式 `uri:""`
// 任何错误将触发Panic流程中断
func (r *Request) MustBindPathParams(object any) {
	err := r.BindPathParams(object)
	if err != nil {
		panic(&internalPanic{
			rawError:   err,
			statusCode: http.StatusBadRequest,
		})
	}
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
		panic(&internalPanic{
			statusCode: http.StatusBadRequest,
			rawError:   errors.New("param name = " + name + " not set"),
		})
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
				panic(&internalPanic{
					statusCode: http.StatusBadRequest,
					rawError:   errors.New("param name = " + name + " not set"),
				})
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
		panic(&internalPanic{
			statusCode: http.StatusBadRequest,
			rawError:   errors.New("param name = " + name + " not set"),
		})
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
		panic(&internalPanic{
			statusCode: http.StatusBadRequest,
			rawError:   errors.New("param name = " + name + " not set"),
		})
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
	err := r.BindQueryParams(object)
	if err != nil {
		panic(&internalPanic{
			statusCode: http.StatusBadRequest,
			rawError:   err,
		})
	}
}

// --------------- body 参数

// BindBodyJson 将请求body数据绑定到json结构体中
func (r *Request) BindBodyJson(object any) error {
	return r.ctx.ShouldBindJSON(object)
}

// MustBindBodyJson 将请求body数据绑定到json结构体中
// 任何错误将触发Panic流程中断
func (r *Request) MustBindBodyJson(object any) {
	err := r.BindBodyJson(object)
	if err != nil {
		panic(&internalPanic{
			statusCode: http.StatusBadRequest,
			rawError:   err,
		})
	}
}

// BindBodyForm 将请求body表单数据绑定到from结构体中
func (r *Request) BindBodyForm(object any) error {
	return r.ctx.ShouldBindWith(object, binding.FormPost)
}

// MustBindBodyForm 将请求body表单数据绑定到from结构体中
// 任何错误将触发Panic流程中断
func (r *Request) MustBindBodyForm(object any) {
	err := r.BindBodyForm(object)
	if err != nil {
		panic(&internalPanic{
			statusCode: http.StatusBadRequest,
			rawError:   err,
		})
	}
}

// GetRawBodyData 将请求body以字节数据返回
func (r *Request) GetRawBodyData() ([]byte, error) {
	return r.ctx.GetRawData()
}

// MustGetRawBodyData 将请求body以字节数据返回
// 任何错误将触发Panic流程中断
func (r *Request) MustGetRawBodyData() []byte {
	v, err := r.GetRawBodyData()
	if err != nil {
		panic(&internalPanic{
			statusCode: http.StatusBadRequest,
			rawError:   err,
		})
	}
	return v
}

// MustGetRawBodyString 将请求body以字符串返回
// 任何错误将触发Panic流程中断
func (r *Request) MustGetRawBodyString() string {
	return conversion.FromBytes(r.MustGetRawBodyData())
}

// GetFormValue 获取Form表单的值
func (r *Request) GetFormValue(name string) (string, bool) {
	return r.ctx.GetPostForm(name)
}

// MustGetFormValue 获取Form表单的值
// 任何错误将触发Panic流程中断
func (r *Request) MustGetFormValue(name string) string {
	v, ok := r.GetFormValue(name)
	if !ok {
		panic(&internalPanic{
			statusCode: http.StatusBadRequest,
			rawError:   errors.New("param name = " + name + " not set"),
		})
	}
	return v
}

// GetFormArray 获取Form表单的值
func (r *Request) GetFormArray(name string) ([]string, bool) {
	return r.ctx.GetPostFormArray(name)
}

// MustGetFormArray 获取Form表单的值
// 任何错误将触发Panic流程中断
func (r *Request) MustGetFormArray(name string) []string {
	v, ok := r.GetFormArray(name)
	if !ok {
		panic(&internalPanic{
			statusCode: http.StatusBadRequest,
			rawError:   errors.New("param name = " + name + " not set"),
		})
	}
	return v
}

// GetFormMap 获取Form表单的值
func (r *Request) GetFormMap(name string) (map[string]string, bool) {
	return r.ctx.GetPostFormMap(name)
}

// MustGetFormMap 获取Form表单的值
// 任何错误将触发Panic流程中断
func (r *Request) MustGetFormMap(name string) map[string]string {
	v, ok := r.GetFormMap(name)
	if !ok {
		panic(&internalPanic{
			statusCode: http.StatusBadRequest,
			rawError:   errors.New("param name = " + name + " not set"),
		})
	}
	return v
}

// GetFormFile 获取上传文件内容
func (r *Request) GetFormFile(name string) (*multipart.FileHeader, error) {
	return r.ctx.FormFile(name)
}

// MustGetFormFile 获取上传文件内容
// 任何错误将触发Panic流程中断
func (r *Request) MustGetFormFile(name string) *multipart.FileHeader {
	v, err := r.ctx.FormFile(name)
	if err != nil {
		panic(&internalPanic{
			statusCode: http.StatusBadRequest,
			rawError:   err,
		})
	}
	return v
}

// SaveUploadedFile 保存上传的文件内容 name: form name dirPath: 保存的路径 (文件夹) filename: 保存的文件名 若不指定则为源文件名
func (r *Request) SaveUploadedFile(name string, dirPath string, filename ...string) error {
	file, err := r.GetFormFile(name)
	if err != nil {
		return err
	}
	var dist string
	if len(filename) != 0 {
		dist = dirPath + string(filepath.Separator) + filename[0]
	} else {
		dist = dirPath + string(filepath.Separator) + file.Filename
	}
	return r.ctx.SaveUploadedFile(file, dist)
}

// MustSaveUploadedFile 保存上传的文件内容 name: form name dirPath: 保存的路径 (文件夹) filename: 保存的文件名 若不指定则为源文件名
// 任何错误将触发Panic流程中断
func (r *Request) MustSaveUploadedFile(name string, dirPath string, filename ...string) {
	err := r.SaveUploadedFile(name, dirPath, filename...)
	if err != nil {
		panic(&internalPanic{
			statusCode: http.StatusBadRequest,
			rawError:   err,
		})
	}
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
	if err != nil {
		panic(&internalPanic{
			statusCode: http.StatusBadRequest,
			rawError:   err,
		})
	}
	return v
}

// SetValue 向gin上下文绑定数据
func (r *Request) SetValue(key string, value interface{}) {
	r.ctx.Set(key, value)
}

// GetValue 从gin上下文获取数据
func (r *Request) GetValue(key string) (interface{}, bool) {
	return r.ctx.Get(key)
}
