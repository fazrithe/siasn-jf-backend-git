package metricutil

import (
	"context"
	"github.com/fazrithe/siasn-jf-backend-git/libs/logutil"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

type PrometheusServer struct {
	server http.Server
	Logger logutil.Logger
}

func NewPrometheusServer(listenAddress string) *PrometheusServer {
	prometheusServer := PrometheusServer{
		server: http.Server{
			Addr: listenAddress,
		},
		Logger: logutil.NewStdLogger(logutil.IsSupportColor(), "Prometheus"),
	}

	return &prometheusServer
}

func (ps *PrometheusServer) Start() {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	ps.server.Handler = mux
	ps.Logger.Infof("Server listening to %s", ps.server.Addr)
	err := ps.server.ListenAndServe()
	if err != http.ErrServerClosed {
		ps.Logger.Errorf("cannot start prometheus server, got %v", err)
	}
}

func (ps *PrometheusServer) Shutdown() {
	err := ps.server.Shutdown(context.Background())
	if err != nil && err != http.ErrServerClosed {
		ps.Logger.Errorf("cannot shutdown prometheus server, got %v", err)
	}

	ps.Logger.Infof("prometheus server shutting down")
}
