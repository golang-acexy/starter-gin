# starter-gin

基于`github.com/gin-gonic/gin`封装的http服务组件

---

#### 功能说明

屏蔽其他原始框架细节，提供统一的http服务注册方法，用户只需要关系Request/Response即可

- 初始化
    ```go
    loader := parent.NewStarterLoader([]parent.Starter{
        &ginstarter.GinStarter{
            ListenAddress: ":8080",
            DebugModule:   true,
            Routers: []ginstarter.Router{
                &router.DemoRouter{},
                &router.ParamRouter{},
                &router.AbortRouter{},
                &router.BasicAuthRouter{},
                &router.MyRestRouter{},
            },
            InitFunc: func(instance *gin.Engine) {
                instance.GET("/ping", func(context *gin.Context) {
                    context.String(http.StatusOK, "alive")
                })
                instance.GET("/err", func(context *gin.Context) {
                    context.Status(500)
                })
            },
            DisabledDefaultIgnoreHttpStatusCode: true,
            DisableMethodNotAllowedError:        true,
            RecoverHandlerResponse: func(ctx *gin.Context, err any) ginstarter.Response {
                logger.Logrus().Errorln("Request catch exception", err)
                return ginstarter.RespTextPlain("something error", http.StatusOK)
            },
            DisableHttpStatusCodeHandler: true,
        },
    })
    
    err := loader.Start()
    if err != nil {
        fmt.Printf("%+v\n", err)
        return
    }
    
    sys.ShutdownHolding()
    ```

- 注册路由
    ```go
    type RouterInfo struct {
        // GroupPath 路由分组路径
        GroupPath string
    
        // BasicAuthAccount 如果指定基于BasicAuth认证的账户，则该GroupPath下资源将需要权限认证
        BasicAuthAccount *BasicAuthAccount
    }
    ```
  
    然后在初始化时指定级自动完成注册

    ```go
    type AbortRouter struct { // 业务路由
    }
    
    func (a *AbortRouter) Info() *ginstarter.RouterInfo {
        return &ginstarter.RouterInfo{ // 配置路由信息
            GroupPath: "abort",        // 
        }
    }
    
    // Handlers 路由功能
    func (a *AbortRouter) Handlers(router *ginstarter.RouterWrapper) {
        // 注册一个GET方法处理器
        router.GET("invoke", a.invoke())
    }
    
    func (a *AbortRouter) invoke() ginstarter.HandlerWrapper {
        return func(request *ginstarter.Request) (ginstarter.Response, error) {
            
            // TODO: 完成业务
            
            // 响应一个http状态码203
            return ginstarter.RespAbortWithStatus(203), nil
        }
    }
    ```
- 请求
    框架已封装常用的参数获取方法，在Handler中直接通过request执行，以`Must`开始的方法将在参数不满足条件时直接触发Panic，快速实现验参数不通过中断请求
    
    ```go
    // HttpMethod 获取请求方法
    HttpMethod() string
    
    // FullPath 获取请求全路径
    FullPath() string
    
    // RawGinContext 获取原始Gin上下文
    RawGinContext() *gin.Context
    
    // RequestIP 尝试获取请求方客户端IP
    RequestIP() string
    
    
    // GetPathParam 获取path路径参数 /:id/
    GetPathParam(name string) string
    
    // GetPathParams 获取path路径参数 /:id/ 多个参数
    GetPathParams(names ...string) map[string]string
    
    // BindPathParams /:id/ 绑定结构体用于接收UriPath参数 结构体标签格式 `uri:""`
    BindPathParams(object any) error
    
    // MustBindPathParams /:id/ 绑定结构体用于接收UriPath参数 结构体标签格式 `uri:""`
    // 任何错误将触发Panic流程中断
    MustBindPathParams(object any)
    ...
    ```
