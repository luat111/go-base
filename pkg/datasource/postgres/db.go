package postgres

import (
	"go-base/pkg/common"
	"go-base/pkg/config"
	"go-base/pkg/logger"

	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
	logger *DBLogger
	config *DBConfig
}

func New(config config.Config) *DB {
	logger := logger.NewLogger(common.PGPrefix)
	dbLog := newDbLogger(logger)
	db, dbConfig := newPostgreSQLInstance(config, logger, dbLog)

	dbInstance := &DB{
		DB:     db,
		logger: dbLog,
		config: dbConfig,
	}

	go retryConnection(dbInstance)

	return dbInstance
}

func (d *DB) Close() error {
	instance, err := d.DB.DB()

	if err != nil {
		return err
	}

	return instance.Close()
}
