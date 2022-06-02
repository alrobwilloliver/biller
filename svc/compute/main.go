package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"biller/lib/ff"
	"biller/lib/logger"
	"biller/lib/postgresql"
	"biller/lib/service"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/cors"

	"biller/svc/compute/billingaccount"
	"biller/svc/compute/project"
	"biller/svc/compute/store"

	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()
	logger := logger.Get()

	if err := run(ctx, os.Args[1:], logger); err != nil {
		logger.Fatal("fatal error", zap.Error(err))
	}
}

func run(ctx context.Context, args []string, logger *zap.Logger) error {
	var (
		environment         string
		pgDatabase          string
		pgHost              string
		pgPass              string
		pgPort              int
		pgUser              string
		runner              string
		shutdownGracePeriod time.Duration
	)

	{
		fs := flag.NewFlagSet("compute", flag.ExitOnError)
		fs.StringVar(&environment, "environment", "local", "")
		fs.StringVar(&pgHost, "pg-host", "localhost", "the host to use when connecting to Postgresql")
		fs.StringVar(&pgDatabase, "pg-database", "compute", "the database to use when connected to Postgresql")
		fs.StringVar(&pgPass, "pg-pass", "", "the password to use when connecting to Postgresql")
		fs.IntVar(&pgPort, "pg-port", 5432, "the port to use when connecting to Postgresql")
		fs.StringVar(&pgUser, "pg-user", "compute", "the user to use when connecting to Postgresql")
		fs.DurationVar(&shutdownGracePeriod, "shutdown-grace-period", 0, "")
		// TODO we need tasks for, polling one vm/host state, update demander balances, update supplier earnings, supplier payments/trasnactions to kill bill
		fs.StringVar(&runner, "runner", "", `Choose which background task to run, either "onePoller" or "earningsRollup". Leave empty to run the market server itself.`)

		err := ff.Fill(fs, args)

		if err != nil {
			return fmt.Errorf("failed to get configuration: %w", err)
		}
	}

	registry := prometheus.NewRegistry()

	var postgresqlQueries *store.TxQueries
	{
		postgresqlClientConfig := &postgresql.ClientConfig{
			User:     pgUser,
			Pass:     pgPass,
			Host:     pgHost,
			Port:     pgPort,
			Database: pgDatabase,
			Logger:   logger,
		}

		postgresqlDb, err := postgresql.NewClient(ctx, postgresqlClientConfig, registry)
		if err != nil {
			logger.Error("error connecting to postgresql", zap.Error(err))
		} else {
			defer postgresqlDb.Close()
			postgresqlQueries = store.NewTxQueries(postgresqlDb)
		}
	}

	promServerConfig := &service.PrometheusServerConfig{
		ListenAddr: ":9090",
	}

	if runner != "" {
		var backgroundTaskConfig service.BackgroundServiceConfig

		switch runner {
		case "biller":
			backgroundTaskConfig = service.BackgroundServiceConfig{
				Environment:      environment,
				Interval:         time.Hour * 24,
				Name:             "biller",
				PrometheusServer: promServerConfig,
				Task:             billingaccount.NewBiller(postgresqlQueries, logger),
			}

		default:
			return fmt.Errorf("incorrect task name %q", runner)
		}

		backgroundTask := service.NewBackground(backgroundTaskConfig, registry, logger)
		backgroundTask.Run(ctx)

	} else {
		svc := service.New(
			service.ServiceConfig{
				Environment: environment,
				GRPCGateway: &service.GRPCGatewayConfig{
					Cors: &cors.Options{
						AllowedOrigins:     []string{"https://staging.compute.cudo.org"},
						AllowCredentials:   true,
						AllowedMethods:     []string{"DELETE", "GET", "PATCH", "POST", "PUT"},
						AllowedHeaders:     []string{"Content-Type", "X-Session-Token"},
						ExposedHeaders:     []string{},
						OptionsPassthrough: false,
						MaxAge:             0,
					},
					GRPCServerAddr: ":9000",
					ListenAddr:     ":8080",
				},
				GRPCServices: map[string]*service.GRPCServiceConfig{
					"compute": {
						ListenAddr: ":9000",
						Name:       "compute",
					},
				},
				PrometheusServer:    promServerConfig,
				ShutdownGracePeriod: shutdownGracePeriod,
			},
			registry,
			logger,
		)

		billingAccountServiceHandler := billingaccount.NewServer(postgresqlQueries, logger)
		billingaccount.RegisterBillingAccountServiceServer(svc.GRPCServices["compute"].GRPCServer, billingAccountServiceHandler)
		err := billingaccount.RegisterBillingAccountServiceHandler(ctx, svc.GRPCGateway.GatewayMux, svc.GRPCGateway.GRPCClientConn)
		if err != nil {
			return fmt.Errorf("failed to register grpc-gateway service account handler: %w", err)
		}

		projectServiceHandler := project.NewServer(postgresqlQueries, logger)
		project.RegisterProjectServiceServer(svc.GRPCServices["compute"].GRPCServer, projectServiceHandler)
		err = project.RegisterProjectServiceHandler(ctx, svc.GRPCGateway.GatewayMux, svc.GRPCGateway.GRPCClientConn)
		if err != nil {
			return fmt.Errorf("failed to register grpc-gateway service project handler: %w", err)
		}

		svc.Run(ctx)
	}

	return nil
}
