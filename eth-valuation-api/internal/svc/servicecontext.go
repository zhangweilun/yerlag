package svc

import (
	"log"
	"time"

	"eth-valuation-api/internal/config"
	"eth-valuation-api/internal/fetcher"
	"eth-valuation-api/internal/model"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ServiceContext holds all dependencies for the service handlers.
type ServiceContext struct {
	Config      config.Config
	DB          *gorm.DB
	RedisDB     *redis.Client
	DataFetcher fetcher.DataFetcher
}

// NewServiceContext creates a new ServiceContext with the given config.
// It initializes the GORM database connection, runs auto-migration,
// sets up the Redis client, and creates the DataFetcher.
func NewServiceContext(c config.Config) *ServiceContext {
	db := mustInitDB(c.Database)
	rdb := initRedis(c.Redis)

	df := fetcher.NewDataFetcher(rdb, fetcher.DataFetcherConfig{
		MaxRetries:     3,
		RetryDelay:     1 * time.Second,
		RequestTimeout: 10 * time.Second,
	})

	return &ServiceContext{
		Config:      c,
		DB:          db,
		RedisDB:     rdb,
		DataFetcher: df,
	}
}

// mustInitDB opens a GORM database connection and runs auto-migration.
// It panics on failure because the service cannot operate without a database.
func mustInitDB(cfg config.DatabaseConfig) *gorm.DB {
	db, err := gorm.Open(mysql.Open(cfg.DataSource), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := model.AutoMigrate(db); err != nil {
		log.Fatalf("failed to auto-migrate database: %v", err)
	}

	return db
}

// initRedis creates a new Redis client from the given config.
func initRedis(cfg config.RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Host,
		Password: cfg.Pass,
		DB:       cfg.DB,
	})
}
