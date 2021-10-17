package tools

import (
	"bytes"
	"text/template"

	log "github.com/sirupsen/logrus"
)

func Template(obj interface{}, data string) string {
	template := template.New("template")

	f := map[string]interface{}{
		"trunc": trunc,
	}

	t, err := template.Funcs(f).Parse(data)
	if err != nil {
		log.Panic(err)
	}

	var tpl bytes.Buffer
	err = t.Execute(&tpl, obj)
	if err != nil {
		log.Panic(err)
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
