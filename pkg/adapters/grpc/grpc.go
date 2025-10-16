package grpc

import "google.golang.org/grpc"

type Client struct {
	grpc.ClientConn
}
