package ginstarter

import (
	"strings"

	"github.com/acexy/golang-toolkit/util/coll"
	"github.com/acexy/golang-toolkit/util/str"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Additional validation rules include the domain tag.

var typeDesc = []string{
	"email",
	"url",
	"uuid",
	"ip",
	"ipv4",
	"ipv6",
	"hexcolor",
	"rgb",
	"alpha",
	"alphanum",
	"numeric",
	"base64",
	"datetime",
	"domain",
}

// friendlyValidatorMessage produces concise validation error descriptions.
func friendlyValidatorMessage(errors validator.ValidationErrors) string {
	builder := str.NewBuilder()
	for i, vErr := range errors {
		// Describe the invalid field.
		builder.WriteString(str.LowFirstChar(vErr.Field()))
		// Describe its validation rule.
		tag := vErr.Tag()
		if coll.SliceContains(typeDesc, tag) {
			builder.WriteString(" mismatch type ").WriteString(tag)
		} else {
			builder.WriteString(" ").WriteString(tag)
		}
		// Include the rule parameter when present.
		param := vErr.Param()
		if param != "" {
			builder.WriteString(" ").WriteString(param)
		}
		if i != len(errors)-1 {
			builder.WriteString("; ")
		}
	}
	return builder.ToString()
}

func registerValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("domain", domainValidator)
	}
}

// domainValidator accepts ASCII domain names and localhost, without DNS lookup.
func domainValidator(fl validator.FieldLevel) bool {
	return validDomain(fl.Field().String())
}

func validDomain(domain string) bool {
	if strings.EqualFold(domain, "localhost") {
		return true
	}
	if len(domain) > 253 {
		return false
	}
	labels := strings.Split(domain, ".")
	if len(labels) < 2 {
		return false
	}
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for i := 0; i < len(label); i++ {
			c := label[i]
			if !asciiLetter(c) && (c < '0' || c > '9') && c != '-' {
				return false
			}
		}
	}
	// Preserve the existing alphabetic suffix contract; no registry lookup is performed.
	suffix := labels[len(labels)-1]
	if len(suffix) < 2 {
		return false
	}
	for i := 0; i < len(suffix); i++ {
		if !asciiLetter(suffix[i]) {
			return false
		}
	}
	return true
}

func asciiLetter(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}
