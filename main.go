package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/adamvduke/vigil/client"
	"github.com/adamvduke/vigil/watcher"
)

const (
	defaultPollInterval = 2
)

func main() {
	// flags only used when run as a server
	listenPath := flag.String("listen_path", "/tmp/vigil.sock", "path to the unix socket where vigil will listen for commands")
	pollInterval := flag.Int("poll_interval", defaultPollInterval, "seconds between polling the file system for changes")
	cwd := flag.Bool("cwd", true, "if vigil should watch the current working directory")

	// flags only used when run as a client
	runAsClient := flag.Bool("client", false, "if vigil should operate as a client rather than server/watcher")
	path := flag.String("path", "", "a path for vigil to watch, only used when operating as a client")

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
	pollDuration, err := time.ParseDuration(fmt.Sprintf("%ds", *pollInterval))
	if err != nil {
		log.Fatal(err)
	}
	watcher.Start(*listenPath, *cwd, pollDuration, flag.Args())
}
