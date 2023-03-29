package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	"github.com/hashicorp/go-version"
	"gopkg.in/yaml.v3"
)

type GitwatcherConfig struct {
	LogEverything bool     `yaml:"log-everything"`
	StrictMode    bool     `yaml:"strict-mode"`
	HideStdout    bool     `yaml:"hide-stdout"`
	Interval      uint64   `yaml:"interval"`
	Shell         string   `yaml:"shell"`
	ShellArgs     []string `yaml:"args"`
}

type RepoConfig struct {
	Cmd   string   `yaml:"cmd"`
	Shell string   `yaml:"shell"`
	Args  []string `yaml:"args"`
}

type GithubResp struct {
	TagName string `json:"tag_name"`
	HtmlUrl string `json:"html_url"`
	Assets  []struct {
		DownloadUrl string `json:"browser_download_url"`
		Name        string `json:"name"`
	} `json:"assets"`
}

var BRANCH string = "?" //-ldflags
var COMMIT string = "?" //-ldflags

const VERSION string = "1.2.4"

var childProcess *exec.Cmd

var gwConfig GitwatcherConfig = GitwatcherConfig{Interval: 60, LogEverything: false, StrictMode: false, HideStdout: false}

// This is my first program written in go, so it may be unstable. Better use --strict-mode

