package format

import (
	"text/template"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/manifoldco/promptui"
)

var FuncMap = template.FuncMap{
	"humanBytes": func(n int64) string {
		return humanize.Bytes(uint64(n))
	},
	"bufferToString": func(b []byte) string { return string(b) },
	"shorten":        func(s string) string { return s[0:8] },
	"timeToDuration": func(i int64) string {
		return humanize.Time(time.Unix(i, 0))
	},
}

func ParseTemplate(body string) *template.Template {
	tpl, err := template.New("").Funcs(promptui.FuncMap).Funcs(FuncMap).Parse(body)
	if err != nil {
		panic(err)
	}
	return tpl
}
