package ginmodule

import (
	"context"
	"github.com/acexy/golang-toolkit/logger"
	"github.com/gin-gonic/gin"
	"github.com/golang-acexy/starter-parent/parentmodule/declaration"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

var server *http.Server

const (
	defaultListenAddress = ":8080"
)

var (
	disabledDefaultIgnoreHttpStatusCode bool
	ignoreHttpStatusCode                []int
)

type GinModule struct {

	// 自定义Gin模块的组件属性
	GinModuleConfig *declaration.ModuleConfig

	// 模块组件在启动时执行初始化
	GinInterceptor func(instance *gin.Engine)

	// * 注册业务路由
	Routers []Router
	// * 注册服务监听地址 :8080 (默认)
	ListenAddress string // ip:port

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

	// gin config
	DebugModule        bool
	MaxMultipartMemory int64

	// 关闭包裹405错误展示，使用404代替
	DisableMethodNotAllowedError bool

	// 开启尝试获取真实IP
	ForwardedByClientIP bool
}

func (g *GinModule) ModuleConfig() *declaration.ModuleConfig {
	if g.GinModuleConfig != nil {
		return g.GinModuleConfig
	}
	return &declaration.ModuleConfig{
		ModuleName:               "Gin",
		UnregisterPriority:       0,
		UnregisterAllowAsync:     true,
		UnregisterMaxWaitSeconds: 30,
		LoadInterceptor: func(instance interface{}) {
			if g.GinInterceptor != nil {
				g.GinInterceptor(instance.(*gin.Engine))
			}
		},
	}
}

func (g *GinModule) Register() (interface{}, error) {

	var err error
	if g.DebugModule {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	gin.DefaultWriter = &logrusLogger{log: logger.Logrus(), level: logrus.DebugLevel}
	gin.DefaultErrorWriter = &logrusLogger{log: logger.Logrus(), level: logrus.ErrorLevel}
	ginEngin := gin.New()

	ginEngin.Use(recoverHandler())
	if g.RecoverHandlerResponse != nil {
		defaultRecoverHandlerResponse = g.RecoverHandlerResponse
	}

	if g.MaxMultipartMemory > 0 {
		ginEngin.MaxMultipartMemory = g.MaxMultipartMemory
	}

	ginEngin.ForwardedByClientIP = g.ForwardedByClientIP

	if !g.DisableMethodNotAllowedError {
		ginEngin.HandleMethodNotAllowed = true
	}

	if !g.DisableHttpStatusCodeHandler {
		ginEngin.Use(responseRewriteHandler())
		ginEngin.Use(httpStatusCodeHandler())
		disabledDefaultIgnoreHttpStatusCode = g.DisabledDefaultIgnoreHttpStatusCode
		ignoreHttpStatusCode = g.IgnoreHttpStatusCode
		if g.HttpStatusCodeCodeHandlerResponse != nil {
			defaultHttpStatusCodeHandlerResponse = g.HttpStatusCodeCodeHandlerResponse
		}
	}

	if g.ResponseDataStructDecoder != nil {
		defaultResponseDataDecoder = g.ResponseDataStructDecoder
	}

	if len(g.Routers) > 0 {
		registerRouter(ginEngin, g.Routers)
	}

	if g.ListenAddress == "" {
		g.ListenAddress = defaultListenAddress
	}

	server = &http.Server{
		Addr:    g.ListenAddress,
		Handler: ginEngin,
	}
	status := make(chan error)
	go func() {
		logger.Logrus().Traceln(g.ModuleConfig().ModuleName, "started")
		if err = server.ListenAndServe(); err != nil {
			status <- err
		}
		close(status)
	}()
	v, ok := <-status
	if !ok {
		return ginEngin, err
	}
	return nil, v
}

func (g *GinModule) Unregister(maxWaitSeconds uint) (gracefully bool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(maxWaitSeconds)*time.Second)
	defer cancel()
	if err = server.Shutdown(ctx); err != nil {
		gracefully = false
	} else {
		gracefully = true
	}
	return
}
