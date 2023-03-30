//go:build !windows

package main

import (
	"strconv"
	"syscall"
)

func killProcessGroup(id int) {
	logInfo("killing process "+strconv.FormatInt(int64(id), 10)+"...", false)
	err := syscall.Kill(-id, syscall.SIGKILL)
	if err != nil {
		logWarning("failed to kill process " + strconv.FormatInt(int64(id), 10) + ":\n\t" + err.Error())
	} else {
		logInfo("\tdone", false)
	}
}

func configProcess() {
	childProcess.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
