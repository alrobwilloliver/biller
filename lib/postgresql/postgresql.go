package postgresql

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ClientConfig struct {
	User       string
	Pass       string
	Host       string
	Port       int
	Database   string
	ClientKey  string
	ClientCert string
	CA         string
	Logger     *zap.Logger
}

func (c *ClientConfig) SSLEnabled() bool {
	return c.ClientKey != "" && c.ClientCert != "" && c.CA != ""
}

func (c *ClientConfig) DSN() string {
	dsn := fmt.Sprintf(
		"user=%s password='%s' host=%s port=%d database=%s",
		c.User,
		c.Pass,
		c.Host,
		c.Port,
		c.Database,
	)

	if c.SSLEnabled() {
		dsn = fmt.Sprintf("%s sslmode=verify-ca", dsn)
	} else {
		dsn = fmt.Sprintf("%s sslmode=disable", dsn)
	}

	return dsn

	/*
		* dbname - The name of the database to connect to
		* user - The user to sign in as
		* password - The user's password
		* host - The host to connect to. Values that start with / are for unix
		domain sockets. (default is localhost)
		* port - The port to bind to. (default is 5432)
		* sslmode - Whether or not to use SSL (default is require, this is not
		the default for libpq)
		* fallback_application_name - An application_name to fall back to if one isn't provided.
		* connect_timeout - Maximum wait for connection, in seconds. Zero or
		not specified means wait indefinitely.
		* sslcert - Cert file location. The file must contain PEM encoded data.
		* sslkey - Key file location. The file must contain PEM encoded data.
		* sslrootcert - The location of the root certificate file. The file
		must contain PEM encoded data.
		Valid values for sslmode are:

		* disable - No SSL
		* require - Always SSL (skip verification)
		* verify-ca - Always SSL (verify that the certificate presented by the
		server was signed by a trusted CA)
		* verify-full - Always SSL (verify that the certification presented by
		the server was signed by a trusted CA and the server host name
		matches the one in the certificate)
		See http://www.postgresql.org/docs/current/static/libpq-connect.html#LIBPQ-CONNSTRING for more information about connection string parameters.
	*/
}

func NewClient(ctx context.Context, cfg *ClientConfig, registry prometheus.Registerer) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(cfg.DSN())
	if cfg.Logger != nil {
		config.ConnConfig.Logger = &wrappedLogger{logger: cfg.Logger.WithOptions(zap.AddCallerSkip(1))}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to parse pgx config: %w", err)
	}
	if cfg.SSLEnabled() {
		rootCertPool := x509.NewCertPool()
		if ok := rootCertPool.AppendCertsFromPEM([]byte(cfg.CA)); !ok {
			return nil, fmt.Errorf("failed to append PEM.")
		}
		clientCert, err := tls.X509KeyPair([]byte(cfg.ClientCert), []byte(cfg.ClientKey))
		if err != nil {
			return nil, fmt.Errorf("failed to produce x509 keypair: %w", err)
		}

		config.ConnConfig.TLSConfig = &tls.Config{
			RootCAs:            rootCertPool,
			Certificates:       []tls.Certificate{clientCert},
			InsecureSkipVerify: true,
		}
	}
	db, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgresql: %w", err)
	}

	err = instrumentDBStats(db, registry)
	if err != nil {
		return nil, err
	}

	return db, nil
}

type wrappedLogger struct {
	logger *zap.Logger
}

// Log copied from pgx / log / zapadapter but the info level is altered to output as debug.
func (pl *wrappedLogger) Log(ctx context.Context, level pgx.LogLevel, msg string, data map[string]interface{}) {
	fields := make([]zapcore.Field, len(data))
	i := 0
	for k, v := range data {
		fields[i] = zap.Any(k, v)
		i++
	}

	switch level {
	case pgx.LogLevelTrace:
		pl.logger.Debug(msg, append(fields, zap.Stringer("PGX_LOG_LEVEL", level))...)
	case pgx.LogLevelDebug:
		pl.logger.Debug(msg, fields...)
	case pgx.LogLevelInfo:
		pl.logger.Debug(msg, fields...)
	case pgx.LogLevelWarn:
		pl.logger.Warn(msg, fields...)
	case pgx.LogLevelError:
		pl.logger.Error(msg, fields...)
	default:
		pl.logger.Error(msg, append(fields, zap.Stringer("PGX_LOG_LEVEL", level))...)
	}
}

func instrumentDBStats(pool *pgxpool.Pool, registry prometheus.Registerer) error {
	var err error

	err = registry.Register(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "go_postgresql_acquired_connections",
		Help: "The number of currently acquired connections in the pool",
	}, func() float64 {
		return float64(pool.Stat().AcquiredConns())
	}))
	if err != nil {
		return err
	}

	err = registry.Register(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "go_postgresql_constructing_connections",
		Help: "The number of connections with construction in progress in the pool.",
	}, func() float64 {
		return float64(pool.Stat().ConstructingConns())
	}))
	if err != nil {
		return err
	}

	err = registry.Register(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "go_postgresql_idle_connections",
		Help: "The number of currently idle connections in the pool.",
	}, func() float64 {
		return float64(pool.Stat().IdleConns())
	}))
	if err != nil {
		return err
	}

	err = registry.Register(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "go_postgresql_max_connections",
		Help: "The maximum size of the pool.",
	}, func() float64 {
		return float64(pool.Stat().MaxConns())
	}))
	if err != nil {
		return err
	}

	err = registry.Register(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "go_postgresql_total_connections",
		Help: "The total number of resources currently in the pool. The sum of connections being constructed, acquired and idle.",
	}, func() float64 {
		return float64(pool.Stat().TotalConns())
	}))
	if err != nil {
		return err
	}

	return nil
}
