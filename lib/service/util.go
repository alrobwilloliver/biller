package service

import (
	"context"
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"
)

func startHTTPServer(srv *http.Server, log *zap.Logger) context.Context {
	// make a context that is done when the server ends
	listenCtx, cancel := context.WithCancel(context.Background())
	go func() {
		var err error
		if srv.TLSConfig != nil {
			// TLS config has already been provided, so we pass empty strings
			err = srv.ListenAndServeTLS("", "")
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				log.Debug("http server closed", zap.String("address", srv.Addr))
			} else {
				log.Error("http server closed with error", zap.Error(err), zap.String("address", srv.Addr))
			}
		}
		cancel()
	}()
	return listenCtx
}

func shutdownHTTPServer(srv *http.Server, timeout time.Duration, log *zap.Logger) {
	// don't pass a ctx in as we want a separate timeout that won't be cancelled
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	log.Debug("stopping http server", zap.String("address", srv.Addr))
	err := srv.Shutdown(ctx)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("error during graceful shutdown of http server", zap.Error(err), zap.String("addr", srv.Addr))
	}
}
