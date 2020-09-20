package internal

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// WriteBytesToFile writes the byte array into the file
func WriteBytesToFile(filename string, data []byte) {
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

	_, err = w.Write(data)
	if err != nil {
		panic(err)
	}
	err = w.Flush()
	if err != nil {
		panic(err)
	}
}

// WriteToFile writes the data into the file
func WriteToFile(filename, data string) {
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

// ReplaceTextInFile replaces test in the file at the position b and e
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

// GetAllFilePathsInDirectory get list of all files in the directory
func GetAllFilePathsInDirectory(dir string) ([]string, error) {
	var paths []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return paths, nil
}
