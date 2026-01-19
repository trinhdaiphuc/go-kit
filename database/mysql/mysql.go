package mysql

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/prometheus"

	l "github.com/trinhdaiphuc/go-kit/log"
)

// Config is settings of a Config server. It contains almost same fields as mysql.Config
// but with some different field names and tags.
type Config struct {
	Host                  string        `json:"host" mapstructure:"host"`
	Port                  int           `json:"port" mapstructure:"port"`
	Database              string        `json:"database" mapstructure:"database"`
	Username              string        `json:"username" mapstructure:"username"`
	Password              string        `json:"password" mapstructure:"password"`
	Timezone              string        `json:"timezone" mapstructure:"timezone"`
	MaxIdleConnections    int           `json:"max_idle_connections" mapstructure:"max_idle_connections"`
	MaxOpenConnections    int           `json:"max_open_connections" mapstructure:"max_open_connections"`
	MaxConnectionLifeTime time.Duration `json:"max_connection_life_time" mapstructure:"max_connection_life_time"`
}

func (m *Config) GetTimezone() string {
	location, err := time.LoadLocation(m.Timezone)
	if err != nil {
		return "UTC"
	}
	return location.String()
}

// FormatDSN returns Config DSN from settings.
func (m *Config) FormatDSN() string {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=%s", m.Username, m.Password, m.Host, m.Port, m.Database, m.GetTimezone())
	return dsn
}

type Clause func(tx *gorm.DB)

func ConnectMySQL(cfg *Config, serviceName string) (*gorm.DB, func(), error) {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,  // Slow SQL threshold
			LogLevel:                  logger.Error, // Log level
			IgnoreRecordNotFoundError: true,         // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      true,         // Don't include params in the SQL log
			Colorful:                  false,        // Disable color
		},
	)
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       cfg.FormatDSN(), // data source name
		DefaultStringSize:         256,             // default size for string fields
		DisableDatetimePrecision:  true,            // disable datetime precision, which not supported before MySQL 5.6
		DontSupportRenameIndex:    true,            // drop & create when rename index, rename index not supported before MySQL 5.7, MariaDB
		DontSupportRenameColumn:   true,            // `change` when rename column, rename column not supported before MySQL 8, MariaDB
		SkipInitializeWithVersion: false,           // auto configure based on currently MySQL version
	}), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		l.Bg().Error("Error open mySQL", zap.Error(err))
		return nil, nil, err
	}

	// Ping the database
	if err := CheckConnectMySQL(db); err != nil {
		l.Bg().Error("Error querying SELECT 1", zap.Error(err))
		return nil, nil, err
	}
	l.Bg().Info("Connected to database")

	if err := db.Use(otelgorm.NewPlugin()); err != nil {
		l.Bg().Error("Set up tracing plugin failed", zap.Error(err))
		return nil, nil, err
	}

	err = db.Use(prometheus.New(prometheus.Config{
		DBName:          cfg.Database, // `DBName` as metrics label
		RefreshInterval: 15,           // refresh metrics interval (default 15 seconds)
		StartServer:     false,        // start http server to expose metrics
		MetricsCollector: []prometheus.MetricsCollector{
			&prometheus.MySQL{VariableNames: []string{"Threads_running"}},
		},
		Labels: map[string]string{
			"service_name": serviceName, // config custom labels if necessary
		},
	}))
	if err != nil {
		l.Bg().Error("Set up prometheus plugin failed", zap.Error(err))
		return nil, nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		l.Bg().Error("Error get sql db", zap.Error(err))
		return nil, nil, err
	}

	cleanup := func() {
		if err := sqlDB.Close(); err != nil {
			l.Bg().Error("Error close sql db", zap.Error(err))
			return
		}
		l.Bg().Info("Closed sql db")
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConnections)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConnections)
	sqlDB.SetConnMaxLifetime(cfg.MaxConnectionLifeTime)
	return db, cleanup, nil
}

func CheckConnectMySQL(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("error get sql db: %w", err)
	}
	return sqlDB.Ping()
}
