package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

func jsonPathInFile(filename string, items []string) *XPathResult {
	file, err := os.Open(filename)
	if file != nil {
		defer func() {
			if err := file.Close(); err != nil {
				log.Panic(err)
			}
		}()
	}

	if err != nil {
		log.Panic(err)
	}
	if err != nil {
		log.Panic(err)
	}

	data := make(map[string]bool)
	for _, x := range items {
		data[x] = true
	}

	result := &XPathResult{items: map[string]*XPathItem{}}

	path := ""
	value := ""
	item := ""
	key := true
	decoder := json.NewDecoder(file)
	for {
		t, err := decoder.Token()

		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("Error decoding token: %s", err)
			break
		} else if t == nil {
			break
		}
		switch se := t.(type) {
		case json.Delim:
			tmp := se.String()
			if tmp == "{" {
				key = true
				path = path + value + "/"
			} else if tmp == "}" {
				path = path[0:strings.LastIndex(path, "/")]
				path = path[0 : strings.LastIndex(path, "/")+1]
			} else if tmp == "]" {
				key = true
			} else if tmp == "[" {
				key = false
			}
		default:
			value = fmt.Sprintf("%v", se)
			if key {
				item = path + value
				key = false
			} else {
				if data[item] {
					result.items[item] = &XPathItem{value: value, index: decoder.InputOffset() - 1}
				}
				key = true
			}
		}

		if len(result.items) >= len(items) {
			break
		}
	}
	return result
}
