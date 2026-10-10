package ginstarter

import (
	"context"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/acexy/golang-toolkit/util/coll"
	"github.com/gin-gonic/gin"
	"github.com/golang-acexy/starter-parent/parent"
	"github.com/libp2p/go-reuseport"
	"github.com/sirupsen/logrus"
)

var ginRuntimeState atomic.Pointer[ginRuntime]
var ginLifecycleLock sync.Mutex

type ginRuntime struct {
	server   *http.Server
	engine   *gin.Engine
	config   *GinConfig
	stopping atomic.Bool
}

// GinConfig configures the singleton server. Treat configuration and callback state as
// immutable after startup; callbacks can execute concurrently for different requests.
type GinConfig struct {
	// InitFunc is invoked by the parent loader after startup using the default Setting.
	InitFunc func(instance *gin.Engine)
	// Routers registers business routes during startup; nil entries are ignored.
	Routers []Router
	// ListenAddress is the TCP listen address; an empty value defaults to :8080.
	ListenAddress string
	// UseReusePortModel enables SO_REUSEPORT where supported by the operating system.
	UseReusePortModel bool
	// HidePanicErrorDetails hides unknown panic text but preserves explicitly assigned
	// HTTP status codes. Framework validation messages remain visible.
	HidePanicErrorDetails bool
	// PanicResolver converts recovered errors to client-visible text. It must not panic.
	PanicResolver PanicResolver
	// DisableBadHTTPCodeResolver preserves non-200 responses without REST conversion.
	DisableBadHTTPCodeResolver bool
	// DisableDefaultIgnoreHTTPCode removes the built-in redirect/no-content ignore list.
	DisableDefaultIgnoreHTTPCode bool
	// IgnoreHTTPCode adds statuses that bypass conversion, even when defaults are disabled.
	IgnoreHTTPCode []int
	// BadHTTPCodeResolver replaces the entire non-ignored, non-200 response.
	// Preserve desired entity headers and cookies explicitly; this callback must not panic.
	BadHTTPCodeResolver BadHTTPCodeResolver
	// GlobalPreInterceptors run before route interceptors and handlers in registration order.
	// A nonnil Response or error skips the remaining pre-list and route scope; global posts
	// still run. Returning nil, nil continues, including after native writes or Gin Abort.
	GlobalPreInterceptors []PreInterceptor
	// GlobalPostInterceptors run after route posts in registration order. A nonnil Response
	// replaces the current response; nil retains it. An error stops this post-list.
	GlobalPostInterceptors []PostInterceptor
	// ResponseBodyEncoder encodes structured bodies; nil selects JSON. Raw byte bodies
	// bypass it. JSON-marked responses retain that media type after custom encoding.
	ResponseBodyEncoder ResponseBodyEncoder
	// TraceIDResponse supplies the Trace-Id header at final write time. It receives no
	// request argument, must support concurrent calls, and must not panic.
	TraceIDResponse func() string
	// DebugModule sets Gin's process-wide debug mode; false selects release mode.
	DebugModule bool
	// MaxMultipartMemory is the in-memory multipart threshold, not a body-size limit.
	// Nonpositive values retain Gin's default threshold; larger parts use temporary files.
	MaxMultipartMemory int64
	// MaxRequestBodyBytes limits request-body reads; nonpositive values mean unlimited.
	MaxRequestBodyBytes int64
	// DisableMethodNotAllowedError uses not-found handling instead of method-not-allowed.
	DisableMethodNotAllowedError bool
	// DisableForwardedByClientIP disables forwarded-client-IP resolution in Gin.
	DisableForwardedByClientIP bool
}

// GinStarter owns the process-wide Gin server. Configure it before parent startup;
// setting fields or registering routes while serving requests is not supported.
type GinStarter struct {
	// Config supplies static configuration when LazyConfig is nil.
	Config GinConfig
	// LazyConfig takes precedence over Config and is evaluated once on first resolution.
	LazyConfig func() GinConfig
	config     *GinConfig
	configOnce sync.Once
	// GinSetting overrides the default parent lifecycle setting, including its init hook.
	GinSetting *parent.Setting
}

// getConfig resolves and caches configuration once.
func (g *GinStarter) getConfig() *GinConfig {
	g.configOnce.Do(func() {
		if g.LazyConfig != nil {
			config := g.LazyConfig()
			g.config = &config
		} else {
			config := g.Config
			g.config = &config
		}
	})
	return g.config
}

// Setting returns the custom setting or a default non-restartable parent setting.
func (g *GinStarter) Setting() *parent.Setting {
	if g.GinSetting != nil {
		return g.GinSetting
	}
	config := g.getConfig()
	return parent.NewSetting(
		"Gin-Starter",
		false,
		1,
		false,
		time.Second*30,
		func(instance any) {
			if config.InitFunc != nil {
				config.InitFunc(instance.(*gin.Engine))
			}
		})
}

