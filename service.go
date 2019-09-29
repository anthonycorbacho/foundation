package foundation

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"contrib.go.opencensus.io/exporter/jaeger"

	"contrib.go.opencensus.io/exporter/prometheus"
	"github.com/pkg/errors"
	"github.com/soheilhy/cmux"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/b3"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
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

// WithPrometheusExporter creates a new HTTP server that will provide a HTTP endpoint /metrics
// that will report application metrics.
// if addr is nil, default binding port will be :9090.
// will panic if cannot bind http to the addr.
func (s *Service) WithPrometheusExporter(addr string) {
	if addr == "" {
		addr = ":9090"
	}

	name := sanitizeName(s.name)
	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace: name,
	})
	if err != nil {
		panic(err)
	}
	// Ensure that we register it as a stats exporter.
	view.RegisterExporter(pe)

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", pe)
		panic(http.ListenAndServe(addr, mux))

	}()
}

// WithJaegerExporter create a new export that will report trace to the given collector and agent endpoint.
// if collector and/or agent addr are nil (empty) default address will be used.
// will panic on error
func (s *Service) WithJaegerExporter(collectorAddr, agentAddr string, tags ...jaeger.Tag) {
	if collectorAddr == "" {
		collectorAddr = "http://127.0.0.1:14268/api/traces"
	}

	if agentAddr == "" {
		agentAddr = "127.0.0.1:6831"
	}

	var tt []jaeger.Tag
	if tags != nil {
		tt = append(tt, tags...)
	}
	exporter, err := jaeger.NewExporter(jaeger.Options{
		CollectorEndpoint: collectorAddr,
		AgentEndpoint:     agentAddr,
		Process: jaeger.Process{
			ServiceName: s.name,
			Tags:        tt,
		},
	})
	if err != nil {
		panic(err)
	}

	trace.RegisterExporter(exporter)
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

}

func sanitizeName(name string) string {
	re := regexp.MustCompile(`\W`)
	return re.ReplaceAllString(name, "_")
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
	// TODO(anthony): add better heath check. you now what, add better support for HTTP.
	go func() {
		servemux := http.NewServeMux()
		servemux.HandleFunc("/_ready", func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, "ok") })
		_ = view.Register(
			ochttp.ServerRequestCountView,
			ochttp.ServerRequestBytesView,
			ochttp.ServerResponseBytesView,
			ochttp.ServerLatencyView,
			ochttp.ServerRequestCountByMethod,
			ochttp.ServerResponseCountByStatusCode)

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
