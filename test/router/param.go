package router

import (
	"mime/multipart"

	"github.com/acexy/golang-toolkit/logger"
	"github.com/golang-acexy/starter-gin/ginstarter"
)

const paramTestPathPrefix = "/param-test/"

// ParamRouter 验证 path、query、header、cookie、JSON、Form、multipart 和原始正文解析。
// 每个 Handler 都会输出带有 [param-test] 前缀的场景日志，便于同时核对响应与服务端解析结果。
//
// 启动方式（分别启动，不要同时占用 8080 端口）：
//
// go test ./test -run '^TestGinDefault$' -count=1 -v
// go test ./test -run '^TestGinCustomer$' -count=1 -v
//
// 正常场景：
//
// curl -i 'http://127.0.0.1:8080/param-test/path/12/acexy'
// curl --globoff -i 'http://127.0.0.1:8080/param-test/query?id=12&name=acexy&tag=go&tag=gin&filter[role]=admin'
// curl -i 'http://127.0.0.1:8080/param-test/metadata' -H 'X-Request-ID: request-1' --cookie 'session=session-1'
// curl -i 'http://127.0.0.1:8080/param-test/body/json' -H 'Content-Type: application/json; charset=utf-8' --data '{"id":12,"name":"acexy"}'
// curl -i 'http://127.0.0.1:8080/param-test/body/auto' -H 'Content-Type: application/json' --data '{"id":12,"name":"acexy"}'
// curl -i 'http://127.0.0.1:8080/param-test/body/auto' -H 'Content-Type: application/x-www-form-urlencoded' --data 'id=12&name=acexy'
// curl -i 'http://127.0.0.1:8080/param-test/body/form' -H 'Content-Type: application/x-www-form-urlencoded' --data 'id=12&name=acexy'
// curl -i 'http://127.0.0.1:8080/param-test/body/form-values' -H 'Content-Type: application/x-www-form-urlencoded' --data 'name=acexy&tag=go&tag=gin&filter[role]=admin'
// curl -i 'http://127.0.0.1:8080/param-test/body/multipart' -F 'id=12' -F 'name=acexy' -F 'file=@ginstarter/error.go;type=text/plain'
// curl -i 'http://127.0.0.1:8080/param-test/body/raw-repeat' -H 'Content-Type: application/json' --data '{"id":12,"name":"acexy"}'
//
// 异常场景：
//
// curl -i 'http://127.0.0.1:8080/param-test/path/not-number/acexy'
// curl -i 'http://127.0.0.1:8080/param-test/query?name=acexy'
// curl -i 'http://127.0.0.1:8080/param-test/metadata' -H 'X-Request-ID: request-1'
// curl -i 'http://127.0.0.1:8080/param-test/body/json' -H 'Content-Type: application/json' --data '{bad-json}'
// curl -i 'http://127.0.0.1:8080/param-test/body/json' -H 'Content-Type: application/json' --data '{"id":"wrong","name":"acexy"}'
// curl -i 'http://127.0.0.1:8080/param-test/body/auto' -H 'Content-Type: text/plain' --data 'unsupported'
// curl -i 'http://127.0.0.1:8080/param-test/body/form-values' -H 'Content-Type: application/x-www-form-urlencoded' --data 'name=%ZZ&tag=go&filter[role]=admin'
// curl -i 'http://127.0.0.1:8080/param-test/body/multipart' -H 'Content-Type: multipart/form-data' --data 'missing-boundary'
// curl -i 'http://127.0.0.1:8080/param-test/body/json' -H 'Content-Type: application/json' --data-binary @README.md
// curl -i 'http://127.0.0.1:8080/param-test/body/form-values' -H 'Content-Type: application/x-www-form-urlencoded' --data-binary @README.md
type ParamRouter struct{}

func (d *ParamRouter) Info() *ginstarter.RouterInfo {
	return &ginstarter.RouterInfo{
		GroupPath: "param-test",
	}
}

func (d *ParamRouter) Handlers(router *ginstarter.RouterWrapper) {
	router.GET("path/:id/:name", d.path())
	router.GET("query", d.query())
	router.GET("metadata", d.metadata())
	router.POST("body/json", d.json())
	router.POST("body/auto", d.auto())
	router.POST("body/form", d.form())
	router.POST("body/form-values", d.formValues())
	router.POST("body/multipart", d.multipart())
	router.POST("body/raw-repeat", d.rawRepeat())
}

