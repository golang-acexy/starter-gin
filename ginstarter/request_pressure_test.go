package ginstarter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"runtime"
	"runtime/debug"
	"sync"
	"testing"
	"time"
	"weak"

	"github.com/gin-gonic/gin"
)

type pressureRouter struct{ register func(*RouterWrapper) }

func (r *pressureRouter) Info() *RouterInfo         { return &RouterInfo{} }
func (r *pressureRouter) Handlers(w *RouterWrapper) { r.register(w) }

type pressureJSON struct {
	Message string `json:"message" binding:"required"`
}

type requestPressureRefs struct {
	state  weak.Pointer[requestState]
	entity weak.Pointer[ResponseEntity]
	body   weak.Pointer[byte]
}

type requestPressureObserver struct {
	sync.Mutex
	refs []requestPressureRefs
}

func (o *requestPressureObserver) observe(r *Request, response Response) (Response, error) {
	ref := requestPressureRefs{state: weak.Make(getRequestState(r.ctx))}
	if response != nil {
		ref.entity = weak.Make(response.Data())
	}
	if value, ok := r.ctx.Get(ginCtxKeyRequestBody); ok {
		if body, ok := value.([]byte); ok && len(body) > 0 {
			ref.body = weak.Make(&body[0])
		}
	}
	o.Lock()
	o.refs = append(o.refs, ref)
	o.Unlock()
	return nil, nil
}

func (o *requestPressureObserver) take() []requestPressureRefs {
	o.Lock()
	defer o.Unlock()
	refs := o.refs
	o.refs = nil
	return refs
}

func liveRequestPressureRefs(refs []requestPressureRefs) int {
	live := 0
	for _, ref := range refs {
		if ref.state.Value() != nil {
			live++
		}
		if ref.entity.Value() != nil {
			live++
		}
		if ref.body.Value() != nil {
			live++
		}
	}
	return live
}

func startPressureGin(tb testing.TB, observer *requestPressureObserver, extra func(*RouterWrapper)) *gin.Engine {
	tb.Helper()
	if ginRuntimeState.Load() != nil {
		tb.Fatal("test requires an empty Gin runtime")
	}
	config := GinConfig{ListenAddress: "127.0.0.1:0", MaxRequestBodyBytes: 1 << 20, MaxMultipartMemory: 1024}
	if observer != nil {
		config.GlobalPostInterceptors = []PostInterceptor{observer.observe}
	}
	config.Routers = []Router{&pressureRouter{register: func(w *RouterWrapper) {
		w.POST("/status", func(r *Request) (Response, error) { return RespHTTPStatusCode(http.StatusOK), nil })
		w.POST("/raw", func(r *Request) (Response, error) { return RespTextPlain(r.MustGetRawBodyData()), nil })
		w.POST("/raw-twice", func(r *Request) (Response, error) {
			first := r.MustGetRawBodyData()
			second := r.MustGetRawBodyData()
			if !bytes.Equal(first, second) {
				return nil, fmt.Errorf("body reread mismatch")
			}
			return RespTextPlain(second), nil
		})
		w.POST("/json", func(r *Request) (Response, error) {
			var body pressureJSON
			r.MustBindBodyJSON(&body)
			return RespJSON(body), nil
		})
		w.POST("/native", func(r *Request) (Response, error) {
			r.GinContext().Status(http.StatusOK)
			return nil, nil
		})
		if extra != nil {
			extra(w)
		}
	}}}
	starter := &GinStarter{Config: config}
	engine, err := starter.Start()
	if err != nil {
		tb.Fatal(err)
	}
	tb.Cleanup(func() {
		if _, stopped, err := starter.Stop(5 * time.Second); err != nil || !stopped {
			forceClosePressureGin()
			tb.Errorf("cleanup shutdown: stopped=%v err=%v", stopped, err)
		}
	})
	return engine.(*gin.Engine)
}

func forceClosePressureGin() {
	if current := ginRuntimeState.Swap(nil); current != nil {
		_ = current.server.Close()
	}
}

// pressureWriter discards output bytes so the harness does not retain responses.
type pressureWriter struct {
	header       http.Header
	status, size int
}

func (w *pressureWriter) Header() http.Header { return w.header }
func (w *pressureWriter) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
	}
}
func (w *pressureWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	w.size += len(body)
	return len(body), nil
}
func (w *pressureWriter) reset() { clear(w.header); w.status = 0; w.size = 0 }

func pressureRequest(path string, body []byte) *http.Request {
	return &http.Request{Method: http.MethodPost, URL: &url.URL{Path: path},
		Header: http.Header{"Content-Type": []string{gin.MIMEJSON}},
		Body:   io.NopCloser(bytes.NewReader(body)), ContentLength: int64(len(body))}
}