// TODO: webhook-mode
func main() {
	gwConfig.ShellArgs = []string{"-c", "$cmd"}
	gwConfig.Shell = "bash"

	if runtime.GOOS == "windows" {
		gwConfig.ShellArgs = []string{"/C$cmd"}
		gwConfig.Shell = "cmd.exe"
	} else if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		logWarning("gitwatcher has not been tested on " + runtime.GOOS + ", use at your own risk")
	}

	cfgPath, err := os.UserConfigDir()

	if err == nil {
		os.MkdirAll(cfgPath, os.ModePerm)
		cfgPath = filepath.Join(cfgPath, "gitwatcher.yml")

		if _, err := os.Stat(cfgPath); err == nil {
			data, err := os.ReadFile(cfgPath)
			if err == nil {
				err = yaml.Unmarshal(data, &gwConfig)

				if err != nil {
					logWarning("config parse error: '" + err.Error() + "'")
				}
			} else {
				logWarning("config read error: '" + err.Error() + "'")
			}
		}
	}

	var configPath string = filepath.Join(".gitwatcher", "config.yml")

	var args []string = os.Args[1:]

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-i", "--interval":
			i += 1
			parsedInterval, err := strconv.ParseUint(args[i], 10, 64)
			if err != nil || parsedInterval <= 0 {
				logError(args[i] + ": invalid interval")
				return
			}
			gwConfig.Interval = parsedInterval
			break
		case "-h", "--help":
			logInfo("gitwatcher v"+VERSION+"\n\nUsage: gitwatcher [options]\n\t-i --interval <seconds>\t\t Specify pull interval.\n\t-d --hide-stdout <true/false>\t Hides child process's stdout.\n\t-s --strict-mode <true/false>\t Enable strict mode.\n\t-l --log-everything <true/false> Log each action.\n\t-h --help\t\t\t Print usage.\n\t-v --version\t\t\t Print current version.\n\t--config-path\t\t\t Print config path.\n\t--check-for-updates\t\t Check for newer versions on github.\n\t--update\t\t\t Update to a newer version.\n\t--init\t\t\t\t Initialize .gitwatcher/config.yml.", true)
			return
		case "-v", "--version":
			logInfo("gitwatcher v"+VERSION+", "+BRANCH+"/"+COMMIT, true)
			return
		case "-l", "--log-everything":
			if i+1 < len(args) && (args[i+1] == "true" || args[i+1] == "false") {
				i += 1
				gwConfig.LogEverything = args[i] == "true"
			} else {
				gwConfig.LogEverything = true
				logWarning(args[i] + ": better use '" + args[i] + " true'")
			}
			break
		case "-s", "--strict-mode":
			if i+1 < len(args) && (args[i+1] == "true" || args[i+1] == "false") {
				i += 1
				gwConfig.StrictMode = args[i] == "true"
			} else {
				gwConfig.StrictMode = true
				logWarning(args[i] + ": better use '" + args[i] + " true'")
			}
			break
		case "--check-for-updates":
			checkForUpdates(false)
			return
		case "--update":
			checkForUpdates(true)
			return
		case "--config-path":
			logInfo(cfgPath, true)
			return
		case "--init":
			if _, err := os.Stat(".gitwatcher"); err != nil {
				os.Mkdir(".gitwatcher", os.ModePerm)
				logInfo("created .gitwatcher", true)
			}

			if _, err := os.Stat(configPath); err != nil {
				f, _ := os.Create(configPath)
				f.WriteString("# generated by gitwatcher\n\n# default config\ndefault:\n  cmd: echo hello world\n  shell: sh\n  args:\n    - -c\n    - $cmd\n\n# config for '" + runtime.GOOS + "'\n" + runtime.GOOS + ":\n  cmd: echo hello world\n")
				f.Close()
				logInfo("created "+configPath, true)
			}
			return
		case "-d", "--hide-stdout":
			if i+1 < len(args) && (args[i+1] == "true" || args[i+1] == "false") {
				i += 1
				gwConfig.HideStdout = args[i] == "true"
			} else {
				gwConfig.HideStdout = true
				logWarning(args[i] + ": better use '" + args[i] + " true'")
			}
			break
		default:
			logError(args[i] + ": invalid option")
			return
		}
	}

	cmd := exec.Command("git", "--version")

	var out bytes.Buffer
	cmd.Stdout = &out

	err = cmd.Run()

	if err != nil {
		logError("'" + err.Error() + "', do you have git installed?")
		return
	}

	logInfo("gitwatcher v"+VERSION+", "+trim(out.String())+"\n\t- pull interval: "+strconv.FormatUint(gwConfig.Interval, 10)+" (seconds)\n\t- platform: "+runtime.GOOS+" ("+gwConfig.Shell+" "+strings.Join(gwConfig.ShellArgs, " ")+")", true)
	if gwConfig.StrictMode {
		logInfo("\t- strict mode: enabled", true)
	}
	if gwConfig.HideStdout {
		logInfo("\t- child's stdout: hidden", true)
	}
	logInfo("\t- log level: everything", false)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		logInfo("cleaning up...", false)
		if childProcess != nil {
			killProcessGroup(childProcess.Process.Pid)
		}
		os.Exit(1)
	}()

	var firstCheck bool = true

	for {
		logInfo("executing 'git pull'...", false)
		cmd := exec.Command("git", "pull")

		var out bytes.Buffer
		var stdErr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stdErr

		err := cmd.Run()

		var outStr string = strings.ToLower(trim(out.String()) + trim(stdErr.String()))

		if err != nil {
			strictError(err.Error() + ", '" + outStr + "'")
		}
		if strings.HasPrefix(outStr, "fatal:") || strings.HasPrefix(outStr, "error:") {
			strictError("'" + outStr + "'")
		}
		logInfo("\t'"+outStr+"'", false)

		if firstCheck || (outStr != "already up to date." && len(outStr) > 0) {
			logInfo(time.Now().Format("15:04")+" - restarting... ", true)
			if _, err := os.Stat(configPath); err == nil {
				if childProcess != nil {
					logInfo("\tkilling previous process...", false)

					killProcessGroup(childProcess.Process.Pid)
				}

				logInfo("\treading '"+configPath+"'...", false)
				data, err := os.ReadFile(configPath)
				if err == nil {
					logInfo("\tparsing...", false)

					parsedCfg := map[string]RepoConfig{}

					err = yaml.Unmarshal(data, &parsedCfg)

					cfg, ok := parsedCfg[runtime.GOOS]

					if !ok {
						cfg, ok = parsedCfg["default"]
					}

					if ok {
						if cfg.Shell == "" {
							cfg.Shell = gwConfig.Shell
							logInfo("\tshell not specified, using '"+gwConfig.Shell+"'", false)
						}
						if len(cfg.Args) == 0 {
							cfg.Args = make([]string, len(gwConfig.ShellArgs))
							logInfo("\tshell args not specified, using ["+strings.Join(gwConfig.ShellArgs, ", ")+"]", false)

							for i := 0; i < len(gwConfig.ShellArgs); i++ {
								cfg.Args[i] = gwConfig.ShellArgs[i]
							}
						}
						if len(cfg.Cmd) > 0 {
							for i := 0; i < len(cfg.Args); i++ {
								cfg.Args[i] = strings.ReplaceAll(cfg.Args[i], "$cmd", cfg.Cmd)
							}

							logInfo("\texecuting '"+cfg.Shell+" "+strings.Join(cfg.Args, " ")+"'", false)

							childProcess = exec.Command(cfg.Shell, cfg.Args...)

							if !gwConfig.HideStdout {
								childProcess.Stdout = os.Stdout
								childProcess.Stderr = os.Stderr
							}

							configProcess()

							err = childProcess.Start()

							if err != nil {
								strictError("failed to start: '" + err.Error() + "'")
							}
						} else {
							strictError("config.yml: cmd not specified")
						}
					} else {
						strictError("config.yml: no suitable config found (default | " + runtime.GOOS + ")")
					}
				} else {
					strictError("error reading config.yml: '" + err.Error() + "'")
				}
			} else {
				strictError("config.yml not found")
			}
		}

		firstCheck = false
		time.Sleep(time.Duration(gwConfig.Interval) * time.Second)
	}
}

