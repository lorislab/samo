package yaml

import (
	"io/ioutil"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// ReplaceValueInYaml replace values in the YAML file
func ReplaceValueInYaml(filename string, data map[string]string) {

	obj := make(map[interface{}]interface{})

	fileBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Panic(err)
	}
	err = yaml.Unmarshal(fileBytes, &obj)
	if err != nil {
		log.Panic(err)
	}
	for k, v := range data {
		replace(obj, k, v)
	}

	fileBytes, err = yaml.Marshal(&obj)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	err = ioutil.WriteFile(filename, fileBytes, 0666)
	if err != nil {
		log.Panic(err)
	}
}

func replace(obj map[interface{}]interface{}, k string, v string) {

	keys := strings.Split(k, ".")
	var tmp interface{}
	size := len(keys)

	tmp = obj
	for i := 0; i < size-1; i++ {
		a := get(tmp, keys[i])
		if a == nil {
			b := make(map[interface{}]interface{})
			set(tmp, keys[i], b)
			tmp = b
		} else {
			tmp = a
		}
	}
	set(tmp, keys[size-1], v)
}

func get(this interface{}, key string) interface{} {
	return this.(map[interface{}]interface{})[key]
}

func set(this interface{}, key string, value interface{}) {
	this.(map[interface{}]interface{})[key] = value
}
