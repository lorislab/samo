package xml

import (
	"encoding/xml"
	"io"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

// XPathItem item in the xpath search
type XPathItem struct {
	Value string
	Index int64
}

func (r XPathItem) Begin() int64 {
	return r.Index - int64(len(r.Value))
}
func (r XPathItem) End() int64 {
	return r.Index
}

// XPathResult is a result of the search
type XPathResult struct {
	Items map[string]*XPathItem
}

// IsEmpty return true if the result is empty
func (r XPathResult) IsEmpty() bool {
	return len(r.Items) == 0
}

// FindXPathInFile find the xpath items in the file
func FindXPathInFile(filename string, items []string) *XPathResult {
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

	result := &XPathResult{Items: map[string]*XPathItem{}}

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
			if data[path] {
				result.Items[path] = &XPathItem{Value: se.Name.Local, Index: decoder.InputOffset()}
			}
		case xml.EndElement:
			path = strings.TrimSuffix(path, "/"+se.Name.Local)
		case xml.CharData:
			if data[path] {
				result.Items[path] = &XPathItem{Value: string(se.Copy()), Index: decoder.InputOffset()}
				delete(data, path)
				if len(data) == 0 {
					return result
				}
			}
		}
	}
	return result
}
