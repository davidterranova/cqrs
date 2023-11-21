package pg

import (
	"database/sql"
	"time"

	"github.com/davidterranova/cqrs/xstrings"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const postgresDriverName = "pgx"

type DBConfig struct {
	Name               string                `envconfig:"NAME" required:"true"`
	ConnString         xstrings.SecretString `envconfig:"CONN_STRING" required:"true"`
	MaxOpenConnections int                   `envconfig:"MAX_OPEN_CONNECTIONS" default:"25"`
	MaxIdleConnections int                   `envconfig:"MAX_IDLE_CONNECTIONS" default:"10"`
	ConnMaxIdleTime    time.Duration         `envconfig:"CONN_MAX_IDLE_TIME" default:"1m"`
	ConnMaxLifetime    time.Duration         `envconfig:"CONN_MAX_LIFETIME" default:"5m"`
}

// DSN "user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai"
func Open(cfg DBConfig) (*gorm.DB, error) {
	sqlDB, err := sql.Open(postgresDriverName, string(cfg.ConnString))
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConnections)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConnections)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return gorm.Open(
		postgres.New(postgres.Config{
			Conn: sqlDB,
		}),
		&gorm.Config{})
}