- 响应 框架已封装常用响应体方法，方法均以`Resp`开始

    ```go
    // RespJson 响应Json数据
    func RespJson(data any, httpStatusCode ...int) Response
    
    // RespXml 响应Xml数据
    func RespXml(data any, httpStatusCode ...int) Response
    
    // RespYaml 响应Yaml数据
    func RespYaml(data any, httpStatusCode ...int) Response
    
    // RespToml 响应Toml数据
    func RespToml(data any, httpStatusCode ...int) Response
    
    // RespTextPlain 响应Json数据
    func RespTextPlain(data string, httpStatusCode ...int) Response
    
    // RespRedirect 响应重定向
    func RespRedirect(url string, httpStatusCode ...int) Response
    ```
    #### 特别的Rest响应，默认框架已定制一套Rest响应标准
    ```go
    // RestRespStatusStruct 框架默认的Rest请求状态结构
    type RestRespStatusStruct struct {
    
        // 标识请求系统状态 200 标识网络请求层面的成功 见StatusCode
        StatusCode    StatusCode    `json:"statusCode"`
        StatusMessage StatusMessage `json:"statusMessage"`
    
        // 业务错误码 仅当StatusCode为200时进入业务错误判断
        BizErrorCode    *BizErrorCode    `json:"bizErrorCode"`
        BizErrorMessage *BizErrorMessage `json:"bizErrorMessage"`
    
        // 系统响应时间戳
        Timestamp int64 `json:"timestamp"`
    }
    
    // RestRespStruct 框架默认的Rest请求结构
    type RestRespStruct struct {
    
        // 请求状态描述
        Status *RestRespStatusStruct `json:"status"`
    
        // 仅当StatusCode为200 无业务错误码BizErrorCode 响应成功数据
        Data any `json:"data"`
    }
    ```
    通过`RespRest`开始的方法名执行该结构体的Rest风格响应，如果需要在此基础上响应更多信息，则可以使用`NewRespRest()`创建Rest响应实例，设置head、cookie等其他信息

    > 如果你要自定义Rest结构体响应风格

    参考test/router/myrest.go

    方法1 通过NewRespRest()实例，每次传输自定义的Rest结构体响应

    ```go
    ginstarter.NewRespRest().SetDataResponse(&RestStruct{
			Code: 200,
			Msg:  "success",
			Data: "invoke",
		})
    ```

    方法2 实现`Response`接口，参照NewRespRest方法设计，定制并使用自定义Response数据响应

    ```go// 自实现Response响应数据
    func (m *MyRestRouter) m3() ginstarter.HandlerWrapper {
        return func(request *ginstarter.Request) (ginstarter.Response, error) {
            response := &MyRestResponse{}
            response.setData(&RestStruct{
                Code: 200,
                Msg:  "success",
                Data: "my rest impl",
            })
            return response, nil
        }
    }
    
    type MyRestResponse struct {
        responseData *ginstarter.ResponseData
    }
    
    func (m *MyRestResponse) Data() *ginstarter.ResponseData {
        return m.responseData
    }
    
    func (m *MyRestResponse) setData(data *RestStruct) {
        m.responseData = ginstarter.NewResponseData()
        m.responseData.SetData(json.ToJsonBytes(data))
    }
    ```
  
#### 高级用法

```go
type GinStarter struct {

    ...
	
	// 模块组件在启动时执行初始化
	InitFunc func(instance *gin.Engine)

	// 自定义异常响应处理 如果不指定则使用默认方式
	RecoverHandlerResponse RecoverHandlerResponse

	// 禁用错误包装处理器 在出现非200响应码或者异常时，将自动进行转化
	DisableHttpStatusCodeHandler bool
	// 在启用非200响应码自动处理后，指定忽略需要自动包裹响应码
	IgnoreHttpStatusCode []int
	// 关闭系统内置的忽略的http状态码
	DisabledDefaultIgnoreHttpStatusCode bool
	// 在出现非200响应码或者异常时具体响应策略 如果不指定则使用默认处理器 仅在UseHttpStatusCodeHandler = true 生效
	HttpStatusCodeCodeHandlerResponse HttpStatusCodeCodeHandlerResponse

	// 响应数据的结构体解码器 默认为JSON方式解码
	// 在使用NewRespRest响应结构体数据时解码为[]byte数据的解码器
	// 如果自实现Response接口将不使用解码器
	ResponseDataStructDecoder ResponseDataStructDecoder

	// 关闭包裹405错误展示，使用404代替
	DisableMethodNotAllowedError bool

	// 禁用尝试获取真实IP
	DisableForwardedByClientIP bool
	...
}
```

- RecoverHandlerResponse

    框架自动捕获Panic异常，并执行默认处理方案，如果你想替换默认处理逻辑，可以注册该方法的具体实现逻辑

- DisableHttpStatusCodeHandler
    
    默认情况，当遇到非200响应码时，会自动进行响应处理，如果你不想使用默认处理方案，可以禁用该功能。IgnoreHttpStatusCode允许配置非200的例外错误码，不触发此逻辑

- HttpStatusCodeCodeHandlerResponse
    
    在DisableHttpStatusCodeHandler为false时，如果你想替换默认的非200错误码响应处理方案，可以注册该方法的具体实现逻辑 