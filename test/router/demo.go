package router

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-acexy/starter-gin/ginstarter"
)

// DemoRouter implements the manual integration example.
type DemoRouter struct {
}

// Info returns the example route-group configuration.
func (d *DemoRouter) Info() *ginstarter.RouterInfo {
	return &ginstarter.RouterInfo{
		GroupPath: "demo",
	}
}

// Handlers registers example handlers during startup.
func (d *DemoRouter) Handlers(router *ginstarter.RouterWrapper) {

	router.MATCH([]string{http.MethodGet, http.MethodPost}, "more", d.more())

	// The /demo/exception route demonstrates returned-error recovery.
	router.GET("error1", d.error1())
	router.GET("error2", d.error2())
	router.GET("error3", d.error3())

	// The /demo/hold route delays completion for five seconds.
	router.GET("hold", d.hold())

	router.GET("empty", d.empty())

	router.GET("redirect", d.redirect())
	router.GET("common", d.common())
}

func (d *DemoRouter) more() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		fmt.Println("invoke")
		// Build a custom REST body with additional HTTP metadata.
		return ginstarter.NewRespRest().DataBuilder(func() *ginstarter.ResponseEntity {
			return ginstarter.NewEmptyResponseEntity().SetStatusCode(http.StatusAccepted).SetBody([]byte("success")).AddHeader("test", "test")
		}), nil
	}
}

func (d *DemoRouter) error1() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		// Return an error to enter recovery.
		return nil, errors.New("return error")
	}
}

func (d *DemoRouter) error2() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		// Demonstrate recovery from an unhandled panic.
		panic("panic exception")
	}
}

func (d *DemoRouter) error3() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		// Demonstrate recovery from an unhandled panic.
		ii := 1
		ii -= 1
		i := 1 / ii
		fmt.Println(i)
		return nil, nil
	}
}

func (d *DemoRouter) hold() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		fmt.Println("invoke")
		time.Sleep(time.Second * 5)
		return ginstarter.RespTextPlain([]byte("text")), nil
	}
}

func (d *DemoRouter) empty() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		return nil, nil
	}
}

func (d *DemoRouter) redirect() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		request.GinContext().Redirect(http.StatusMovedPermanently, "https://google.com")
		return nil, nil
	}
}

func (d *DemoRouter) common() ginstarter.HandlerWrapper {
	return func(request *ginstarter.Request) (ginstarter.Response, error) {
		return ginstarter.NewCommonResp().DataBuilder(func() *ginstarter.ResponseEntity {
			return ginstarter.NewEmptyResponseEntity().SetBody([]byte("success"))
		}), nil
	}
}
