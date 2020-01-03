package format

import (
	"text/template"

	"github.com/manifoldco/promptui"
)

var FuncMap = template.FuncMap{
	"bufferToString": func(b []byte) string { return string(b) },
	"shorten":        func(s string) string { return s[0:8] },
}

func ParseTemplate(body string) *template.Template {
	tpl, err := template.New("").Funcs(promptui.FuncMap).Funcs(FuncMap).Parse(body)
	if err != nil {
		panic(err)
	}
	return tpl
}
