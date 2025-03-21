package cacheredis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/trinhdaiphuc/go-kit/log"
	"github.com/trinhdaiphuc/go-kit/metrics"
)

// Config configuration
type Config struct {
	Addresses       string        `json:"addresses,omitempty" mapstructure:"addresses"`
	Host            string        `json:"host,omitempty" mapstructure:"host"`
	Password        string        `json:"password,omitempty" mapstructure:"password"`
	Prefix          string        `json:"prefix,omitempty" mapstructure:"prefix"`
	ReadTimeout     time.Duration `json:"read_timeout,omitempty" mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout,omitempty" mapstructure:"write_timeout"`
	Port            int32         `json:"port,omitempty" mapstructure:"port"`
	DB              int           `json:"db,omitempty" mapstructure:"db"`
	PoolSize        int           `json:"pool_size,omitempty" mapstructure:"pool_size"`
	IdleConnections int           `json:"idle_connections,omitempty" mapstructure:"idle_connections"`
	Cluster         bool          `json:"cluster" mapstructure:"cluster"`
	EnableMonitor   bool          `json:"enable_monitor" mapstructure:"enable_monitor"`
	EnableTracing   bool          `json:"enable_tracing" mapstructure:"enable_tracing"`
}

func (c *Config) GetAddresses() []string {
	if len(c.Addresses) == 0 {
		return []string{fmt.Sprintf("%s:%d", c.Host, c.Port)}
	}
	return strings.Split(c.Addresses, ",")
}

func NewClient(cfg *Config) (redis.UniversalClient, func(), error) {
	client := redis.NewUniversalClient(
		&redis.UniversalOptions{
			Addrs:        cfg.GetAddresses(),
			Password:     cfg.Password,
			PoolSize:     cfg.PoolSize,
			MinIdleConns: cfg.IdleConnections,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
		},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := client.Ping(ctx).Err()
	if err != nil {
		log.Bg().Error("Ping failed", zap.Error(err))
		return nil, nil, err
	}

	if cfg.EnableTracing {
		if err := redisotel.InstrumentTracing(client, redisotel.WithDBStatement(false)); err != nil {
			return nil, nil, err
		}
	}

	if cfg.EnableMonitor {
		client.AddHook(metrics.NewRedisHook())
	}

	cleanup := func() {
		err := client.Close()
		if err != nil {
			log.Bg().Error("Close redis connection failed", zap.Error(err))
		}
	}
	return client, cleanup, nil
}

func NewClusterClient(cfg *Config) (redis.UniversalClient, func(), error) {
	client := redis.NewClusterClient(
		&redis.ClusterOptions{
			Addrs:        cfg.GetAddresses(),
			Password:     cfg.Password,
			PoolSize:     cfg.PoolSize,
			MinIdleConns: cfg.IdleConnections,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
		},
	)

	log.Bg().Info("Connecting to redis cluster", zap.Any("addresses", cfg.GetAddresses()))

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := client.Ping(ctx).Err()
	if err != nil {
		log.Bg().Error("Ping failed", zap.Error(err))
		return nil, nil, err
	}

	if cfg.EnableTracing {
		if err := redisotel.InstrumentTracing(client, redisotel.WithDBStatement(false)); err != nil {
			return nil, nil, err
		}
	}

	if cfg.EnableMonitor {
		client.AddHook(metrics.NewRedisHook())
	}

	cleanup := func() {
		err := client.Close()
		if err != nil {
			log.Bg().Error("Close redis connection failed", zap.Error(err))
		}
	}

	log.Bg().Info("Redis client connected")
	return client, cleanup, nil
}
