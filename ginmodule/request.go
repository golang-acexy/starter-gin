package ginmodule

import (
	"github.com/gin-gonic/gin"
	"mime/multipart"
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

func (r *Request) Param(key string) string {
	return r.ctx.Param(key)
}

func (r *Request) Params(keys ...string) map[string]string {
	result := make(map[string]string)
	if len(keys) > 0 {
		for _, value := range keys {
			result[value] = r.Param(value)
		}
	}
	return result
}

func (r *Request) BindUri(object any) error {
	return r.ctx.ShouldBindUri(object)
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
