//go:build windows

package main

import (
	"os/exec"
	"strconv"
)

func killProcessGroup(id int) {
	logInfo("killing process "+strconv.FormatInt(int64(id), 10)+" with 'taskkill'...", false)
	cmd := exec.Command("taskkill", "/T", "/F", "/PID", strconv.FormatInt(int64(id), 10))
	err := cmd.Run()
	if err != nil {
		logWarning("failed to kill process " + strconv.FormatInt(int64(id), 10) + ":\n\t" + err.Error())
	} else {
		logInfo("\tdone", false)
	}
}

func configProcess() {

}
