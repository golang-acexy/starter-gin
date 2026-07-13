# starter-gin

**starter-gin** is the HTTP server starter for the golang-acexy starter/cloud ecosystem. It wraps Gin with parent-managed lifecycle, structured routing, request binding, interceptor chains, panic recovery, and a single framework response pipeline.

## Requirements

Current module Go version: **1.25.8**.

~~~bash
go get github.com/golang-acexy/starter-gin
~~~

## Starter Usage

~~~go
starter := &ginstarter.GinStarter{
    Config: ginstarter.GinConfig{
        ListenAddress: ":8080",
        Routers: []ginstarter.Router{
            &UserRouter{},
        },
        MaxRequestBodyBytes: 8 << 20,
    },
}

loader := parent.InitStarterLoader([]parent.Starter{starter})
if err := loader.Start(); err != nil {
    panic(err)
}
~~~

The starter is singleton-based. Starting another **GinStarter** while one server is already running returns **ErrGinStarterAlreadyStarted**. Use the parent loader to coordinate startup and shutdown.

Use **LazyConfig** when configuration must be resolved at startup time. It takes precedence over the static **Config** value.

## Register Routes

A router defines one Gin route group and its optional pre/post interceptors.

~~~go
type UserRouter struct{}

func (r *UserRouter) Info() *ginstarter.RouterInfo {
    return &ginstarter.RouterInfo{
        GroupPath: "/users",
    }
}

func (r *UserRouter) Handlers(router *ginstarter.RouterWrapper) {
    router.GET("/:id", r.get())
    router.POST1("", []string{gin.MIMEJSON}, r.create())
}

type userPath struct {
    ID uint `uri:"id" binding:"required"`
}

type userCreate struct {
    Name string `json:"name" binding:"required"`
}

func (r *UserRouter) get() ginstarter.HandlerWrapper {
    return func(request *ginstarter.Request) (ginstarter.Response, error) {
        input := userPath{}
        request.MustBindPathParams(&input)
        return ginstarter.RespRestSuccess(input), nil
    }
}

func (r *UserRouter) create() ginstarter.HandlerWrapper {
    return func(request *ginstarter.Request) (ginstarter.Response, error) {
        input := userCreate{}
        request.MustBindBodyJSON(&input)
        return ginstarter.RespRestSuccess(input), nil
    }
}
~~~

**RouterWrapper** supports **GET**, **POST**, **PUT**, **PATCH**, **DELETE**, **HEAD**, **OPTIONS**, **TRACE**, and **MATCH**. Methods ending in **1**, such as **POST1**, additionally restrict the request media type. Media types are parsed with **mime.ParseMediaType**, so parameters such as **charset=utf-8** are accepted while malformed values are rejected.

A handler has one response contract:

~~~go
type HandlerWrapper func(request *Request) (Response, error)
~~~

Returning a non-nil response stops the remaining handlers for that route. Returning an error enters the framework recovery pipeline.

## Request Binding

Request metadata:

- **Method**, **RoutePattern**, **RequestPath**, and **RequestURI**
- **Host**, **Protocol**, and **ClientIP**
- **GetHeader**, **GetCookie**, and **MustGetCookie**
- **GinContext** for integrations that genuinely require the raw Gin context

Path and query binding:

~~~go
type queryInput struct {
    ID   uint   `form:"id" binding:"required"`
    Name string `form:"name" binding:"required"`
}

func handle(request *ginstarter.Request) {
    id := request.GetPathParam("id")

    input := queryInput{}
    request.MustBindQueryParams(&input)

    tags, exists := request.GetQueryParamArray("tag")
    _ = id
    _ = tags
    _ = exists
}
~~~

Body binding:

~~~go
type bodyInput struct {
    ID   uint   `json:"id" form:"id" binding:"required"`
    Name string `json:"name" form:"name" binding:"required"`
}

input := bodyInput{}
request.MustBindBodyAuto(&input)
~~~

**BindBodyAuto** and **MustBindBodyAuto** support:

- **application/json**
- **application/x-www-form-urlencoded**
- **multipart/form-data**

Use **BindBodyJSON** or **BindBodyForm** when the expected representation is already known. JSON raw-body reads are cached, so **GetRawBodyData** can be called repeatedly before or after JSON binding.

Form getters return parsing errors instead of silently dropping malformed form data:

~~~go
value, exists, err := request.GetFormValue("name")
values, exists, err := request.GetFormArray("tag")
mapping, exists, err := request.GetFormMap("filter")
~~~

Methods prefixed with **Must** stop request processing through the framework panic pipeline. Request binding errors normally map to HTTP 400, unsupported media types to 415, and bodies exceeding **MaxRequestBodyBytes** to 413.

### File Uploads

~~~go
targetPath := filepath.Join(uploadDir, "user-42-upload.bin")
if err := request.SaveUploadedFile("file", targetPath); err != nil {
    return nil, err
}
~~~

**targetPath** is the complete destination path and is owned by the caller. Parent directories are created automatically and an existing target file is overwritten. Use **GetFormFile** and custom storage logic when overwrite prevention, atomic writes, object storage, or content validation is required.

## Validation

Gin validation uses the **binding** struct tag:

~~~go
type input struct {
    Email  string `binding:"required,email"`
    Domain string `binding:"required,domain"`
}
~~~

The starter registers the custom **domain** validation tag during startup. Validation failures are converted into concise request messages by the recovery pipeline.

## Responses

Every framework-managed response implements:

~~~go
type Response interface {
    Data() *ResponseEntity
}
~~~

Common helpers:

~~~go
return ginstarter.RespRestSuccess(data), nil
return ginstarter.RespRestBadParameters("invalid request"), nil
return ginstarter.RespRestUnauthorized(), nil
return ginstarter.RespRestBizError(code, message), nil

