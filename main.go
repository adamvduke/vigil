package main

import (
	"flag"
	"log"
	"strings"
	"time"

	"github.com/adamvduke/vigil/client"
	"github.com/adamvduke/vigil/watcher"
)

//go:generate protoc --go_out=proto/ --go-grpc_out=proto/ --proto_path=proto/ proto/vigil.proto

const (
	defaultPollInterval = 5 * time.Second
)

type excludes []string

func (e *excludes) String() string {
	return strings.Join(*e, ",")
}

func (e *excludes) Set(value string) error {
	*e = append(*e, value)
	return nil
}

func main() {
	// flags only used when run as a server
	excludes := excludes{".git", ".svn", ".hg"}
	flag.Var(&excludes, "exclude", "a path component to exclude from the list of currently watched files, can be used multiple times")
	listenPath := flag.String("listen_path", "/tmp/vigil.sock", "path to the unix socket where vigil will listen for commands")
	poll := flag.Bool("poll", false, "if vigil should poll for changes rather than use inotify")
	pollDuration := flag.Duration("poll_interval", defaultPollInterval,
		"time interval between polling operations, accepts a value parseable by time.ParseDuration, e.g. 5s, 300ms, etc... "+
			"https://pkg.go.dev/time#ParseDuration")
	cwd := flag.Bool("cwd", true, "if vigil should watch the current working directory")

	// flags only used when run as a client
	runAsClient := flag.Bool("client", false, "if vigil should operate as a client rather than server/watcher")
	path := flag.String("path", "", "a path to add to the list of currently watched files, only used when operating as a client")

	flag.Parse()

	if *runAsClient {
		wClient := &client.WatcherClient{Addr: "unix:" + *listenPath}
		var paths []string
		var err error
		if *path != "" {
			paths, err = wClient.AddWatch(*path)
		} else {
			paths, err = wClient.WatchedPaths()
		}
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Listening for changes to", paths)

		return
	}

	if len(flag.Args()) == 0 {
		log.Fatal("must provide a program to run")
	}
	c := &watcher.Config{
		ListenPath:   *listenPath,
		Cwd:          *cwd,
		Poll:         *poll,
		PollDuration: *pollDuration,
		Excludes:     excludes,
		CmdArgs:      flag.Args(),
	}
	watcher.Start(c)
}
