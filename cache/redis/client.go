package cacheredis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.uber.org/zap"

	"github.com/trinhdaiphuc/go-kit/log"
)

// Default configuration values
const (
	DefaultPort            int32         = 6379
	DefaultPoolSize        int           = 10
	DefaultIdleConnections int           = 5
	DefaultReadTimeout     time.Duration = 3 * time.Second
	DefaultWriteTimeout    time.Duration = 3 * time.Second
	DefaultConnMaxIdleTime time.Duration = 30 * time.Minute
)

// Config configuration
type Config struct {
	Addresses       string        `json:"addresses,omitempty" mapstructure:"addresses"`
	Host            string        `json:"host,omitempty" mapstructure:"host"`
	Password        string        `json:"password,omitempty" mapstructure:"password"`
	Prefix          string        `json:"prefix,omitempty" mapstructure:"prefix"`
	MasterName      string        `json:"master_name,omitempty" mapstructure:"master_name"` // For Sentinel mode
	ReadTimeout     time.Duration `json:"read_timeout,omitempty" mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout,omitempty" mapstructure:"write_timeout"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time,omitempty" mapstructure:"conn_max_idle_time"` // Close connections after remaining idle for this duration
	Port            int32         `json:"port,omitempty" mapstructure:"port"`
	DB              int           `json:"db,omitempty" mapstructure:"db"`
	PoolSize        int           `json:"pool_size,omitempty" mapstructure:"pool_size"`
	IdleConnections int           `json:"idle_connections,omitempty" mapstructure:"idle_connections"`
	Cluster         bool          `json:"cluster" mapstructure:"cluster"`
	EnableMonitor   bool          `json:"enable_monitor" mapstructure:"enable_monitor"`
	EnableTracing   bool          `json:"enable_tracing" mapstructure:"enable_tracing"`
}

// SetDefaults applies default values for unset configuration fields
func (c *Config) SetDefaults() {
	if c.Port == 0 {
		c.Port = DefaultPort
	}
	if c.PoolSize == 0 {
		c.PoolSize = DefaultPoolSize
	}
	if c.IdleConnections == 0 {
		c.IdleConnections = DefaultIdleConnections
	}
	if c.ReadTimeout == 0 {
		c.ReadTimeout = DefaultReadTimeout
	}
	if c.WriteTimeout == 0 {
		c.WriteTimeout = DefaultWriteTimeout
	}
	if c.ConnMaxIdleTime == 0 {
		c.ConnMaxIdleTime = DefaultConnMaxIdleTime
	}
}

func (c *Config) GetAddresses() []string {
	if len(c.Addresses) == 0 {
		return []string{fmt.Sprintf("%s:%d", c.Host, c.Port)}
	}
	return strings.Split(c.Addresses, ",")
}

func NewClient(cfg *Config) (redis.UniversalClient, func(), error) {
	cfg.SetDefaults()

	// Enable Cluster mode for Redis platform
	if cfg.Cluster {
		return NewClusterClient(cfg)
	}

	client := redis.NewUniversalClient(
		&redis.UniversalOptions{
			MasterName:      cfg.MasterName,
			Addrs:           cfg.GetAddresses(),
			Password:        cfg.Password,
			PoolSize:        cfg.PoolSize,
			MinIdleConns:    cfg.IdleConnections,
			ConnMaxIdleTime: cfg.ConnMaxIdleTime,
			ReadTimeout:     cfg.ReadTimeout,
			WriteTimeout:    cfg.WriteTimeout,
			Protocol:        3,
		},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := client.Ping(ctx).Err()
	if err != nil {
		log.Bg().Error("Ping failed", log.Error(err))
		return nil, nil, err
	}

	if cfg.EnableTracing {
		if err := redisotel.InstrumentTracing(client, redisotel.WithDBStatement(false)); err != nil {
			return nil, nil, err
		}
	}

	if cfg.EnableMonitor {
		exp, err := prometheus.New()
		if err != nil {
			return nil, nil, err
		}
		metricProvider := metric.NewMeterProvider(metric.WithReader(exp))
		if err := redisotel.InstrumentMetrics(client, redisotel.WithMeterProvider(metricProvider)); err != nil {
			return nil, nil, err
		}
	}

	cleanup := func() {
		err := client.Close()
		if err != nil {
			log.Bg().Error("Close redis connection failed", log.Error(err))
		}
	}
	return client, cleanup, nil
}

func NewClusterClient(cfg *Config) (redis.UniversalClient, func(), error) {
	cfg.SetDefaults()

	client := redis.NewClusterClient(
		&redis.ClusterOptions{
			Addrs:           cfg.GetAddresses(),
			Password:        cfg.Password,
			PoolSize:        cfg.PoolSize,
			MinIdleConns:    cfg.IdleConnections,
			ConnMaxIdleTime: cfg.ConnMaxIdleTime,
			ReadTimeout:     cfg.ReadTimeout,
			WriteTimeout:    cfg.WriteTimeout,
			Protocol:        3,
		},
	)

	log.Bg().Info("Connecting to redis cluster", zap.Any("addresses", cfg.GetAddresses()))

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := client.Ping(ctx).Err()
	if err != nil {
		log.Bg().Error("Ping failed", log.Error(err))
		return nil, nil, err
	}

	if cfg.EnableTracing {
		if err := redisotel.InstrumentTracing(client, redisotel.WithDBStatement(false)); err != nil {
			return nil, nil, err
		}
	}

	if cfg.EnableMonitor {
		exp, err := prometheus.New()
		if err != nil {
			return nil, nil, err
		}
		metricProvider := metric.NewMeterProvider(metric.WithReader(exp))
		if err := redisotel.InstrumentMetrics(client, redisotel.WithMeterProvider(metricProvider)); err != nil {
			return nil, nil, err
		}
	}

	cleanup := func() {
		err := client.Close()
		if err != nil {
			log.Bg().Error("Close redis connection failed", log.Error(err))
		}
	}

	log.Bg().Info("Redis client connected")
	return client, cleanup, nil
}
