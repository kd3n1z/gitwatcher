# gitwatcher
gitwatcher will be looking for changes in git, immediately pull them, and then restart your app (you can specify start command in .gitwatcher/config.yml)

Here's a [demo](https://github.com/KD3n1z/gitwatcher-demo)

## Usage
gitwatcher [options]

Options:
- <code>-i --interval \<seconds\></code>
    - Specify pull interval.
- <code>-l --log-everything \<true/false\></code>
    - Log each action.
- <code>-d --hide-stdout \<true/false\></code>
    - Hides child process's stdout.
- <code>-s --strict-mode \<true/false\></code>
    - Enable strict mode.
    - _Strict mode is basically "terminate, if you can". For example, if there was an error while executing "git pull", in strict mode the program will terminate, in default mode it will continue to work._
- <code>-h --help</code>
    - Print usage.
- <code>-v --version</code>
    - Print current version.
- <code>--config-path</code>
    - Print config path.
- <code>--check-for-updates</code>
    - Check for newer versions on github.
- <code>--update</code>
    - Update to a newer version.
    - _Only for Unix-like platforms._
- <code>--test</code>
    - Execute command in config.yml and exit.
- <code>--init</code>
    - Initializes .gitwatcher/config.yml.


You can also specify default gitwatcher config in gitwatcher.yml:
```yaml
# gitwatcher --config-path
log-everything: false
strict-mode: false
hide-stdout: false
check-for-updates: false
interval: 60
shell: bash
args:
  - -c
  - source ~/.nvm/nvm.sh && $cmd # just an example
```

## Installation
1. Download archive for your platform from [/releases](https://github.com/KD3n1z/gitwatcher/releases)
2. Unzip it
3. Move it to your bin directory:<br>
    - typically <code>/usr/local/bin</code> on macOS/Linux<br>
    - typically <code>C:\Windows</code> on Windows

## Building
Requirements:
- [Go](https://go.dev/)
- [Git](https://git-scm.com/)

Clone this repository and then execute these commands:<br>
- <code>mkdir dest</code>
- <code>go build -ldflags="-s -w -X main.COMMIT=$(git rev-parse HEAD) -X main.BRANCH=$(git rev-parse --abbrev-ref HEAD)" -o "dest/gitwatcher" ./src/gitwatcher/</code><br>

or, if you have make installed:<br>
- <code>make build</code>

Made with ❤️ and [~~C#~~](https://github.com/KD3n1z/gitwatcher-sharp) Go.
