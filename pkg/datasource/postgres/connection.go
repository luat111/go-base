package postgres

import (
	"fmt"
	"go-base/pkg/config"
	"go-base/pkg/logger"
	"strconv"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	maxDBIdleConns = 10
	maxDBOpenConns = 100
	defaultDBPort  = 5432
	// maxConnLifeTime = 30 * time.Minute
)

type DBConfig struct {
	HostName    string
	User        string
	Password    string
	Port        string
	Database    string
	SSLMode     string
	MaxIdleConn int
	MaxOpenConn int
	Charset     string
}

func newPostgreSQLInstance(config config.Config, logger logger.ILogger, dbLogger *DBLogger) (*gorm.DB, *DBConfig) {
	dbConfig := getDBConfig(config)
	dbConnectionString := getConnectionString(dbConfig)

	db, err := gorm.Open(postgres.Open(dbConnectionString), &gorm.Config{
		Logger: dbLogger,
	})

	if err != nil {
		logger.Error("Failed to connect to the Postgres", "err", err.Error())
	} else {
		logger.Info("Connected to Postgres")
	}

	setupExtension(db)

	db = db.Set("gorm:auto_preload", true)

	return db, dbConfig
}

func getConnectionString(dbConfig *DBConfig) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbConfig.HostName, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.Database, dbConfig.SSLMode)
}

func getDBConfig(configs config.Config) *DBConfig {
	return &DBConfig{
		HostName:    configs.Get("PG_HOST"),
		User:        configs.Get("PG_USER"),
		Password:    configs.Get("PG_PWD"),
		Port:        configs.GetOrDefault("PG_PORT", strconv.Itoa(defaultDBPort)),
		Database:    configs.Get("PG_DB"),
		MaxOpenConn: maxDBOpenConns,
		MaxIdleConn: maxDBIdleConns,
		SSLMode:     configs.GetOrDefault("DB_SSL_MODE", "disable"),
		Charset:     configs.Get("DB_CHARSET"),
	}
}

func retryConnection(database *DB) {
	const connRetryFrequencyInSeconds = 10
	db, _ := database.DB.DB()
	for {
		if db.Ping() != nil {
			database.logger.baseLogger.Warn("Connection to postgres database lost")

			for {
				err := db.Ping()
				if err == nil {
					database.logger.baseLogger.Info("Reconnected to postgres")

					break
				}

				database.logger.baseLogger.Warn("Retrying connect to postgres")

				time.Sleep(connRetryFrequencyInSeconds * time.Second)
			}
		}

		time.Sleep(connRetryFrequencyInSeconds * time.Second)
	}
}
