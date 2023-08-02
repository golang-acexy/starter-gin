package ginmodule

import (
	"github.com/gin-gonic/gin"
	"mime/multipart"
	"net/http"
	"path/filepath"
)

type Request struct {
	ctx *gin.Context
}

func (r *Request) FullPath() string {
	return r.ctx.FullPath()
}

func (r *Request) GinContext() *gin.Context {
	return r.ctx
}

func (r *Request) RequestIP() string {
	return r.ctx.ClientIP()
}

// UriPathParam 获取path路径参数 /:id/
func (r *Request) UriPathParam(key string) string {
	return r.ctx.Param(key)
}

func (r *Request) UriPathParams(keys ...string) map[string]string {
	result := make(map[string]string)
	if len(keys) > 0 {
		for _, value := range keys {
			result[value] = r.UriPathParam(value)
		}
	}
	return result
}

// BindUriPathParams 绑定结构体用于接收UriPath参数
// 任何异常将触发panic响应请求参数错误
func (r *Request) BindUriPathParams(object any) {
	err := r.ctx.ShouldBindUri(object)
	if err != nil {
		r.ctx.Status(http.StatusBadRequest)
		panic(err)
	}
}

func (r *Request) BindJson(object any) error {
	return r.ctx.ShouldBindJSON(object)
}

func (r *Request) RawData() ([]byte, error) {
	return r.ctx.GetRawData()
}

func (r *Request) RawDataString() (content string, err error) {
	bytes, err := r.RawData()
	if err != nil {
		return
	}
	content = string(bytes)
	return
}

func (r *Request) FormFile(name string) (*multipart.FileHeader, error) {
	return r.ctx.FormFile(name)
}

func (r *Request) SaveUploadedFile(name string, dirPath string, fileName ...string) error {
	file, err := r.FormFile(name)
	if err != nil {
		return err
	}
	var dist string
	if len(fileName) != 0 {
		dist = dirPath + string(filepath.Separator) + fileName[0]
	} else {
		dist = dirPath + string(filepath.Separator) + file.Filename
	}
	return r.ctx.SaveUploadedFile(file, dist)
}

func (r *Request) HeaderValue(key string) string {
	return r.ctx.GetHeader(key)
}
