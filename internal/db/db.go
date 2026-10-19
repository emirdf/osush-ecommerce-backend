package db

import (
	"context"
	"fmt"
	"osush/internal/config"
	"osush/internal/log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v4/pgxpool"
)

func (ldb *LoggedDB) GetDB() *pgxpool.Pool {
	return ldb.DB
}

func (ldb *LoggedDB) Close() {
	if ldb.DB != nil {
		ldb.DB.Close()
	}
}

func NewDB(cfg *config.Config, logger log.Logger) (*LoggedDB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)

	databaseURL := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	logger.Info("connecting to database",
		log.String("host", cfg.DBHost),
		log.String("port", cfg.DBPort),
		log.String("dbname", cfg.DBName),
	)

	var pool *pgxpool.Pool
	var err error

	maxRetries := 5

	for i := 1; i <= maxRetries; i++ {
		poolConfig, parseErr := pgxpool.ParseConfig(dsn)

		if parseErr != nil {
			err = parseErr
		} else {
			poolConfig.MaxConns = 25
			poolConfig.MinConns = 5
			poolConfig.MaxConnLifetime = time.Hour
			poolConfig.MaxConnIdleTime = 20 * time.Minute

			pool, err = pgxpool.ConnectConfig(
				context.Background(),
				poolConfig,
			)
		}

		if err == nil {
			err = pool.Ping(context.Background())
		}

		if err == nil {
			logger.Info("Successfully connected to PostgreSQL")
			break
		}

		if pool != nil {
			pool.Close()
			pool = nil
		}

		if i == maxRetries {
			logger.Error(
				"failed to connect to database after max retries",
				log.Error(err),
			)

			return nil, fmt.Errorf(
				"failed to connect after %d attempts: %w",
				maxRetries,
				err,
			)
		}

		wait := time.Duration(i) * time.Second

		logger.Warn(
			"Database connection failed, retrying...",
			log.Int("attempt", i),
			log.Duration("waiting", wait),
		)

		time.Sleep(wait)
	}

	if err := RunMigrations("migrations", databaseURL); err != nil {
		logger.Error(
			"failed to run migrations",
			log.Error(err),
		)

		pool.Close()

		return nil, fmt.Errorf(
			"failed to run migrations: %w",
			err,
		)
	}

	return NewLoggedDB(pool, logger, cfg.Env), nil
}

func RunMigrations(migrationsPath, databaseURL string) error {

	m, err := migrate.New(
		"file://"+migrationsPath,
		databaseURL,
	)
	if err != nil {
		//log.Printf("error create migration: %v\n", err)
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		//log.Printf("error migration: %v\n", err)
		return err
	}

	//log.Println("successfull does migration")
	return nil
}

// func RunMigrations(cfg *config.Config, logger log.Logger) error {

// 	logger.Info("starting database migrations")

// 	dsn := fmt.Sprintf(
// 		"postgresql://%s:%s@%s:%s/%s?sslmode=disable",
// 		cfg.DBUser,
// 		cfg.DBPassword,
// 		cfg.DBHost,
// 		cfg.DBPort,
// 		cfg.DBName,
// 	)

// 	m, err := migrate.New(
// 		"file://migrations",
// 		dsn,
// 	)

// 	if err != nil {

// 		logger.Error(
// 			"failed to initialize migrations",
// 			log.Error(err),
// 		)

// 		return fmt.Errorf(
// 			"failed to initialize migration: %w",
// 			err,
// 		)
// 	}

// 	defer func() {
// 		sourceErr, dbErr := m.Close()

// 		if sourceErr != nil {
// 			logger.Warn(
// 				"failed to close migration source",
// 				log.Error(sourceErr),
// 			)
// 		}

// 		if dbErr != nil {
// 			logger.Warn(
// 				"failed to close migration database",
// 				log.Error(dbErr),
// 			)
// 		}
// 	}()

// 	err = m.Up()

// 	if err != nil && err != migrate.ErrNoChange {

// 		logger.Error(
// 			"database migrations failed",
// 			log.Error(err),
// 		)

// 		return fmt.Errorf(
// 			"migrations failed: %w",
// 			err,
// 		)
// 	}

// 	if err == migrate.ErrNoChange {

// 		logger.Info("no new database migrations")

// 		return nil
// 	}

// 	logger.Info("database migrations applied successfully")

// 	return nil
// }
