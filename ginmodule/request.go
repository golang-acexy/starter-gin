package ginmodule

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"mime/multipart"
	"net/http"
	"path/filepath"
)

type Request struct {
	ctx *gin.Context
}

// HttpMethod 获取请求方法
func (r *Request) HttpMethod() string {
	return r.ctx.Request.Method
}

// FullPath 获取请求全路径
func (r *Request) FullPath() string {
	return r.ctx.FullPath()
}

// RawGinContext 获取原始Gin上下文
func (r *Request) RawGinContext() *gin.Context {
	return r.ctx
}

// RequestIP 尝试获取请求方客户端IP
func (r *Request) RequestIP() string {
	return r.ctx.ClientIP()
}

// UriPathParam 获取path路径参数 /:id/
func (r *Request) UriPathParam(name string) string {
	return r.ctx.Param(name)
}

// UriPathParams 获取path路径参数 多个参数
func (r *Request) UriPathParams(names ...string) map[string]string {
	result := make(map[string]string, len(names))
	if len(names) > 0 {
		for _, name := range names {
			result[name] = r.UriPathParam(name)
		}
	}
	return result
}

// BindUriPathParams 绑定结构体用于接收UriPath参数
// 任何异常将触发panic响应请求参数错误 `uri:""`
func (r *Request) BindUriPathParams(object any) {
	err := r.ctx.ShouldBindUri(object)
	if err != nil {
		r.ctx.Status(http.StatusBadRequest)
		panic(err)
	}
}

// UriQueryParam 获取 uri Query参数值 /?a=b&c=d
// return string: 参数值(可能是类型零值) bool: 请求方是否传递
func (r *Request) UriQueryParam(name string) (string, bool) {
	return r.ctx.GetQuery(name)
}

// UriQueryParamArray 获取 uri Query参数值 /?a=b&c=d
func (r *Request) UriQueryParamArray(name string) ([]string, bool) {
	return r.ctx.GetQueryArray(name)
}

// UriQueryParamMap 获取 uri Query参数值 /?a=b&c=d 多个参数
func (r *Request) UriQueryParamMap(name string) (map[string]string, bool) {
	return r.ctx.GetQueryMap(name)
}

func (r *Request) UriQueryParams(names ...string) map[string]string {
	result := make(map[string]string, len(names))
	if len(names) > 0 {
		for _, name := range names {
			result[name], _ = r.UriQueryParam(name)
		}
	}
	return result
}

// BindUriQueryParams 绑定结构体用于接收UriQuery参数
// 任何异常将触发panic响应请求参数错误 `form:""`
func (r *Request) BindUriQueryParams(object any) {
	err := r.ctx.ShouldBindQuery(object)
	if err != nil {
		r.ctx.Status(http.StatusBadRequest)
		panic(err)
	}
}

// ShouldBindUriQueryParams 绑定结构体用于接收UriQuery参数
func (r *Request) ShouldBindUriQueryParams(object any) {
	_ = r.ctx.ShouldBindQuery(object)
}

// BindBodyJson 将请求body数据绑定到json结构体中
// 任何异常将触发panic响应请求参数错误 `json:""`
func (r *Request) BindBodyJson(object any) {
	err := r.ctx.ShouldBindJSON(object)
	if err != nil {
		r.ctx.Status(http.StatusBadRequest)
		panic(err)
	}
}

// ShouldBindBodyJson 将请求body数据绑定到json结构体中
func (r *Request) ShouldBindBodyJson(object any) {
	_ = r.ctx.ShouldBindJSON(object)
}

// BindBodyForm 将请求body表单数据绑定到from结构体中
// 任何异常将触发panic响应请求参数错误 `form:""`
func (r *Request) BindBodyForm(object any) {
	err := r.ctx.ShouldBindWith(object, binding.FormPost)
	if err != nil {
		r.ctx.Status(http.StatusBadRequest)
		panic(err)
	}
}

// ShouldBindBodyForm 将请求body表单数据绑定到from结构体中
func (r *Request) ShouldBindBodyForm(object any) {
	_ = r.ctx.ShouldBindWith(object, binding.FormPost)
}

// RawData 将请求body以字节数据返回
// 任何异常将触发panic响应请求参数错误
func (r *Request) RawData() []byte {
	rawData, err := r.ctx.GetRawData()
	if err != nil {
		r.ctx.Status(http.StatusBadRequest)
		panic(err)
	}
	return rawData
}

// RawDataString 将原始请求的body以字符串数据返回
// 任何异常将触发panic响应请求参数错误
func (r *Request) RawDataString() string {
	bytes := r.RawData()
	return string(bytes)
}

// FormFile 获取上传文件内容
// 任何异常将触发panic响应请求参数错误
// request name: form name
func (r *Request) FormFile(name string) *multipart.FileHeader {
	file, err := r.ctx.FormFile(name)
	if err != nil {
		r.ctx.Status(http.StatusBadRequest)
		panic(err)
	}
	return file
}

// SaveUploadedFile 保存上传的文件内容
// 任何异常将触发panic响应请求参数错误
// request	name: form name
//
//	dirPath: 保存的路径 (文件夹)
//	filename: 保存的文件名 若不指定则为源文件名
func (r *Request) SaveUploadedFile(name string, dirPath string, filename ...string) error {
	file := r.FormFile(name)
	var dist string
	if len(filename) != 0 {
		dist = dirPath + string(filepath.Separator) + filename[0]
	} else {
		dist = dirPath + string(filepath.Separator) + file.Filename
	}
	return r.ctx.SaveUploadedFile(file, dist)
}

// HeaderValue 获取Head name对应的参数值
func (r *Request) HeaderValue(key string) string {
	return r.ctx.GetHeader(key)
}
