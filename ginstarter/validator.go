package ginstarter

import (
	"github.com/acexy/golang-toolkit/util/slice"
	"github.com/acexy/golang-toolkit/util/str"
	"github.com/go-playground/validator/v10"
)

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
			builder.WriteString(" type ").WriteString(tag)
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
