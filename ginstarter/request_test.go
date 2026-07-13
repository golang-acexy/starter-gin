package ginstarter

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newRequestForTest(method, contentType string, body *bytes.Buffer) *Request {
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(method, "/", body)
	context.Request.Header.Set("Content-Type", contentType)
	return &Request{ctx: context}
}

func TestBindBodyAutoMultipart(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if err := writer.WriteField("name", "acexy"); err != nil {
		t.Fatal(err)
	}
	fileWriter, err := writer.CreateFormFile("file", "test.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = fileWriter.Write([]byte("content")); err != nil {
		t.Fatal(err)
	}
	if err = writer.Close(); err != nil {
		t.Fatal(err)
	}

	target := struct {
		Name string                `form:"name" binding:"required"`
		File *multipart.FileHeader `form:"file" binding:"required"`
	}{}
	request := newRequestForTest(http.MethodPost, writer.FormDataContentType(), body)
	if err = request.BindBodyAuto(&target); err != nil {
		t.Fatal(err)
	}
	if target.Name != "acexy" || target.File == nil || target.File.Filename != "test.txt" {
		t.Fatalf("unexpected multipart binding result: %+v", target)
	}
}

func TestGetFormValueReturnsBodyLimitError(t *testing.T) {
	body := bytes.NewBufferString("name=acexy")
	request := newRequestForTest(http.MethodPost, gin.MIMEPOSTForm, body)
	request.ctx.Request.Body = http.MaxBytesReader(request.ctx.Writer, request.ctx.Request.Body, 4)

	_, _, err := request.GetFormValue("name")
	var maxBytesError *http.MaxBytesError
	if !errors.As(err, &maxBytesError) {
		t.Fatalf("expected MaxBytesError, got %v", err)
	}
	if status := requestErrorStatus(err, http.StatusBadRequest); status != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status 413, got %d", status)
	}
}

func TestBindBodyAutoUnsupportedContentType(t *testing.T) {
	request := newRequestForTest(http.MethodPost, gin.MIMEPlain, bytes.NewBufferString("value"))
	err := request.BindBodyAuto(&struct{}{})
	if !errors.Is(err, ErrUnsupportedContent) {
		t.Fatalf("expected ErrUnsupportedContent, got %v", err)
	}
	if status := requestErrorStatus(err, http.StatusBadRequest); status != http.StatusUnsupportedMediaType {
		t.Fatalf("expected status 415, got %d", status)
	}
}

func TestGetFormValueParsesOnce(t *testing.T) {
	request := newRequestForTest(http.MethodPost, gin.MIMEPOSTForm, bytes.NewBufferString("name=acexy"))
	value, ok, err := request.GetFormValue("name")
	if err != nil || !ok || value != "acexy" {
		t.Fatalf("unexpected first form value: value=%q ok=%v err=%v", value, ok, err)
	}

	request.ctx.Request.Body = http.NoBody
	value, ok, err = request.GetFormValue("name")
	if err != nil || !ok || value != "acexy" {
		t.Fatalf("unexpected cached form value: value=%q ok=%v err=%v", value, ok, err)
	}
}
