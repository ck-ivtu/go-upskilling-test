package repository

import (
	"fmt"
	"go-test/internal/model"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	Connection *gorm.DB
	Logger     *zap.Logger
}

type ConnectionParams struct {
	Host     string
	Port     string
	User     string
	Password string
	DbName   string
	SSLMode  string
	Logger   *zap.Logger
}

func NewRepository(params *ConnectionParams) (*Repository, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		params.Host, params.Port, params.User, params.Password, params.DbName, params.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		params.Logger.Error("Unable to connect to the database", zap.Error(err))
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return &Repository{Connection: db, Logger: params.Logger}, nil
}

func (r *Repository) Migrate() error {
	models := []interface{}{
		&model.DeliveryPackage{},
	}

	for _, migrationModel := range models {
		if err := r.Connection.AutoMigrate(migrationModel); err != nil {
			r.Logger.Error("Failed to migrate model", zap.String("model", fmt.Sprintf("%T", migrationModel)), zap.Error(err))
			return fmt.Errorf("failed to migrate model %v: %w", migrationModel, err)
		}
	}

	r.Logger.Info("All migrations ran successfully")
	return nil
}

func (r *Repository) CloseConn() error {
	connection, err := r.Connection.DB()
	if err != nil {
		r.Logger.Error("Failed to retrieve the raw database connection", zap.Error(err))
		return fmt.Errorf("unable to retrieve raw DB connection: %w", err)
	}

	if err := connection.Close(); err != nil {
		r.Logger.Error("Failed to close database connection", zap.Error(err))
		return fmt.Errorf("unable to close database connection: %w", err)
	}

	r.Logger.Info("Database connection closed successfully")
	return nil
}
