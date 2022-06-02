// Package service provides a framework for a typical microservice, providing health checking and monitoring
// whilst possible to write a service without it, it is not recommended as you are likely to diverge from
// organisation standards.
package service

import (
	"context"
	"errors"
	"os/signal"
	"syscall"
	"time"

	"biller/lib/version"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type Runner interface {
	Run(ctx context.Context) error
}

type BackgroundServiceConfig struct {
	Environment      string
	Interval         time.Duration
	Sleep            time.Duration
	Name             string
	PrometheusServer *PrometheusServerConfig
	Task             Runner
	Timeout          time.Duration
}

type BackgroundService struct {
	config           BackgroundServiceConfig
	log              *zap.Logger
	prometheusServer *PrometheusServer
	registry         *prometheus.Registry
}

// New returns a new instance of a background service that repeats a run then a sleep for the duration if the sleep parameter is used.
// If an interval is specified, it returns a new instance of a background service that runs everytime a ticker ticks.
func NewBackground(config BackgroundServiceConfig, registry *prometheus.Registry, log *zap.Logger) *BackgroundService {
	log.Info(
		"creating background task",
		zap.String("buildTime", version.BuildTime),
		zap.String("environment", config.Environment),
		zap.String("service", config.Name),
		zap.String("version", version.Version),
	)

	s := &BackgroundService{
		config:   config,
		log:      log,
		registry: registry,
	}

	if s.config.PrometheusServer != nil {
		s.prometheusServer = newPrometheusServer(*s.config.PrometheusServer, registry, log)
	}

	return s
}

func (s *BackgroundService) Run(ctx context.Context) error {
	// make a context that is cancelled when the process receives a SIGINT or SIGTERM
	signalCtx, stopSignalCtx := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)

	promServerCtx := context.Background()
	if s.prometheusServer != nil {
		promServerCtx = s.prometheusServer.start()
		defer s.prometheusServer.shutdown()
	}

	s.log.Info("starting background service", zap.String("name", s.config.Name))

	err := errors.New("background service must either have an interval or a sleep duration specified")

	if s.config.Sleep != 0 {
		if s.config.Timeout == 0 {
			s.config.Timeout = 10 * time.Minute
		}
		err = s.run(func() <-chan time.Time { return time.After(s.config.Sleep) }, signalCtx, promServerCtx)
	}
	if s.config.Interval != 0 {
		if s.config.Timeout == 0 {
			s.config.Timeout = s.config.Interval
		}
		ticker := time.NewTicker(s.config.Interval)
		defer ticker.Stop()
		err = s.run(func() <-chan time.Time { return ticker.C }, signalCtx, promServerCtx)
	}

	if err != nil {
		return err
	}

	s.log.Info("stopping background service", zap.String("name", s.config.Name))

	// reset signal handling so signals are no longer caught
	stopSignalCtx()
	return nil
}

// Runs the task at every interval period. If not complete in that time, it is
// cancelled and logged, and we re-run it.
func (s *BackgroundService) run(next func() <-chan time.Time, signalCtx context.Context, promServerCtx context.Context) error {
	for {
		ctx, cancel := context.WithTimeout(signalCtx, s.config.Timeout)
		s.log.Debug("running task", zap.String("name", s.config.Name))
		err := s.config.Task.Run(ctx)
		cancel()
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				s.log.Warn("task timed out")
			} else {
				s.log.Error("task error", zap.Error(err))
			}
		}

		select {
		// wait here until the timer expires, the parent context is done, or the prometheus server ends
		case <-next():
			// if the sleep time expires, we run the task again
		case <-signalCtx.Done():
			// if the parent context is finished, we exit
			return signalCtx.Err()
		case <-promServerCtx.Done():
			s.log.Error("prometheus http server exited early")
			return promServerCtx.Err()
		}
	}
}
