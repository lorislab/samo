package internal

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
	log "github.com/sirupsen/logrus"
)

// Writes the data into the file
func WriteToFile(filename , data string) {
	dir := filepath.Dir(filename)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		panic(err)
	}

	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	w := bufio.NewWriter(file)

	_, err = w.WriteString(data)
	if err != nil {
		panic(err)
	}
	err = w.Flush()
	if err != nil {
		panic(err)
	}
}

// Replace test in the file at the position b and e
func ReplaceTextInFile(filename, text string, b, e int64) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Panic(err)
	}
	result := string(buf)
	result = result[:b] + text + result[e:]
	err = ioutil.WriteFile(filename, []byte(result), 0666)
	if err != nil {
		log.Panic(err)
	}
}
