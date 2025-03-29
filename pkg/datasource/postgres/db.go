package postgres

import (
	"go-base/pkg/config"
	"go-base/pkg/logger"

	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
	logger *DBLogger
	config *DBConfig
}

func New(config config.Config, sysLog logger.ILogger) *DB {
	dbLog := newDbLogger(sysLog)
	db, dbConfig := newPostgreSQLInstance(config, sysLog, dbLog)

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
