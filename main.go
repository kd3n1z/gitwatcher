package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
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
	Shell string `yaml:"shell"`
	Args []string `yaml:"args"`
}

const VERSION string = "1.2.0";
var logEverything bool = false;
var strictMode bool = false;

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
			case "-s", "--strict":
				strictMode = true;
				break;
			default:
				logError(args[i] + ": invalid option");
				return;
		}
	}
	
	var shellArgs []string = []string{"-c", "$cmd"};
	var shell string = "sh";

	if runtime.GOOS == "windows" {
		shellArgs = []string{"/C$cmd"};
		shell = "cmd";
	} else if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
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

	logInfo("gitwatcher v" + VERSION + ", " + trim(out.String()) + "\n\t- pull interval: " + strconv.FormatUint(interval, 10) + " (seconds)\n\t- platform: " + runtime.GOOS + " (" + shell + " " + strings.Join(shellArgs, " ")  + ")", true);
	if strictMode {
		logInfo("\t- strict mode: enabled", true);
	}
	logInfo("\t- log level: everything", false);

	c := make(chan os.Signal);
    signal.Notify(c, os.Interrupt, syscall.SIGTERM);
    go func() {
        <-c
		logInfo("cleaning up...", false);
		if childProcess != nil {
			killProcessGroup(childProcess.Process.Pid);
		}
        os.Exit(1);
    }();

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
			strictError(err.Error() + ", '" + outStr + "'");
		}
		if strings.HasPrefix(outStr, "fatal:") || strings.HasPrefix(outStr, "error:") {
			strictError("'" + outStr + "'");
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
	
				cfg := Config{Shell: shell};
				cfg.Args = make([]string, len(shellArgs));
				for i, value := range shellArgs {
					cfg.Args[i] = value;
				}
				err = yaml.Unmarshal(data, &cfg);
				
				logInfo(cfg.Cmd + "\n[" + strings.Join(cfg.Args, ",") + "]", true);
	
				if(err != nil) {
					strictError("parse error: '" + err.Error() + "'");
				}

				for i := 0; i < len(cfg.Args); i++ {
					cfg.Args[i] = strings.ReplaceAll(cfg.Args[i], "$cmd", cfg.Cmd);
				}
				
				logInfo("\trunning '" + cfg.Shell + " " + strings.Join(cfg.Args, " ") + "'", false);

				childProcess = exec.Command(cfg.Shell, cfg.Args...);
				childProcess.SysProcAttr = &syscall.SysProcAttr{Setpgid: true};

				err = childProcess.Start();
	
				if(err != nil) {
					strictError("failed to start: '" + err.Error() + "'");
				}
			}else{
				strictError("config.yml not found");
			}
		}

		firstCheck = false;
		time.Sleep(time.Duration(interval) * time.Second);
	}
}

func killProcessGroup(id int) {
	err := syscall.Kill(-id, syscall.SIGKILL);

	logInfo("killing process " + strconv.FormatInt(int64(id), 10) + "...", false);
	if err != nil {
		logWarning("failed to kill process " + strconv.FormatInt(int64(id), 10) + ":\n\t" + err.Error());
	}else{
		logInfo("\tdone", false);
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

func strictError(str string) {
	if strictMode {
		logError(str);
		os.Exit(1);
	}else{
		logWarning(str);
	}
}