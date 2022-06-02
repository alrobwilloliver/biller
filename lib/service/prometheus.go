// Package svc provides a framework for a typical microservice, providing health checking and monitoring
// whilst possible to write a service without svc, it is not recommended as you are likely to diverge from
// organisation standards.
package service

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

const defaultPrometheusListenAddr = ":9090"

type PrometheusServerConfig struct {
	ListenAddr          string
	ShutdownGracePeriod time.Duration
}

type PrometheusServer struct {
	config   PrometheusServerConfig
	log      *zap.Logger
	server   *http.Server
	registry *prometheus.Registry
}

func newPrometheusServer(config PrometheusServerConfig, registry *prometheus.Registry, log *zap.Logger) *PrometheusServer {
	if config.ListenAddr == "" {
		config.ListenAddr = defaultPrometheusListenAddr
	}

	if config.ShutdownGracePeriod == 0 {
		config.ShutdownGracePeriod = defaultHTTPServerShutdownGracePeriod
	}

	s := &PrometheusServer{
		config:   config,
		log:      log,
		registry: registry,
	}

	return s
}

func (s *PrometheusServer) start() context.Context {
	s.log.Debug("starting prometheus server", zap.String("addr", s.config.ListenAddr))
	errorLog, err := zap.NewStdLogAt(s.log, zap.ErrorLevel)
	if err != nil {
		s.log.Warn("could not attach error logger to prometheus server")
	}
	s.server = &http.Server{
		Addr:    s.config.ListenAddr,
		Handler: promhttp.HandlerFor(s.registry, promhttp.HandlerOpts{ErrorLog: errorLog}),
	}

	ctx := startHTTPServer(s.server, s.log)

	return ctx
}

func (s *PrometheusServer) shutdown() {
	shutdownHTTPServer(s.server, s.config.ShutdownGracePeriod, s.log)
}
