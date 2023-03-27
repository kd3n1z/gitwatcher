# gitwatcher
gitwatcher will be looking for changes in git, immediately pull them, and then restart your app (you can specify start command in .gitwatcher/config.yml)

Here's a [demo](https://github.com/KD3n1z/gitwatcher-demo)

## Usage
gitwatcher [options]

Options:
- <code>-i --interval \<seconds\></code>
    - Specify pull interval.
- <code>-l --log-everything</code>
    - Log each action.
- <code>-h --help</code>
    - Print usage.
- <code>-v --version</code>
    - Print current version.
- <code>-s --strict-mode</code>
    - Enable strict mode.
    - _Strict mode is basically "terminate, if you can". For example, if there was an error when executing "git pull", in strict mode the program will terminate, in default mode it will continue to work._
- <code>--check-for-updates</code>
    - Check for newer versions on github.
- <code>--update</code>
    - Update to a newer version.
    - _Only for Unix-like systems._
- <code>--init</code>
    - Initializes .gitwatcher/config.yml.
        

## Installing
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
- <code>go build -ldflags="-s -w -X main.COMMIT=$(git rev-parse HEAD) -X main.BRANCH=$(git rev-parse --abbrev-ref HEAD)" -o "dest/gitwatcher"</code><br>

or, if you have make installed:<br>
- <code>make build</code>

Made with ❤️ and [~~C#~~](https://github.com/KD3n1z/gitwatcher-sharp) Go.
