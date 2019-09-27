package grpc

import (
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"
)

// NewServer creates a gRPC server that will be by default
// recover from panic and setup for observability.
func NewServer(opts ...grpc.ServerOption) *grpc.Server {
	opts = append(opts,
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(grpc_recovery.StreamServerInterceptor())),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(grpc_recovery.UnaryServerInterceptor())),
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
	)
	srv := grpc.NewServer(opts...)
	return srv
}

// NewClient create a new gRPC client setup for observability.
func NewClient(addr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	opts = append(opts, grpc.WithStatsHandler(&ocgrpc.ClientHandler{}))
	return grpc.Dial(addr, opts...)
}
