// Package svc provides a framework for a typical microservice, providing health checking and monitoring
// whilst possible to write a service without svc, it is not recommended as you are likely to diverge from
// organisation standards.
package service

import (
	"context"
	"net"
	"time"

	"biller/lib/logger/grpctrace"
	"biller/lib/service/grpcversion"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

const defaultGRPCListenAddr = ":9000"
const defaultConnectionTimeout = 30 * time.Second
const defaultRPCTimeout = 30 * time.Second

type GRPCServiceConfig struct {
	ConnectionTimeout time.Duration
	ListenAddr        string
	Name              string
	RPCTimeout        time.Duration
	SentryDSN         string
	TLSCertFile       string
	TLSKeyFile        string
}

type GRPCService struct {
	config     GRPCServiceConfig
	GRPCServer *grpc.Server
	health     *health.Server
	log        *zap.Logger
	registry   *prometheus.Registry
}

func newGRPCService(config GRPCServiceConfig, registry *prometheus.Registry, log *zap.Logger) *GRPCService {
	if config.ConnectionTimeout == 0 {
		config.ConnectionTimeout = defaultConnectionTimeout
	}
	if config.RPCTimeout == 0 {
		config.RPCTimeout = defaultRPCTimeout
	}

	s := &GRPCService{
		config:   config,
		log:      log,
		registry: registry,
	}

	var unaryInterceptors []grpc.UnaryServerInterceptor
	var streamInterceptors []grpc.StreamServerInterceptor

	var grpcMetrics *grpc_prometheus.ServerMetrics
	if s.registry != nil {
		svcNameLabel := prometheus.Labels{
			"grpc_service_name": s.config.Name,
		}
		// prometheus metrics for grpc calls
		grpcMetrics = grpc_prometheus.NewServerMetrics(grpc_prometheus.WithConstLabels(svcNameLabel))
		grpcMetrics.EnableHandlingTimeHistogram(grpc_prometheus.WithHistogramConstLabels(svcNameLabel))
		unaryInterceptors = append(unaryInterceptors, grpcMetrics.UnaryServerInterceptor())
		streamInterceptors = append(streamInterceptors, grpcMetrics.StreamServerInterceptor())
	}

	// debug log the name of the grpc method and the call latency
	unaryInterceptors = append(unaryInterceptors, grpctrace.UnaryInterceptor(log))

	// add a per rpc timeout
	unaryInterceptors = append(unaryInterceptors, TimeoutUnaryInterceptor(config.RPCTimeout))

	serverOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
		grpc.ConnectionTimeout(s.config.ConnectionTimeout),
	}

	if s.config.TLSCertFile != "" && s.config.TLSKeyFile != "" {
		log.Debug("enabling server tls for grpc server")
		var err error
		serverTLS, err := credentials.NewServerTLSFromFile(s.config.TLSCertFile, s.config.TLSKeyFile)
		if err != nil {
			log.Warn("could not enable server tls from provided cert and key files", zap.Error(err))
		}
		serverOpts = append(serverOpts, grpc.Creds(serverTLS))
	}

	s.GRPCServer = grpc.NewServer(serverOpts...)

	if grpcMetrics != nil {
		grpcMetrics.InitializeMetrics(s.GRPCServer)
		// register gRPC metrics on our custom registry
		if err := registry.Register(grpcMetrics); err != nil {
			log.Error("could not register gRPC server metrics", zap.Error(err))
		}
	}
	s.health = health.NewServer()
	// the default is set to SERVING, but we are just starting up so we set it to NOT_SERVING
	s.health.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)

	if config.ListenAddr == "" {
		s.config.ListenAddr = defaultGRPCListenAddr
	}

	// Register meta services used for debugging/healthchecking
	grpc_health_v1.RegisterHealthServer(s.GRPCServer, s.health)
	reflection.Register(s.GRPCServer)
	grpcversion.Register(s.GRPCServer)

	return s
}

// Run starts the main service as well as several debugging/healthcheck services such as prometheus.
// It should only be called once and should be called once any desired services have been attached.
// If the provided context expires
func (s *GRPCService) run() error {
	s.log.Info("starting grpc listener", zap.String("addr", s.config.ListenAddr), zap.String("service", s.config.Name))
	return runGRPCServer(s.GRPCServer, s.config.ListenAddr, s.log)
}

func runGRPCServer(srv *grpc.Server, listenAddr string, log *zap.Logger) error {
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Error(err.Error())
	} else {
		if err := srv.Serve(lis); err != nil {
			log.Error("grpc server closed with error", zap.Error(err), zap.Any("server", srv))
		} else {
			log.Debug("grpc server closed", zap.String("addr", listenAddr))
		}
	}
	return err
}

// Mark service unhealthy.
func (s *GRPCService) notReady() {
	s.health.Shutdown()
	s.log.Info("service not ready", zap.String("name", s.config.Name))
}

// Mark service healthy.
func (s *GRPCService) ready() {
	s.health.Resume()
}

func (s *GRPCService) shutdown() {
	s.log.Info("shutting down grpc listener", zap.String("addr", s.config.ListenAddr))
	s.GRPCServer.GracefulStop()
}

func TimeoutUnaryInterceptor(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return handler(timeoutCtx, req)
	}
}
