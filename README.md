# vigil

_A file monitor that runs a command on file changes_

I originally named the WIP `vigilo` from the latin, meaning
`to watch through, spend in watching, do or make while watching`, but then I
found [vigilo](https://github.com/wycats/vigilo) from [wycats](https://github.com/wycats)
that does virtually the same thing. I changed the name to `vigil` based on the
name of one of my favorite [Lamb of God songs](https://www.youtube.com/watch?v=lxgelwqe8-E).

## Install

- Make sure you have [Go](https://golang.org/doc/install) installed.
- `go install github.com/adamvduke/vigil@v0.0.1`

## Usage
```
$ vigil --help
Usage of vigil:
  -client
    	if vigil should operate as a client rather than server/watcher
  -cwd
    	if vigil should watch the current working directory (default true)
  -exclude value
    	a path component to exclude from the list of currently watched files, can be used multiple times (default .git,.svn,.hg)
  -listen_path string
    	path to the unix socket where vigil will listen for commands (default "/tmp/vigil.sock")
  -path string
    	a path to add to the list of currently watched files, only used when operating as a client
  -poll
    	if vigil should poll for changes rather than use inotify
  -poll_interval duration
    	time interval between polling operations, accepts a value parseable by time.ParseDuration, e.g. 5s, 300ms, etc... https://pkg.go.dev/time#ParseDuration (default 5s)
  -version
    	print the version of vigil and exit
```

## Generate the proto stuff

1. Install `protoc`
1. Install `protoc-gen-go`
1. Install `protoc-gen-go-grpc`
1. From the root of the vigil project directory
    `$ go generate`

## How it works
By default the main entry point to the program starts a server and a watcher
that uses file system notifications to recursively watch the current working
directory. Paths containing ".git", ".svn", and ".hg" are excluded. Additional
paths can be excluded via the `-exclude` flag, which can be passed multiple
times.

There are CLI flags to disable watching the current working directory, and
also switch to polling instead of file system notifications.

The main entrypoint also provides a flag to run the program as a client, which
can either make a request to add a path to the list of watched files, or
alternatively return the list of watched files without adding any new paths.
If a directory is passed, it is watched recursively.

## Similar Projects
* [https://github.com/guard/guard](https://github.com/guard/guard)
* [https://github.com/alexch/rerun](https://github.com/alexch/rerun)
* [https://github.com/mynyml/watchr](https://github.com/mynyml/watchr)
* [https://github.com/eaburns/Watch](https://github.com/eaburns/Watch)
* [https://github.com/alloy/kicker](https://github.com/alloy/kicker)
* [https://github.com/eradman/entr](https://github.com/eradman/entr)
* [https://github.com/cespare/reflex](https://github.com/cespare/reflex)
* [https://github.com/mlbileschi/file_system_monitor](https://github.com/mlbileschi/file_system_monitor)
