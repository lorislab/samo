package tools

import (
	"bufio"
	"bytes"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

// ExecCmdOutput execute command with output
func ExecCmdOutput(name string, arg ...string) string {
	log.Debug(name+" ", strings.Join(arg, " "))
	out, err := exec.Command(name, arg...).CombinedOutput()
	log.Debug("Output: ", string(out))
	if err != nil {
		log.Error(string(out))
		log.WithFields(log.Fields{
			"cmd":   name,
			"args":  arg,
			"error": err,
		}).Fatal("Error execute command")
	}
	return string(bytes.TrimRight(out, "\n"))
}

// ExecCmd execute command
func ExecCmd(name string, arg ...string) {
	log.Info(name+" ", strings.Join(arg, " "))
	cmd := exec.Command(name, arg...)

	// enable always error log for the command
	errorReader, err := cmd.StderrPipe()
	if err != nil {
		log.WithFields(log.Fields{
			"name": name,
			"args": arg,
		}).Panic(err)
	}
	scannerError := bufio.NewScanner(errorReader)
	go func() {
		for scannerError.Scan() {
			log.Error(scannerError.Text())
		}
	}()

	// enable info log for the command
	if log.GetLevel() == log.DebugLevel {
		// create a pipe for the output of the script
		cmdReader, err := cmd.StdoutPipe()
		if err != nil {
			log.Panic(err)
		}

		scanner := bufio.NewScanner(cmdReader)
		go func() {
			for scanner.Scan() {
				log.Debug(scanner.Text())
			}
		}()
	}

	err = cmd.Start()
	if err != nil {
		log.Panic(err)
	}

	err = cmd.Wait()
	if err != nil {
		log.Panic(err)
	}
}

func execCmdErr(name string, arg ...string) error {
	log.Debug(name+" ", strings.Join(arg, " "))
	out, err := exec.Command(name, arg...).CombinedOutput()
	log.Debug("Output: ", string(out))
	if err != nil {
		log.Error(string(out))
		log.WithFields(log.Fields{
			"cmd":   name,
			"args":  arg,
			"error": err,
		}).Fatal("Error execute command")
	}
	return err
}

func cmdOutputErr(name string, arg ...string) (string, error) {
	log.Debug(name+" ", strings.Join(arg, " "))
	out, err := exec.Command(name, arg...).CombinedOutput()
	log.Debug("Output: ", string(out))
	return string(bytes.TrimRight(out, "\n")), err
}

func Lpad(data, pad string, length int) string {
	for i := len(data); i < length; i++ {
		data = pad + data
	}
	return data
}
