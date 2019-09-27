package foundation

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/soheilhy/cmux"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/b3"
	"go.opencensus.io/stats/view"
	"google.golang.org/grpc"
)

// Service is a structure that wraps the lower level libraries.
// Its a convenience method for building and initialising services.
type Service struct {
	name string
	addr string

	// shutdown channel to listen for an interrupt or terminate signal from the OS.
	shutdown chan os.Signal
}

// NewService creates a foundation Service.
func NewService(addr string, opts ...Option) *Service {
	s := Service{
		name:     os.Getenv("SERVICE_NAME"),
		addr:     addr,
		shutdown: make(chan os.Signal, 1),
	}

	for _, opt := range opts {
		opt(&s)
	}

	return &s
}

// Serve starts foundation Service.
func (s *Service) Serve(grpcsrv *grpc.Server) error {
	if grpcsrv == nil {
		return errors.New("gRPC server is required")
	}
	if err := view.Register(ocgrpc.DefaultServerViews...); err != nil {
		return errors.Wrap(err, "server error")
	}

	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	// Create multiplexer for gRPC and HTTP.
	mux := cmux.New(l)
	httpMux := mux.Match(cmux.HTTP1Fast())
	grpcMux := mux.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))

	signal.Notify(s.shutdown, os.Interrupt, syscall.SIGTERM)

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Start gRPC server that will be accepting incoming connections on the listener.
	go func() {
		serverErrors <- grpcsrv.Serve(l)
	}()

	// Start HTTP server that will be accepting incoming connections on the listener.
	// TODO(anthony): add prometheus metrics.
	// TODO(anthony): add better heath check. you now what, add better support for HTTP.
	go func() {
		servemux := http.NewServeMux()
		servemux.HandleFunc("/_ready", func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, "ok") })

		httpsrv := http.Server{
			Handler: &ochttp.Handler{
				Handler:     servemux,
				Propagation: &b3.HTTPFormat{},
			},
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		}
		serverErrors <- httpsrv.Serve(httpMux)
	}()

	// Start starts multiplexing the listener.
	go func() {
		serverErrors <- mux.Serve()
	}()

	select {
	case err := <-serverErrors:
		if strings.Contains(err.Error(), "use of closed network connection") {
			return nil
		}
		return errors.Wrap(err, "server error")

	case sig := <-s.shutdown:
		// Closing all mux and listener.
		_ = httpMux.Close()
		_ = grpcMux.Close()
		_ = l.Close()

		// Log the status of this shutdown.
		switch {
		case sig == syscall.SIGSTOP:
			return errors.New("integrity issue caused shutdown")
		}
	}

	return nil
}
