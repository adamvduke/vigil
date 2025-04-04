package watcher

import (
	"context"
	"log"
	"net"
	"os"
	"syscall"
	"time"

	"github.com/vrecan/death/v3"

	pb "github.com/adamvduke/vigil/proto/vigilpb"

	"google.golang.org/grpc"
)

type Config struct {
	ListenPath   string
	Cwd          bool
	Poll         bool
	PollDuration time.Duration
	Excludes     []string
	CmdArgs      []string
}

type watcher interface {
	// watcher lifecycle
	start() error
	stop() error

	// server interactions
	addPath(string) ([]string, error)
	watchedPaths() []string
}

type server struct {
	pb.UnimplementedWatcherServer
	watcher watcher
}

func (s *server) AddWatch(_ context.Context, in *pb.AddWatchRequest) (*pb.AddWatchReply, error) {
	paths, err := s.watcher.addPath(in.GetPath())
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

func Start(config *Config) {
	ch := newProcessRunner(config.CmdArgs)
	var watcher watcher
	if config.Poll {
		watcher = newPollingWatcher(config, ch)
	} else {
		watcher = newNotifyWatcher(config, ch)
	}
	watcher.start()
	defer watcher.stop()
	if config.Cwd {
		dir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		if _, err := watcher.addPath(dir); err != nil {
			log.Fatal(err)
		}
	}

	serve(config.ListenPath, watcher)
}

func serve(listenPath string, watcher watcher) {
	listener, err := net.Listen("unix", listenPath)
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	pb.RegisterWatcherServer(s, &server{watcher: watcher})
	cleanup := prepareShutdown(s, listenPath)
	log.Printf("Listening at: %s, process id: %d\n", listener.Addr(), os.Getpid())
	if err := s.Serve(listener); err != nil {
		cleanup.FallOnSword()
		log.Fatalf("failed to serve: %v", err)
	}
}

func prepareShutdown(rpcServer *grpc.Server, listenPath string) *death.Death {
	cleanup := death.NewDeath(syscall.SIGINT, syscall.SIGTERM)
	go cleanup.WaitForDeathWithFunc(func() {
		log.Println("stopping grpc server...")
		rpcServer.GracefulStop()

		// Cleanup the unix socket on exit.
		log.Printf("removing %s...", listenPath)
		if err := os.Remove(listenPath); err != nil {
			log.Fatal(err)
		}
	})

	return cleanup
}
