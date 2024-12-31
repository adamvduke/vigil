package watcher

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/adamvduke/vigil/proto/vigilpb"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedWatcherServer
	watcher *watcher
}

func (s *server) AddWatch(_ context.Context, in *pb.AddWatchRequest) (*pb.AddWatchReply, error) {
	paths, err := s.watcher.watch(in.GetPath())
	if err != nil {
		return nil, err
	}
	msg := "ack: " + in.GetPath()
	return &pb.AddWatchReply{Message: &msg, WatchedPaths: paths}, nil
}

func (s *server) WatchedPaths(_ context.Context, in *pb.WatchedPathsRequest) (*pb.WatchedPathsReply, error) {
	paths := s.watcher.watchedPaths()
	return &pb.WatchedPathsReply{Paths: paths}, nil
}

func Start(listenPath string, cwd bool, pollDuration time.Duration, args []string) {
	runner := newProcessRunner(args)
	watcher := newWatcher(runner)
	if cwd {
		dir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		if _, err := watcher.watch(dir); err != nil {
			log.Fatal(err)
		}
	}
	go func(d time.Duration) {
		if err := watcher.startPolling(d); err != nil {
			log.Fatal(err)
		}
	}(pollDuration)

	serve(listenPath, watcher)
}

func serve(listenPath string, watcher *watcher) {
	// Cleanup the unix socket on exit.
	exitChan := makeExitChan(listenPath)
	defer close(exitChan)

	listener, err := net.Listen("unix", listenPath)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	s := grpc.NewServer()
	pb.RegisterWatcherServer(s, &server{watcher: watcher})
	log.Printf("Listening at: %s, process id: %d\n", listener.Addr(), os.Getpid())
	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func makeExitChan(listenPath string) chan os.Signal {
	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, os.Interrupt, syscall.SIGTERM)
	go func(sc chan os.Signal) {
		s := <-sc
		log.Printf("received: %v, removing %s and exiting\n", s, listenPath)
		err := os.Remove(listenPath)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}(exitChan)

	return exitChan
}