func TestRequestPoolReleasesPayloads(t *testing.T) {
	observer := &requestPressureObserver{}
	engine := startPressureGin(t, observer, nil)
	oldGC := debug.SetGCPercent(-1)
	defer debug.SetGCPercent(oldGC)
	for _, test := range []struct {
		path   string
		body   []byte
		status StatusCode
	}{
		{"/raw", bytes.Repeat([]byte("x"), 64<<10), 0},
		{"/json", []byte(`{"message":"ok"}`), 0},
		{"/native", nil, 0},
		{"/json", []byte(`{"message":`), StatusCodeBadRequestParameters},
		{"/raw", bytes.Repeat([]byte("x"), (1<<20)+1), StatusCodeUploadLimitExceeded},
	} {
		recorder := httptest.NewRecorder()
		engine.ServeHTTP(recorder, pressureRequest(test.path, test.body))
		if recorder.Code != http.StatusOK {
			t.Fatalf("HTTP status: %d", recorder.Code)
		}
		if test.status != 0 {
			var envelope RestRespStruct
			if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil || envelope.Status == nil || envelope.Status.StatusCode != test.status {
				t.Fatalf("error envelope: status=%+v err=%v", envelope.Status, err)
			}
		}
	}
	refs := observer.take()
	t.Logf("references still live before explicit GC: %d", liveRequestPressureRefs(refs))
	// Gin clears a pooled context on reuse; two idle GCs also expire pool victims.
	runtime.GC()
	t.Logf("references still live after first idle GC: %d", liveRequestPressureRefs(refs))
	runtime.GC()
	if live := liveRequestPressureRefs(refs); live != 0 {
		t.Fatalf("request references retained after idle GCs: %d", live)
	}
	runtime.KeepAlive(engine)
}

func runRequestPressureBatch(t *testing.T, engine *gin.Engine, path string, payload []byte, count int) {
	t.Helper()
	var wg sync.WaitGroup
	for worker := 0; worker < 16; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			writer := &pressureWriter{header: make(http.Header)}
			for i := worker; i < count; i += 16 {
				writer.reset()
				engine.ServeHTTP(writer, pressureRequest(path, payload))
				if writer.status != http.StatusOK || path != "/status" && writer.size != len(payload) {
					t.Errorf("response mismatch: status=%d bytes=%d want=%d", writer.status, writer.size, len(payload))
					return
				}
			}
		}(worker)
	}
	wg.Wait()
}

func TestRequestMemoryPressure(t *testing.T) {
	if os.Getenv("GIN_MEMORY_STRESS") != "1" {
		t.Skip("set GIN_MEMORY_STRESS=1 for batched request retention measurements")
	}
	for _, path := range []string{"/status", "/raw", "/json"} {
		t.Run(path, func(t *testing.T) {
			observer := &requestPressureObserver{}
			engine := startPressureGin(t, observer, nil)
			payload := bytes.Repeat([]byte("x"), 64<<10)
			if path == "/status" {
				payload = nil
			}
			if path == "/json" {
				payload = append(append([]byte(`{"message":"`), payload...), []byte(`"}`)...)
			}
			runRequestPressureBatch(t, engine, path, payload, 100)
			observer.take()
			runtime.GC()
			runtime.GC()
			var baseline runtime.MemStats
			runtime.ReadMemStats(&baseline)
			baseGoroutines := runtime.NumGoroutine()
			for batch := 1; batch <= 5; batch++ {
				runRequestPressureBatch(t, engine, path, payload, 1000)
				refs := observer.take()
				if len(refs) != 1000 {
					t.Fatalf("observed %d requests", len(refs))
				}
				runtime.GC()
				runtime.GC()
				live := liveRequestPressureRefs(refs)
				var stats runtime.MemStats
				runtime.ReadMemStats(&stats)
				t.Logf("batch=%d requests=%d live_refs=%d heap=%d heap_delta=%d goroutines=%d base_goroutines=%d total_alloc_delta=%d", batch, batch*1000, live, stats.HeapAlloc, int64(stats.HeapAlloc)-int64(baseline.HeapAlloc), runtime.NumGoroutine(), baseGoroutines, stats.TotalAlloc-baseline.TotalAlloc)
				if live != 0 {
					t.Fatalf("request graph retained after idle GCs: %d references", live)
				}
			}
			runtime.KeepAlive(engine)
		})
	}
}

func BenchmarkRequestPipeline(b *testing.B) {
	engine := startPressureGin(b, nil, nil)
	native := gin.New()
	native.POST("/status", func(c *gin.Context) { c.Status(http.StatusOK) })
	for _, test := range []struct {
		name, path string
		size       int
		parallel   bool
	}{
		{"native-status", "/status", 0, false}, {"wrapped-status", "/status", 0, false},
		{"raw-1KiB", "/raw", 1024, false}, {"raw-64KiB", "/raw", 65536, false},
		{"raw-twice-64KiB", "/raw-twice", 65536, false},
		{"json-1KiB", "/json", 1024, false}, {"json-64KiB", "/json", 65536, false},
		{"parallel-json-1KiB", "/json", 1024, true},
	} {
		b.Run(test.name, func(b *testing.B) {
			payload := bytes.Repeat([]byte("x"), test.size)
			if test.path == "/json" {
				payload = append(append([]byte(`{"message":"`), payload...), []byte(`"}`)...)
			}
			target := engine
			if test.name == "native-status" {
				target = native
			}
			perform := func(writer *pressureWriter) {
				writer.reset()
				target.ServeHTTP(writer, pressureRequest(test.path, payload))
				if writer.status != http.StatusOK || test.size > 0 && writer.size != len(payload) {
					b.Errorf("response mismatch: %d/%d", writer.status, writer.size)
				}
			}
			writer := &pressureWriter{header: make(http.Header)}
			perform(writer)
			b.ReportAllocs()
			b.SetBytes(int64(len(payload)))
			b.ResetTimer()
			if test.parallel {
				b.RunParallel(func(pb *testing.PB) {
					writer := &pressureWriter{header: make(http.Header)}
					for pb.Next() {
						perform(writer)
					}
				})
			} else {
				for i := 0; i < b.N; i++ {
					perform(writer)
				}
			}
		})
	}
}
