package db

import (
	"context"
	"fmt"

	"github.com/spf13/viper"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const connectionString string = "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Europe/Rome"

func NewDBInstance(ctx context.Context, sugaredLogger *otelzap.SugaredLogger) (*gorm.DB, error) {
	connStr := fmt.Sprintf(connectionString, viper.GetString("postgres_host"), viper.GetString("postgres_port"),
		viper.GetString("postgres_user"), viper.GetString("postgres_password"), viper.GetString("postgres_database"),
		viper.GetString("postgres_ssl"))

	var gormConfig *gorm.Config
	if !sugaredLogger.Desugar().Core().Enabled(zap.DebugLevel) {
		gormConfig = &gorm.Config{}
	} else {
		gormConfig = &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}
	}

	db, err := gorm.Open(postgres.Open(connStr), gormConfig)
	if err != nil {
		return nil, err
	}

	if err = db.Use(otelgorm.NewPlugin()); err != nil {
		return nil, err
	}

	sugaredLogger.Ctx(ctx).Infow("connected to database", "database", viper.GetString("postgres_database"))

	return db, nil
}
