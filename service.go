package foundation

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"contrib.go.opencensus.io/exporter/jaeger"
	"contrib.go.opencensus.io/exporter/prometheus"
	"github.com/pkg/errors"
	"go.opencensus.io/plugin/ocgrpc"
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

// WithJaegerExporter create a new export that will report trace to the given collector endpoint.
// if collector is empty, the default address will be use (http://127.0.0.1:14268/api/traces).
// will panic on error
func (s *Service) WithJaegerExporter(collectorAddr string, sampler trace.Sampler, tags ...jaeger.Tag) {
	if collectorAddr == "" {
		collectorAddr = "http://127.0.0.1:14268/api/traces"
	}

	var tt []jaeger.Tag
	if tags != nil {
		tt = append(tt, tags...)
	}
	exporter, err := jaeger.NewExporter(jaeger.Options{
		CollectorEndpoint: collectorAddr,
		Process: jaeger.Process{
			ServiceName: s.name,
			Tags:        tt,
		},
	})
	if err != nil {
		panic(err)
	}

	trace.RegisterExporter(exporter)
	trace.ApplyConfig(trace.Config{DefaultSampler: sampler})
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

	signal.Notify(s.shutdown, os.Interrupt, syscall.SIGTERM)
	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Start starts multiplexing the listener.
	go func() {
		serverErrors <- grpcsrv.Serve(l)
	}()

	select {
	case err := <-serverErrors:
		if strings.Contains(err.Error(), "use of closed network connection") {
			return nil
		}
		return errors.Wrap(err, "server error")

	case sig := <-s.shutdown:
		// Closing listener.
		_ = l.Close()

		// Log the status of this shutdown.
		switch {
		case sig == syscall.SIGSTOP:
			return errors.New("integrity issue caused shutdown")
		}
	}

	return nil
}
