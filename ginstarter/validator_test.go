package ginstarter

import (
	"bytes"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type domainValidationInput struct {
	Domain string `json:"domain" binding:"required,domain"`
}

func TestDomainValidationTag(t *testing.T) {
	registerValidators()

	validRequest := newRequestForTest(
		http.MethodPost,
		gin.MIMEJSON,
		bytes.NewBufferString(`{"domain":"api.example.com"}`),
	)
	if err := validRequest.BindBodyJSON(&domainValidationInput{}); err != nil {
		t.Fatalf("expected valid domain to pass, got %v", err)
	}

	invalidRequest := newRequestForTest(
		http.MethodPost,
		gin.MIMEJSON,
		bytes.NewBufferString(`{"domain":"invalid-domain"}`),
	)
	err := invalidRequest.BindBodyJSON(&domainValidationInput{})
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		t.Fatalf("expected validator.ValidationErrors, got %T: %v", err, err)
	}
	if len(validationErrors) != 1 || validationErrors[0].Tag() != "domain" {
		t.Fatalf("expected domain validation tag, got %v", validationErrors)
	}
	if message := friendlyValidatorMessage(validationErrors); message != "domain mismatch type domain" {
		t.Fatalf("unexpected friendly validation message: %q", message)
	}
}

func TestDomainValidationBoundaries(t *testing.T) {
	registerValidators()
	label63 := strings.Repeat("a", 63)
	maxDomain := label63 + "." + label63 + "." + label63 + "." + strings.Repeat("a", 61)
	for _, tt := range []struct {
		name  string
		value string
		valid bool
	}{
		{"ordinary", "api.example.com", true},
		{"uppercase", "API.EXAMPLE.COM", true},
		{"localhost", "LOCALHOST", true},
		{"hyphen and digit", "api-2.example.com", true},
		{"punycode label", "xn--bcher-kva.example", true},
		{"63-byte label", label63 + ".com", true},
		{"253-byte name", maxDomain, true},
		{"empty", "", false},
		{"single label", "example", false},
		{"leading hyphen", "-example.com", false},
		{"trailing hyphen", "example-.com", false},
		{"prefix leading hyphen", "-api.example.com", false},
		{"prefix trailing hyphen", "api-.example.com", false},
		{"underscore", "a_.example.com", false},
		{"empty label", "api..example.com", false},
		{"leading dot", ".example.com", false},
		{"trailing dot", "example.com.", false},
		{"64-byte label", strings.Repeat("a", 64) + ".com", false},
		{"254-byte name", maxDomain + "a", false},
		{"Unicode", "例子.com", false},
		{"whitespace", " example.com", false},
		{"numeric suffix", "example.123", false},
		{"short suffix", "example.c", false},
		{"IP address", "127.0.0.1", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			request := newRequestForTest(http.MethodGet, "", &bytes.Buffer{})
			request.ctx.Request.URL.RawQuery = url.Values{"Domain": []string{tt.value}}.Encode()
			err := request.BindQueryParams(&domainValidationInput{})
			if tt.valid && err != nil {
				t.Fatalf("valid hostname rejected: %v", err)
			}
			if !tt.valid {
				var validationErrors validator.ValidationErrors
				if !errors.As(err, &validationErrors) {
					t.Fatalf("invalid hostname accepted or wrong error: %v", err)
				}
			}
		})
	}
}
