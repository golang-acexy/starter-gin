package ginstarter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"
	"weak"

	"github.com/gin-gonic/gin"
)

func waitGinPressure(t *testing.T, description string, ready func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !ready() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", description)
		}
		time.Sleep(time.Millisecond)
	}
}

func TestMultipartTemporaryFilesReleased(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("TMPDIR", tempDir)
	files := make(chan string, 16)
	engine := startPressureGin(t, nil, func(w *RouterWrapper) {
		w.POST("/upload", func(r *Request) (Response, error) {
			header, err := r.GetFormFile("file")
			if err != nil {
				return nil, err
			}
			file, err := header.Open()
			if err != nil {
				return nil, err
			}
			defer file.Close()
			diskFile, ok := file.(*os.File)
			if !ok {
				return nil, fmt.Errorf("expected disk-backed multipart file")
			}
			if n, err := io.Copy(io.Discard, file); err != nil || n != 64<<10 {
				return nil, fmt.Errorf("upload read: bytes=%d err=%v", n, err)
			}
			files <- diskFile.Name()
			return RespTextPlain([]byte("uploaded")), nil
		})
	})
	server := httptest.NewServer(engine)
	t.Cleanup(server.Close)
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "pressure.bin")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(bytes.Repeat([]byte("x"), 64<<10)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	client := server.Client()
	client.Timeout = 5 * time.Second
	for i := 0; i < 16; i++ {
		response, err := client.Post(server.URL+"/upload", writer.FormDataContentType(), bytes.NewReader(body.Bytes()))
		if err != nil {
			t.Fatal(err)
		}
		result, readErr := io.ReadAll(response.Body)
		response.Body.Close()
		if readErr != nil || string(result) != "uploaded" {
			t.Fatalf("upload response: %q err=%v", result, readErr)
		}
		var path string
		select {
		case path = <-files:
		default:
			t.Fatal("no disk upload observed")
		}
		// net/http cleanup may finish just after the client receives the response.
		waitGinPressure(t, "multipart file removal", func() bool {
			_, err := os.Stat(path)
			return errors.Is(err, os.ErrNotExist)
		})
	}
	entries, err := os.ReadDir(tempDir)
	if err != nil || len(entries) != 0 {
		t.Fatalf("temporary uploads remain: count=%d err=%v", len(entries), err)
	}
}

func TestGinShutdownTimeoutThenDrain(t *testing.T) {
	if ginRuntimeState.Load() != nil {
		t.Fatal("test requires an empty Gin runtime")
	}
	entered, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	engine := gin.New()
	engine.GET("/blocked", func(c *gin.Context) { close(entered); <-release; c.Status(http.StatusOK) })
	server := httptest.NewServer(engine)
	// Install a real, addressable HTTP server to exercise Stop with an active request.
	current := &ginRuntime{server: server.Config, engine: engine, config: &GinConfig{}}
	ginRuntimeState.Store(current)
	t.Cleanup(func() { unblock(); server.Close(); ginRuntimeState.CompareAndSwap(current, nil) })
	client := server.Client()
	client.Timeout = 5 * time.Second
	done := make(chan error, 1)
	go func() {
		response, err := client.Get(server.URL + "/blocked")
		if err == nil {
			_, err = io.Copy(io.Discard, response.Body)
			response.Body.Close()
		}
		done <- err
	}()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("handler did not start")
	}
	starter := &GinStarter{}
	graceful, stopped, err := starter.Stop(time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) || graceful || stopped || RawGinEngine() != engine {
		t.Fatalf("timeout state: graceful=%v stopped=%v err=%v", graceful, stopped, err)
	}
	unblock()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("request did not complete")
	}
	graceful, stopped, err = starter.Stop(time.Second)
	if err != nil || !graceful || !stopped || RawGinEngine() != nil {
		t.Fatalf("drained shutdown: graceful=%v stopped=%v err=%v", graceful, stopped, err)
	}
}

func runGinLifecyclePressure(t *testing.T) weak.Pointer[ginRuntime] {
	t.Helper()
	starter := &GinStarter{Config: GinConfig{ListenAddress: "127.0.0.1:0"}}
	if _, err := starter.Start(); err != nil {
		t.Fatal(err)
	}
	ref := weak.Make(ginRuntimeState.Load())
	if _, stopped, err := starter.Stop(time.Second); err != nil || !stopped {
		forceClosePressureGin()
		t.Fatalf("shutdown failed: stopped=%v err=%v", stopped, err)
	}
	return ref
}

func TestGinLifecycleResourcesReleased(t *testing.T) {
	if ginRuntimeState.Load() != nil {
		t.Fatal("test requires an empty Gin runtime")
	}
	refs := make([]weak.Pointer[ginRuntime], 0, 100)
	for i := 0; i < 100; i++ {
		refs = append(refs, runGinLifecyclePressure(t))
	}
	// Each runtime is also captured by its Serve goroutine until that goroutine exits.
	waitGinPressure(t, "HTTP runtime release", func() bool {
		runtime.GC()
		runtime.GC()
		for _, ref := range refs {
			if ref.Value() != nil {
				return false
			}
		}
		return true
	})
}
