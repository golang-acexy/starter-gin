package ginstarter

import (
	"github.com/acexy/golang-toolkit/util/slice"
	"github.com/acexy/golang-toolkit/util/str"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"regexp"
)

/**
拓展验证tag
domain: 域名验证
*/

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

// friendlyValidatorMessage 处理验证框架错误，友好展示错误信息
func friendlyValidatorMessage(errors validator.ValidationErrors) string {
	builder := str.NewBuilder()
	for i, vErr := range errors {
		// 字段名
		builder.WriteString(str.LowFirstChar(vErr.Field()))
		// 验证标签
		tag := vErr.Tag()
		if slice.Contains(typeDesc, tag) {
			builder.WriteString(" mismatch type ").WriteString(tag)
		} else {
			builder.WriteString(" ").WriteString(tag)
		}
		// 标签匹配值
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

// 自定义域名验证器

// 域名验证器
func domainValidator(fl validator.FieldLevel) bool {
	domain := fl.Field().String()
	// 定义一个简单的域名正则表达式
	regex := `^(?i:((([a-zA-Z0-9-_]+)\.)*([a-zA-Z0-9-]{1,63}\.[a-zA-Z]{2,}))|localhost)$`
	match, _ := regexp.MatchString(regex, domain)
	return match
}
