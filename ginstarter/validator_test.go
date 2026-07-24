package ginstarter

import (
	"bytes"
	"errors"
	"net/http"
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
