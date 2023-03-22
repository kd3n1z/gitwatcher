package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Cmd string `yaml:"cmd"`
}

const VERSION string = "1.2.0";
var logEverything bool = false;

var childProcess *exec.Cmd;

func main() {
	var configOsPath string = filepath.Join(".gitwatcher", "config-" + runtime.GOOS + ".yml");
	var configPath string = filepath.Join(".gitwatcher", "config.yml");

	var interval uint64 = 60;

	var args[] string = os.Args[1:];

	for i := 0; i < len(args); i++ {
		switch args[i] {
			case "-i", "--interval":
				i += 1;
				parsedInterval, err := strconv.ParseUint(args[i], 10, 64);
				if err != nil || parsedInterval <= 0 {
					logError(args[i] + ": invalid interval");
					return;
				}
				interval = parsedInterval;
				break;
			case "-v", "--version":
				logInfo("gitwatcher v" + VERSION, true);
				return;
			case "-l", "--log-everything":
				logEverything = true;
				break;
			default:
				logError(args[i] + ": invalid option");
				return;
		}
	}

	if runtime.GOOS != "linux" && runtime.GOOS != "windows" && runtime.GOOS != "darwin" {
		logWarning("gitwatcher has not been tested on " + runtime.GOOS + ", use at your own risk");
	}

	cmd := exec.Command("git", "--version");

    var out bytes.Buffer;
    cmd.Stdout = &out;

    err := cmd.Run();
	
	if err != nil {
        logError("'" + err.Error() + "', do you have git installed?");
		return;
    }

	logInfo("gitwatcher v" + VERSION + ", " + trim(out.String()) + "\n\t- pull interval: " + strconv.FormatUint(interval, 10) + " (seconds)\n\t- platform: " + runtime.GOOS, true);
	logInfo("\t- loglevel: everything", false);

	var firstCheck bool = true;

	for {
		logInfo("running 'git pull'...", false);
		cmd := exec.Command("git", "pull");

		var out bytes.Buffer;
		var stdErr bytes.Buffer;
		cmd.Stdout = &out;
		cmd.Stderr = &stdErr;

		err := cmd.Run();

		var outStr string = strings.ToLower(trim(out.String()) + trim(stdErr.String()));
		
		if err != nil {
			logError(err.Error() + ", '" + outStr + "'");
			return;
		}
		if strings.HasPrefix(outStr, "fatal:") || strings.HasPrefix(outStr, "error:") {
			logError("'" + outStr + "'");
			return;
		}
		logInfo("\t'" + outStr + "'", false);

		if firstCheck || (outStr != "already up to date." && len(outStr) > 0) {
			logInfo(time.Now().Format("15:04") + " - restarting... ", true);

			var currentConfigPath string = configPath;
	
			if _, err := os.Stat(configOsPath); err == nil {
				currentConfigPath = configOsPath;
			}
	
			if _, err := os.Stat(currentConfigPath); err == nil {
				if childProcess != nil {
					logInfo("\tkilling previous process...", false);

					killProcessGroup(childProcess.Process.Pid);
				}

				logInfo("\treading '" + currentConfigPath + "'...", false);
				data, err := os.ReadFile(currentConfigPath);
				if(err != nil) {
					logError("'" + err.Error() + "'");
					return;
				}
	
				logInfo("\tparsing...", false);
	
				cfg := Config{};
				err = yaml.Unmarshal(data, &cfg);
	
				if(err != nil) {
					logError("'" + err.Error() + "'");
					return;
				}
				
				logInfo("\trunning ", false);

				childProcess = exec.Command("bash", "-c", cfg.Cmd);
				childProcess.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

				err = childProcess.Start();
	
				if(err != nil) {
					logWarning("failed to start: '" + err.Error() + "'");
				}
			}else{
				logInfo("\tconfig.yml not found :(", true);
			}
		}

		firstCheck = false;
		time.Sleep(time.Duration(interval) * time.Second);
	}
}

func killProcessGroup(id int) {
	err := syscall.Kill(-id, syscall.SIGKILL);

	if err != nil {
		logWarning("failed to kill process " + strconv.FormatInt(int64(id), 10));
	}
}

func trim(str string)(string) {
	return strings.Trim(str, "\n\t\t ");
}

func logInfo(str string, important bool) {
	if important || logEverything {
		fmt.Println(str);
	}
}

func logError(str string) {
	fmt.Println(color.RedString("error: ") + str);
}

func logWarning(str string) {
	fmt.Println(color.YellowString("warning: ") + str);
}