type paramPathInput struct {
	ID   uint   `uri:"id" binding:"required"`
	Name string `uri:"name"`
}

type paramQueryInput struct {
	ID   uint   `form:"id" binding:"required"`
	Name string `form:"name" binding:"required"`
}

type paramBodyInput struct {
	ID   uint   `json:"id" form:"id" binding:"required"`
	Name string `json:"name" form:"name" binding:"required"`
}

type paramMultipartInput struct {
	ID   uint                  `form:"id" binding:"required"`
	Name string                `form:"name" binding:"required"`
	File *multipart.FileHeader `form:"file" binding:"required"`
}

func (d *ParamRouter) path() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		input := paramPathInput{}
		request.MustBindPathParams(&input)
		result := map[string]any{
			"bound": input,
			"raw":   request.GetPathParams("id", "name", "unknown"),
		}
		logParamResult(request, "path", result)
		return ginstarter.RespRestSuccess(result), nil
	}
}

func (d *ParamRouter) query() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		input := paramQueryInput{}
		request.MustBindQueryParams(&input)
		tags, _ := request.GetQueryParamArray("tag")
		filters, _ := request.GetQueryParamMap("filter")
		result := map[string]any{
			"bound":   input,
			"tags":    tags,
			"filters": filters,
		}
		logParamResult(request, "query", result)
		return ginstarter.RespRestSuccess(result), nil
	}
}

func (d *ParamRouter) metadata() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		result := map[string]any{
			"requestId": request.GetHeader("X-Request-ID"),
			"session":   request.MustGetCookie("session"),
			"method":    request.Method(),
			"path":      request.RequestPath(),
			"clientIp":  request.ClientIP(),
		}
		logParamResult(request, "metadata", result)
		return ginstarter.RespRestSuccess(result), nil
	}
}

func (d *ParamRouter) json() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		input := paramBodyInput{}
		request.MustBindBodyJSON(&input)
		logParamResult(request, "json", input)
		return ginstarter.RespRestSuccess(input), nil
	}
}

func (d *ParamRouter) auto() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		input := paramBodyInput{}
		request.MustBindBodyAuto(&input)
		logParamResult(request, "auto", input)
		return ginstarter.RespRestSuccess(input), nil
	}
}

func (d *ParamRouter) form() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		input := paramBodyInput{}
		request.MustBindBodyForm(&input)
		logParamResult(request, "form", input)
		return ginstarter.RespRestSuccess(input), nil
	}
}

func (d *ParamRouter) formValues() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		name := request.MustGetFormValue("name")
		tags := request.MustGetFormArray("tag")
		filters := request.MustGetFormMap("filter")
		result := map[string]any{
			"name":    name,
			"tags":    tags,
			"filters": filters,
		}
		logParamResult(request, "form-values", result)
		return ginstarter.RespRestSuccess(result), nil
	}
}

func (d *ParamRouter) multipart() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		input := paramMultipartInput{}
		request.MustBindBodyAuto(&input)
		result := map[string]any{
			"id":       input.ID,
			"name":     input.Name,
			"filename": input.File.Filename,
			"size":     input.File.Size,
		}
		logParamResult(request, "multipart", result)
		return ginstarter.RespRestSuccess(result), nil
	}
}

func (d *ParamRouter) rawRepeat() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		first := request.MustGetRawBodyData()
		second := request.MustGetRawBodyData()
		input := paramBodyInput{}
		request.MustBindBodyJSON(&input)
		result := map[string]any{
			"same":       string(first) == string(second),
			"firstSize":  len(first),
			"secondSize": len(second),
			"bound":      input,
		}
		logParamResult(request, "raw-repeat", result)
		return ginstarter.RespRestSuccess(result), nil
	}
}

func logParamResult(request *ginstarter.Request, testCase string, value any) {
	logger.Logrus().Infof(
		"[param-test] case=%s method=%s path=%s result=%+v",
		testCase,
		request.Method(),
		request.RequestPath(),
		value,
	)
}
