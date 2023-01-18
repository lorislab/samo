package tools

import (
	"bytes"
	"text/template"

	"github.com/lorislab/samo/log"
)

func Template(obj interface{}, data string) string {
	temp := template.New("template")

	f := map[string]interface{}{
		"trunc": trunc,
	}

	t, err := temp.Funcs(f).Parse(data)
	if err != nil {
		log.Panic("error parse template data", log.E(err))
	}

	var tpl bytes.Buffer
	err = t.Execute(&tpl, obj)
	if err != nil {
		log.Panic("error execute template", log.E(err))
	}
	return tpl.String()
}

func trunc(c int, s string) string {
	if c < 0 && len(s)+c > 0 {
		return s[len(s)+c:]
	}
	if c >= 0 && len(s) > c {
		return s[:c]
	}
	return s
}
