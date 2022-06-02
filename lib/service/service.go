// Package svc provides a framework for a typical microservice, providing health checking and monitoring
// whilst possible to write a service without svc, it is not recommended as you are likely to diverge from
// organisation standards.
package service

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"biller/lib/version"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

const defaultHTTPServerShutdownGracePeriod = 15 * time.Second

type ServiceConfig struct {
	Environment         string
	GRPCGateway         *GRPCGatewayConfig
	GRPCServices        map[string]*GRPCServiceConfig
	PrometheusServer    *PrometheusServerConfig
	ShutdownGracePeriod time.Duration
}

// Service allows setup for basic dependencies of most listener services e.g health checking
type Service struct {
	config           ServiceConfig
	GRPCGateway      *GRPCGateway
	GRPCServices     map[string]*GRPCService
	prometheusServer *PrometheusServer
	log              *zap.Logger
	registry         *prometheus.Registry
}

// New returns a new instance of service
func New(config ServiceConfig, registry *prometheus.Registry, log *zap.Logger) *Service {
	s := &Service{
		config:   config,
		log:      log,
		registry: registry,
	}

	if s.config.PrometheusServer != nil {
		s.prometheusServer = newPrometheusServer(*s.config.PrometheusServer, registry, log)
	}

	if len(s.config.GRPCServices) > 0 {
		s.GRPCServices = make(map[string]*GRPCService, len(s.config.GRPCServices))
		for svcName, svcConfig := range s.config.GRPCServices {
			s.GRPCServices[svcName] = newGRPCService(*svcConfig, registry, log)
		}
	}

	if s.config.GRPCGateway != nil {
		s.GRPCGateway = newGRPCGateway(*s.config.GRPCGateway, log)
	}

	return s
}

// Run starts the main service as well as several debugging/healthcheck services such as prometheus.
// It should only be called once and should be called once any desired services have been attached.
func (s *Service) Run(ctx context.Context) {
	s.log.Info(
		"running service",
		zap.String("buildTime", version.BuildTime),
		zap.String("environment", s.config.Environment),
		zap.String("version", version.Version),
	)
	// make a context that is cancelled when the process receives a SIGINT or SIGTERM
	signalCtx, stopSignalCtx := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)

	promServerCtx := context.Background()
	if s.prometheusServer != nil {
		promServerCtx = s.prometheusServer.start()
		defer s.prometheusServer.shutdown()
	}

	grpcSvcErrGroup, grpcSvcErrGroupCtx := errgroup.WithContext(ctx)
	if len(s.GRPCServices) > 0 {
		for _, grpcSvc := range s.GRPCServices {
			grpcSvcErrGroup.Go(grpcSvc.run)
			defer grpcSvc.shutdown()
		}
	}

	// make a context that never cancels as a fallback for when the gateway is not configured
	gatewayServerCtx := context.Background()
	if s.GRPCGateway != nil {
		var err error
		gatewayServerCtx, err = s.GRPCGateway.start()
		if err != nil {
			s.log.Error("could not start grpc-gateway", zap.Error(err))
			return
		}
		defer s.GRPCGateway.shutdown()
	}

	// if nothing has caused an error, mark as healthy
	if promServerCtx.Err() == nil &&
		grpcSvcErrGroupCtx.Err() == nil &&
		gatewayServerCtx.Err() == nil &&
		signalCtx.Err() == nil {
		for _, grpcSvc := range s.GRPCServices {
			grpcSvc.ready()
		}
	}

	// wait until any of the listeners stop or a signal is received
	select {
	case <-promServerCtx.Done():
		s.log.Info("prometheus server ended", zap.Error(promServerCtx.Err()))
	case <-grpcSvcErrGroupCtx.Done():
		s.log.Info("grpc server ended", zap.Error(grpcSvcErrGroupCtx.Err()))
	case <-gatewayServerCtx.Done():
		s.log.Info("grpc gateway server ended", zap.Error(gatewayServerCtx.Err()))
	case <-signalCtx.Done():
		s.log.Info("termination signal received")
		// Mark services unhealthy to loadbalancers.
		for _, grpcSvc := range s.GRPCServices {
			grpcSvc.notReady()
		}
		// reset signal handling so signals are no longer caught
		stopSignalCtx()
		if s.config.ShutdownGracePeriod > 0 {
			// Give load balancers time to shift traffic
			s.log.Warn("waiting for termination grace period", zap.Duration("duration", s.config.ShutdownGracePeriod))
			time.Sleep(s.config.ShutdownGracePeriod)
		}
	}
}