func checkForUpdates(selfUpdate bool) {
	resp, err := http.Get("https://api.github.com/repos/KD3n1z/gitwatcher/releases/latest")

	if err != nil {
		logError(err.Error())
		return
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		logError(err.Error())
		return
	}

	parsedResp := GithubResp{}

	err = json.Unmarshal(body, &parsedResp)

	resp.Body.Close()

	if err != nil {
		logError("parse error: '" + err.Error() + "'")
		return
	}

	localV, _ := version.NewVersion(VERSION)
	remoteV, _ := version.NewVersion(parsedResp.TagName)

	if localV.LessThan(remoteV) {
		if selfUpdate {
			logInfo("Updating...", true)

			if runtime.GOOS == "windows" {
				logError("not supported on windows")
				return
			}

			var updateUrl string = ""

			var cOS = runtime.GOOS
			var cArch = runtime.GOARCH
			if cOS == "darwin" {
				cOS = "macos"
			}
			if cArch == "amd64" {
				cArch = "x64"
			}

			for _, asset := range parsedResp.Assets {
				var fName string = strings.ToLower(asset.Name)
				if fName == cOS+".zip" || fName == cOS+"-"+cArch+".zip" {
					updateUrl = asset.DownloadUrl
					break
				}
			}

			if updateUrl == "" {
				logError("asset " + cOS + "-" + cArch + ".zip not found")
				return
			}

			logInfo("\tcreating temp file...", true)

			file, err := os.CreateTemp("", "gwr-update")

			path := file.Name()

			logInfo("\t\t"+path, true)

			if err != nil {
				logError(err.Error())
				return
			}

			logInfo("\tdownloading '"+updateUrl+"'...", true)

			downloadResp, err := http.Get(updateUrl)

			if err != nil {
				logError(err.Error())
				return
			}

			_, err = io.Copy(file, downloadResp.Body)

			if err != nil {
				logError(err.Error())
				return
			}

			file.Close()
			downloadResp.Body.Close()

			reader, err := zip.OpenReader(path)

			if err != nil {
				logError(err.Error())
				return
			}

			logInfo("\tunzipping...", true)
			for _, f := range reader.File {
				if f.Name == "gitwatcher" {
					cPath, err := os.Executable()

					if err != nil {
						logError(err.Error())
						return
					}

					logInfo("\t\topening '"+cPath+"'...", true)
					executable, err := os.OpenFile(cPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())

					if err != nil {
						logError(err.Error())
						return
					}

					logInfo("\t\tdecompressing...", true)

					zippedFile, err := f.Open()

					if err != nil {
						logError(err.Error())
						return
					}

					_, err = io.Copy(executable, zippedFile)

					if err != nil {
						logError(err.Error())
						return
					}

					zippedFile.Close()
					executable.Close()

					break
				}
			}

			reader.Close()

			logInfo("\t\tcleaning up...", true)

			err = os.Remove(path)

			if err != nil {
				logError(err.Error())
			}

			logInfo("\tdone, try gitwatcher --version", true)
		} else {
			logInfo("Update available!\n\tv"+VERSION+" -> "+parsedResp.TagName+"\n\n"+parsedResp.HtmlUrl, true)
		}
	} else {
		logInfo("Already at the latest version.", true)
	}
}

func trim(str string) string {
	return strings.Trim(str, "\n\t\t ")
}

func logInfo(str string, important bool) {
	if important || gwConfig.LogEverything {
		fmt.Println(str)
	}
}

func logError(str string) {
	fmt.Println(color.RedString("error: ") + str)
}

func logWarning(str string) {
	fmt.Println(color.YellowString("warning: ") + str)
}

func strictError(str string) {
	if gwConfig.StrictMode {
		logError("strict: " + str)
		os.Exit(1)
	} else {
		logWarning(str)
	}
}
