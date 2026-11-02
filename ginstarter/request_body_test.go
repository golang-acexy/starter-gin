package ginstarter

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func TestBodyCachePreservesReadAndBindingIsolation(t *testing.T) {
	const payload = `{"name":"original"}`
	for _, bindFirst := range []bool{false, true} {
		name := "raw-first"
		if bindFirst {
			name = "bind-first"
		}
		t.Run(name, func(t *testing.T) {
			request := newRequestForTest(http.MethodPost, gin.MIMEJSON, bytes.NewBufferString(payload))
			bind := func() {
				var target struct {
					Name string `json:"name" binding:"required"`
				}
				if err := request.BindBodyJSON(&target); err != nil || target.Name != "original" {
					t.Fatalf("binding: name=%q err=%v", target.Name, err)
				}
			}
			if bindFirst {
				bind()
			}
			first, err := request.GetRawBodyData()
			if err != nil || string(first) != payload {
				t.Fatalf("first read: %q err=%v", first, err)
			}
			first[0] = 'x'
			second, err := request.GetRawBodyData()
			if err != nil || string(second) != payload {
				t.Fatalf("cached read changed: %q err=%v", second, err)
			}
			second[1] = 'x'
			bind()
			readerData, err := io.ReadAll(request.ctx.Request.Body)
			if err != nil || string(readerData) != payload {
				t.Fatalf("reader changed: %q err=%v", readerData, err)
			}
			// Cached operations must still work after the replay reader has been consumed.
			bind()
			third, err := request.GetRawBodyData()
			if err != nil || string(third) != payload {
				t.Fatalf("read after reader EOF: %q err=%v", third, err)
			}
		})
	}
}

type mutatingJSONTarget struct{ captured []byte }

func (m *mutatingJSONTarget) UnmarshalJSON(data []byte) error {
	m.captured = bytes.Clone(data)
	clear(data)
	return nil
}

func TestJSONUnmarshalerCannotMutateBodyCache(t *testing.T) {
	const payload = `{"name":"original"}`
	request := newRequestForTest(http.MethodPost, gin.MIMEJSON, bytes.NewBufferString(payload))
	var target mutatingJSONTarget
	if err := request.BindBodyJSON(&target); err != nil {
		t.Fatal(err)
	}
	if string(target.captured) != payload {
		t.Fatalf("unexpected input: %q", target.captured)
	}
	body, err := request.GetRawBodyData()
	if err != nil || string(body) != payload {
		t.Fatalf("custom decoder mutated cache: %q err=%v", body, err)
	}
	readerData, err := io.ReadAll(request.ctx.Request.Body)
	if err != nil || string(readerData) != payload {
		t.Fatalf("custom decoder mutated reader: %q err=%v", readerData, err)
	}
}

type partialBodyError struct{}

func (partialBodyError) Read(p []byte) (int, error) { return copy(p, `{"name":`), io.ErrUnexpectedEOF }

func TestBodyReadFailuresDoNotPublishPartialCache(t *testing.T) {
	for _, mode := range []string{"read-error", "size-limit"} {
		for _, bind := range []bool{false, true} {
			t.Run(mode+map[bool]string{false: "/raw", true: "/bind"}[bind], func(t *testing.T) {
				request := newRequestForTest(http.MethodPost, gin.MIMEJSON, bytes.NewBufferString(`{"name":"original"}`))
				if mode == "read-error" {
					request.ctx.Request.Body = io.NopCloser(partialBodyError{})
				} else {
					request.ctx.Request.Body = http.MaxBytesReader(request.ctx.Writer, request.ctx.Request.Body, 8)
				}
				originalBody := request.ctx.Request.Body
				var err error
				if bind {
					err = request.BindBodyJSON(&struct {
						Name string `json:"name"`
					}{})
				} else {
					var body []byte
					body, err = request.GetRawBodyData()
					if body != nil {
						t.Fatalf("partial bytes escaped: %q", body)
					}
				}
				if mode == "read-error" {
					if !errors.Is(err, io.ErrUnexpectedEOF) {
						t.Fatalf("read error: %v", err)
					}
				} else {
					var limitError *http.MaxBytesError
					if !errors.As(err, &limitError) {
						t.Fatalf("size error: %v", err)
					}
				}
				if _, ok := request.ctx.Get(ginCtxKeyRequestBody); ok {
					t.Fatal("failed read published a cache")
				}
				if request.ctx.Request.Body != originalBody {
					t.Fatal("failed read replaced body reader")
				}
			})
		}
	}
}

func TestCachedJSONErrorsAndValidationRemainRepeatable(t *testing.T) {
	for _, payload := range []string{`{}`, `{"name":1}`, `{"name":`} {
		t.Run(payload, func(t *testing.T) {
			request := newRequestForTest(http.MethodPost, gin.MIMEJSON, bytes.NewBufferString(payload))
			for i := 0; i < 2; i++ {
				var target struct {
					Name string `json:"name" binding:"required"`
				}
				err := request.BindBodyJSON(&target)
				switch payload {
				case `{}`:
					var validationErrors validator.ValidationErrors
					if !errors.As(err, &validationErrors) {
						t.Fatalf("validation error: %v", err)
					}
				case `{"name":1}`:
					var typeError *json.UnmarshalTypeError
					if !errors.As(err, &typeError) {
						t.Fatalf("type error: %v", err)
					}
				default:
					if !errors.Is(err, io.ErrUnexpectedEOF) {
						t.Fatalf("syntax error: %v", err)
					}
				}
			}
		})
	}
}

func TestRawBodyEmptyResultsRemainNonNil(t *testing.T) {
	for _, absent := range []bool{false, true} {
		request := newRequestForTest(http.MethodPost, gin.MIMEJSON, &bytes.Buffer{})
		if absent {
			request.ctx.Request.Body = nil
		}
		for i := 0; i < 2; i++ {
			body, err := request.GetRawBodyData()
			if err != nil || body == nil || len(body) != 0 {
				t.Fatalf("empty body: %#v err=%v", body, err)
			}
		}
	}
}