// Start registers routes and begins serving; an active singleton returns ErrGinStarterAlreadyStarted.
// The returned engine is shared. Use the parent loader to invoke the configured init hook.
func (g *GinStarter) Start() (any, error) {
	var err error
	ginLifecycleLock.Lock()
	if runtime := ginRuntimeState.Load(); runtime != nil {
		ginLifecycleLock.Unlock()
		return runtime.engine, ErrGinStarterAlreadyStarted
	}
	defer ginLifecycleLock.Unlock()
	config := g.getConfig()
	if config.DebugModule {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	gin.DefaultWriter = &logrusLogger{level: logrus.DebugLevel}
	gin.DefaultErrorWriter = &logrusLogger{level: logrus.ErrorLevel}

	ginEngine := gin.New()
	registerValidators()
	if config.PanicResolver == nil {
		config.PanicResolver = panicResolver
	}

	if config.MaxMultipartMemory > 0 {
		ginEngine.MaxMultipartMemory = config.MaxMultipartMemory
	}

	ginEngine.ForwardedByClientIP = !config.DisableForwardedByClientIP

	if !config.DisableMethodNotAllowedError {
		ginEngine.HandleMethodNotAllowed = true
	}

	if config.BadHTTPCodeResolver == nil {
		config.BadHTTPCodeResolver = badHTTPCodeResolver
	}

	if config.ResponseBodyEncoder == nil {
		config.ResponseBodyEncoder = jsonResponseBodyEncoder{}
	}
	config.GlobalPreInterceptors = coll.SliceFilter(config.GlobalPreInterceptors, func(p PreInterceptor) bool {
		return p != nil
	})
	config.GlobalPostInterceptors = coll.SliceFilter(config.GlobalPostInterceptors, func(p PostInterceptor) bool {
		return p != nil
	})
	// Install the response pipeline outermost so it commits the final Response once.
	ginEngine.Use(responsePipelineHandler())
	// Keep pre/post interceptors in one scope so posts still run after pre-list short-circuiting.
	ginEngine.Use(func(ctx *gin.Context) {
		defer recoverResponse(ctx)
		state := getRequestState(ctx)
		runPreInterceptors(ctx, config.GlobalPreInterceptors)
		if !state.stopped {
			nextWithRecovery(ctx)
		}
		normalizeResponse(ctx)
		runPostInterceptors(ctx, config.GlobalPostInterceptors)
	})

	ginEngine.NoRoute(func(ctx *gin.Context) {
		setResponse(ctx, RespHTTPStatusCode(http.StatusNotFound))
	})
	ginEngine.NoMethod(func(ctx *gin.Context) {
		setResponse(ctx, RespHTTPStatusCode(http.StatusMethodNotAllowed))
	})

	config.Routers = coll.SliceFilter(config.Routers, func(r Router) bool {
		return r != nil
	})
	if len(config.Routers) > 0 {
		if err = registerRouter(ginEngine, config.Routers); err != nil {
			return nil, err
		}
	}

	if config.ListenAddress == "" {
		config.ListenAddress = ":8080"
	}

	var listener net.Listener
	if config.UseReusePortModel {
		listener, err = reuseport.Listen("tcp", config.ListenAddress)
	} else {
		listener, err = net.Listen("tcp", config.ListenAddress)
	}
	if err != nil {
		return nil, err
	}

	currentServer := &http.Server{
		Addr:    config.ListenAddress,
		Handler: ginEngine,
	}
	runtime := &ginRuntime{server: currentServer, engine: ginEngine, config: config}
	ginRuntimeState.Store(runtime)

	go func() {
		if serveErr := currentServer.Serve(listener); serveErr != nil && serveErr != http.ErrServerClosed {
			logrus.Errorln("gin server stopped unexpectedly:", serveErr)
			ginRuntimeState.CompareAndSwap(runtime, nil)
		}
	}()
	return ginEngine, nil
}

// Stop attempts graceful shutdown within maxWaitTime; a failed shutdown keeps runtime state.
// An absent or already-stopping server returns ErrGinServerNotStarted. Nonpositive waits set an expired deadline.
func (g *GinStarter) Stop(maxWaitTime time.Duration) (gracefully, stopped bool, err error) {
	ginLifecycleLock.Lock()
	runtime := ginRuntimeState.Load()
	if runtime == nil {
		ginLifecycleLock.Unlock()
		return false, true, ErrGinServerNotStarted
	}
	currentServer := runtime.server
	ginLifecycleLock.Unlock()
	if !runtime.stopping.CompareAndSwap(false, true) {
		return false, true, ErrGinServerNotStarted
	}
	ctx, cancel := context.WithTimeout(context.Background(), maxWaitTime)
	defer cancel()
	if err = currentServer.Shutdown(ctx); err != nil {
		runtime.stopping.Store(false)
		gracefully = false
		stopped = false
	} else {
		gracefully = true
		stopped = true
		ginLifecycleLock.Lock()
		ginRuntimeState.CompareAndSwap(runtime, nil)
		ginLifecycleLock.Unlock()
	}
	return
}

// RawGinEngine returns the shared engine, or nil before startup and after successful shutdown.
// It remains available during shutdown; do not mutate route registration while requests are served.
func RawGinEngine() *gin.Engine {
	runtime := ginRuntimeState.Load()
	if runtime == nil {
		return nil
	}
	return runtime.engine
}

func currentGinConfig() *GinConfig {
	runtime := ginRuntimeState.Load()
	if runtime == nil {
		return nil
	}
	return runtime.config
}