return ginstarter.RespJSON(data, http.StatusCreated), nil
return ginstarter.RespJSONRaw(encodedJSON), nil
return ginstarter.RespTextPlain([]byte("ok")), nil
return ginstarter.RespHTTPStatusCode(http.StatusNoContent), nil
~~~

Use **ResponseEntity** when status, headers, cookies, content type, and raw body must be controlled together:

~~~go
entity := ginstarter.NewResponseEntityWithStatusCode(
    gin.MIMEJSON,
    []byte("{\"created\":true}"),
    http.StatusCreated,
).
    SetHeader("Location", "/users/1").
    AddCookie(ginstarter.NewCookie("session", "value", 3600, "/", "", true, true))

return ginstarter.NewCommonResp().SetEntityResponse(entity), nil
~~~

**Body** returns a copy. **UnsafeRawBody** exposes the internal byte slice only for post-interceptors that intentionally need in-place mutation. **DebugString** provides bounded diagnostic output.

Structured bodies are encoded through **ResponseBodyEncoder**. The default encoder uses JSON and can be replaced in **GinConfig**.

## Interceptors and Response Pipeline

Global and route interceptors use the same contracts:

~~~go
type PreInterceptor func(request *Request) (Response, error)
type PostInterceptor func(request *Request, response Response) (Response, error)
~~~

Normal execution order:

~~~text
Global pre interceptors
  -> Route pre interceptors
  -> Handler
  -> Route post interceptors
  -> Global post interceptors
  -> Final response write
~~~

Rules:

- Interceptors execute in slice registration order.
- A pre-interceptor returning a response or error stops later pre-interceptors and the handler.
- Post-interceptors still receive the current response after handler errors or pre-interceptor short-circuiting within their scope.
- A post-interceptor returning a non-nil response replaces the current response.
- A post-interceptor returning an error stops the remaining post-interceptors in that interceptor chain.
- The framework writes the final **Response** once, after all applicable post-interceptors finish.

If application code writes through **GinContext**, the framework detects the committed native response and skips its own final response. Setting headers alone does not commit a native response, so those headers remain available to the framework response.

Built-in interceptors:

- **BasicAuthInterceptor**
- **MediaTypeInterceptor**

## Error Handling

By default, non-ignored error HTTP statuses are converted through **BadHTTPCodeResolver** into the standard REST envelope. The default resolver returns transport HTTP 200 and places the original status in **status.statusCode**.

| Configuration | Behavior |
| --- | --- |
| **DisableBadHTTPCodeResolver: false** | Convert non-ignored error statuses to the REST envelope |
| **DisableBadHTTPCodeResolver: true** | Preserve the original HTTP status and response |
| **HidePanicErrorDetails: true** | Hide unknown panic details while preserving explicitly assigned HTTP status codes |
| **IgnoreHTTPCode** | Preserve selected HTTP statuses without resolver conversion |

Expected 4xx request errors are logged at WARN without a stack trace. Unknown failures and 5xx panics are logged at ERROR with diagnostic stack information.

Use **Request.Panic(statusCode, err)** to stop processing with an explicit HTTP status. Use **PanicResolver** to customize visible error text and **BadHTTPCodeResolver** to customize the final error response.

## Configuration

| Field | Purpose |
| --- | --- |
| **ListenAddress** | HTTP listen address; defaults to **:8080** |
| **Routers** | Business router registrations |
| **InitFunc** | Receives the raw **gin.Engine** after starter initialization |
| **UseReusePortModel** | Enables SO_REUSEPORT where supported by the operating system |
| **DebugModule** | Enables Gin debug mode |
| **MaxRequestBodyBytes** | Maximum complete request-body size; zero means unlimited |
| **MaxMultipartMemory** | Multipart memory threshold; excess data is stored in temporary files |
| **HidePanicErrorDetails** | Hides unknown panic details from clients |
| **PanicResolver** | Converts an error into client-visible text |
| **DisableBadHTTPCodeResolver** | Disables REST conversion for error HTTP statuses |
| **DisableDefaultIgnoreHTTPCode** | Disables the built-in resolver ignore list |
| **IgnoreHTTPCode** | Adds HTTP statuses that bypass resolver conversion |
| **BadHTTPCodeResolver** | Builds responses for non-ignored error statuses |
| **GlobalPreInterceptors** | Global pre-interceptors in registration order |
| **GlobalPostInterceptors** | Global post-interceptors in registration order |
| **ResponseBodyEncoder** | Encodes structured response bodies |
| **TraceIDResponse** | Supplies the value of the **Trace-Id** response header |
| **DisableMethodNotAllowedError** | Uses not-found behavior instead of Gin's 405 handling |
| **DisableForwardedByClientIP** | Disables Gin forwarded-client-IP resolution |

## Common API

- **GinStarter.Start()** starts the singleton HTTP server.
- **GinStarter.Stop(maxWaitTime)** gracefully shuts it down.
- **RawGinEngine()** returns the current raw **gin.Engine**.
- **RespRestSuccess(...)** builds the standard successful REST envelope.
- **RespJSON(...)** builds a structured JSON response.
- **NewCommonResp()** and **NewRespRest()** create customizable response implementations.
- **NewResponseEntity(...)** creates a complete HTTP response entity.
- **GetStatusMessage(...)** returns the framework message for a status code.

## Design Notes

- Register routers before startup through **GinConfig.Routers**.
- Use **Request** and **Response** for framework-managed request processing.
- Use **GinContext** only when direct Gin integration is required.
- The starter owns one server and one Gin engine per process.
- **MaxMultipartMemory** is not an upload-size limit; use **MaxRequestBodyBytes** for a complete request limit.
- **RawGinEngine** is only meaningful while the starter is running.
