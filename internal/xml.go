package internal

import (
	"encoding/xml"
	"io"
	"log"
	"os"
	"strings"
)

// XPathItem item in the xpath search
type XPathItem struct {
	value string
	index int64
}

func (r XPathItem) begin() int64 {
	return r.index - int64(len(r.value))
}
func (r XPathItem) end() int64 {
	return r.index
}

// XPathResult is a result of the search
type XPathResult struct {
	items map[string]*XPathItem
}

// IsEmpty return true if the result is empty
func (r XPathResult) IsEmpty() bool {
	return len(r.items) == 0
}

// FindXPathInFile find the xpath items in the file
func FindXPathInFile(filename string, items []string) *XPathResult {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	data := make(map[string]bool)
	for _, x := range items {
		data[x] = true
	}

	result := &XPathResult{items: map[string]*XPathItem{}}

	path := ""
	decoder := xml.NewDecoder(file)
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
		case xml.StartElement:
			path = path + "/" + se.Name.Local
			break
		case xml.EndElement:
			path = strings.TrimSuffix(path, "/"+se.Name.Local)
			break
		case xml.CharData:
			if data[path] {
				result.items[path] = &XPathItem{value: string(se.Copy()), index: decoder.InputOffset()}
				delete(data, path)
				if len(data) == 0 {
					return result
				}
			}
			break
		}
	}
	return result
}
