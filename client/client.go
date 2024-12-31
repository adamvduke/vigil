package client

import (
	"context"
	"errors"
	"os"
	"time"

	pb "github.com/adamvduke/vigil/proto/vigilpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	errPathEmpty = errors.New("path can not be empty")
)

type WatcherClient struct {
	Addr string
}

func (client *WatcherClient) AddWatch(path string) ([]string, error) {
	if path == "" {
		return nil, errPathEmpty
	}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	conn, err := grpc.NewClient(client.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	c := pb.NewWatcherClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.AddWatch(ctx, &pb.AddWatchRequest{Path: &path})
	if err != nil {
		return nil, err
	}

	return r.GetWatchedPaths(), nil
}

func (client *WatcherClient) WatchedPaths() ([]string, error) {
	conn, err := grpc.NewClient(client.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	c := pb.NewWatcherClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.WatchedPaths(ctx, &pb.WatchedPathsRequest{})
	if err != nil {
		return nil, err
	}

	return r.GetPaths(), nil
}
