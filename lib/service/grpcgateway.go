// Package svc provides a framework for a typical microservice, providing health checking and monitoring
// whilst possible to write a service without svc, it is not recommended as you are likely to diverge from
// organisation standards.
package service

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

const defaultGatewayListenAddr = ":8080"
const defaultHandlerTimeout = 30 * time.Second

type GRPCGatewayConfig struct {
	ClientCA               string
	Cors                   *cors.Options
	ListenAddr             string
	GRPCServerAddr         string
	GRPCServerDialInsecure bool
	HandlerTimeout         time.Duration
	ShutdownGracePeriod    time.Duration
	ServeMuxOpts           []runtime.ServeMuxOption
	TLSCertFile            string
	TLSKeyFile             string
}

// Service allows setup for basic dependencies of most listener services e.g health checking
type GRPCGateway struct {
	config         GRPCGatewayConfig
	GRPCClientConn *grpc.ClientConn
	GatewayMux     *runtime.ServeMux
	handler        http.Handler
	log            *zap.Logger
	server         *http.Server
}

// newGRPCGateway returns a new instance of service
func newGRPCGateway(config GRPCGatewayConfig, log *zap.Logger) *GRPCGateway {
	if config.ListenAddr == "" {
		config.ListenAddr = defaultGatewayListenAddr
	}

	if config.ShutdownGracePeriod == 0 {
		config.ShutdownGracePeriod = defaultHTTPServerShutdownGracePeriod
	}

	if config.HandlerTimeout == 0 {
		config.HandlerTimeout = defaultHandlerTimeout
	}

	if config.Cors == nil {
		config.Cors = &cors.Options{
			AllowedOrigins: []string{"*"},
		}
	}

	dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	if config.GRPCServerDialInsecure {
		creds := credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true,
		})
		dialOpts = []grpc.DialOption{grpc.WithTransportCredentials(creds)}
	}
	grpcClientConn, err := grpc.Dial(config.GRPCServerAddr, dialOpts...)
	if err != nil {
		if grpcClientConn != nil {
			if cerr := grpcClientConn.Close(); cerr != nil {
				log.Sugar().Infof("failed to close conn to %s: %v", config.GRPCServerAddr, cerr)
			}
		}
		log.Sugar().Errorf("failed to dial grpc server %s: %v", config.GRPCServerAddr, err)
	}

	healthCheck := grpc_health_v1.NewHealthClient(grpcClientConn)
	gwmux := runtime.NewServeMux(
		append(
			config.ServeMuxOpts,
			runtime.WithHealthzEndpoint(healthCheck),
		)...,
	)

	gwmuxWithCors := cors.New(*config.Cors).Handler(gwmux)

	return &GRPCGateway{
		config:         config,
		GatewayMux:     gwmux,
		GRPCClientConn: grpcClientConn,
		log:            log,
		handler:        http.TimeoutHandler(gwmuxWithCors, config.HandlerTimeout, ""),
	}

}

func (s *GRPCGateway) start() (context.Context, error) {
	s.log.Debug("starting grpc-gateway listener", zap.String("addr", s.config.ListenAddr))
	errorLog, err := zap.NewStdLogAt(s.log, zap.ErrorLevel)
	if err != nil {
		s.log.Error("could create error logger for grpc-gateway server", zap.Error(err))
	}

	s.server = &http.Server{
		ErrorLog:          errorLog,
		Addr:              s.config.ListenAddr,
		Handler:           s.handler,
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	if s.config.TLSCertFile != "" && s.config.TLSKeyFile != "" {
		s.log.Debug("enabling tls for grpc-gateway service")
		cer, err := tls.LoadX509KeyPair(s.config.TLSCertFile, s.config.TLSKeyFile)
		if err != nil {
			return nil, err
		}
		tlsConfig := tls.Config{
			Certificates: []tls.Certificate{cer},
			MinVersion:   tls.VersionTLS12,
		}
		if s.config.ClientCA != "" {
			s.log.Info("enabling tls client certificate CA check for grpc-gateway service")
			pool := x509.NewCertPool()
			if !pool.AppendCertsFromPEM([]byte(s.config.ClientCA)) {
				return nil, fmt.Errorf("could not append client CA cert to pool for client tls verification")
			}
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
			tlsConfig.ClientCAs = pool
		}
		s.server.TLSConfig = &tlsConfig
	}
	return startHTTPServer(s.server, s.log), nil
}

func (s *GRPCGateway) shutdown() {
	s.log.Info("shutting down grpc-gateway listener", zap.String("addr", s.config.ListenAddr))
	shutdownHTTPServer(s.server, s.config.ShutdownGracePeriod, s.log)
	if s.GRPCClientConn != nil {
		if err := s.GRPCClientConn.Close(); err != nil {
			s.log.Sugar().Warnf("failed to close conn to %s: %v", s.config.GRPCServerAddr, err)
		}
	}
}
