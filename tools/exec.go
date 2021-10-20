package tools

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/lorislab/samo/log"
)

// ExecCmdOutput execute command with output
func ExecCmdOutput(name string, arg ...string) string {
	log.Debug(name, log.F("args", strings.Join(arg, " ")))
	out, err := exec.Command(name, arg...).CombinedOutput()
	log.Debug("Output: " + string(out))
	if err != nil {
		log.Fatal("Error execute command", log.Fields{"cmd": name, "args": arg, "output": string(out)}.E(err))
	}
	return string(bytes.TrimRight(out, "\n"))
}

func ExecCmd(name string, arg ...string) {
	ExecCmdAdv(nil, name, arg...)
}

// ExecCmd execute command
func ExecCmdAdv(exclude []int, name string, arg ...string) {
	args := []string{}
	args = append(args, arg...)
	if len(exclude) > 0 {
		for _, i := range exclude {
			args[i] = "*****"
		}
	}
	log.Debug(name, log.F("args", strings.Join(args, " ")))
	fmt.Println(arg)
	cmd := exec.Command(name, arg...)

	// enable always error log for the command
	errorReader, err := cmd.StderrPipe()
	if err != nil {
		log.Panic("error setup error pipe", log.Fields{"name": name, "args": arg}.E(err))
	}
	scannerError := bufio.NewScanner(errorReader)
	go func() {
		for scannerError.Scan() {
			log.Error(scannerError.Text())
		}
	}()

	// enable info log for the command
	if log.IsDebugLevel() {
		// create a pipe for the output of the script
		cmdReader, err := cmd.StdoutPipe()
		if err != nil {
			log.Panic("error setup output pipe", log.Fields{"name": name, "args": arg}.E(err))
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
		log.Panic("error start command", log.Fields{"name": name, "args": arg}.E(err))
	}

	err = cmd.Wait()
	if err != nil {
		log.Fatal("Error execute command", log.E(err))
	}
}

func execCmdErr(name string, arg ...string) error {
	log.Debug(name, log.F("args", strings.Join(arg, " ")))
	out, err := exec.Command(name, arg...).CombinedOutput()
	log.Debug("Output: " + string(out))
	if err != nil {
		log.Fatal("Error execute command", log.Fields{"cmd": name, "args": arg, "output": string(out)}.E(err))
	}
	return err
}

func CmdOutputErr(name string, arg ...string) (string, error) {
	log.Debug(name, log.F("args", strings.Join(arg, " ")))
	out, err := exec.Command(name, arg...).CombinedOutput()
	log.Debug("Output: " + string(out))
	return string(bytes.TrimRight(out, "\n")), err
}
