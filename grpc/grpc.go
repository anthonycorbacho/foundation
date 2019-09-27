package grpc

import (
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"
)

// NewServer creates a gRPC server that will be by default
// recover from panic and setup for observability.
func NewServer(opts ...grpc.ServerOption) *grpc.Server {
	opts = append(opts,
		grpc.StreamInterceptor(grpcmiddleware.ChainStreamServer(grpcrecovery.StreamServerInterceptor())),
		grpc.UnaryInterceptor(grpcmiddleware.ChainUnaryServer(grpcrecovery.UnaryServerInterceptor())),
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
