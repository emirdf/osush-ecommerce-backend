package db

import (
	"context"
	"osush/internal/log"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type LoggedDB struct {
	DB     *pgxpool.Pool
	logger log.Logger
	env    string
}

func NewLoggedDB(pool *pgxpool.Pool, logger log.Logger, env string) *LoggedDB {
	return &LoggedDB{
		DB:     pool,
		logger: logger,
		env:    env,
	}
}

func (ldb *LoggedDB) ExecContext(
	ctx context.Context,
	query string,
	args ...interface{},
) (pgconn.CommandTag, error) {
	start := time.Now()

	res, err := ldb.DB.Exec(ctx, query, args...)

	if ldb.env == "development" {
		ldb.logger.Info(
			"SQL Exec",
			log.String("query", query),
			log.Any("args", args),
			log.Duration("duration", time.Since(start)),
			log.Error(err),
		)
	}

	return res, err
}

func (ldb *LoggedDB) QueryContext(
	ctx context.Context,
	query string,
	args ...interface{},
) (pgx.Rows, error) {
	start := time.Now()

	rows, err := ldb.DB.Query(ctx, query, args...)

	if ldb.env == "development" {
		ldb.logger.Info(
			"SQL Query",
			log.String("query", query),
			log.Any("args", args),
			log.Duration("duration", time.Since(start)),
			log.Error(err),
		)
	}

	return rows, err
}

func (ldb *LoggedDB) QueryRowContext(
	ctx context.Context,
	query string,
	args ...interface{},
) pgx.Row {
	start := time.Now()

	row := ldb.DB.QueryRow(ctx, query, args...)

	if ldb.env == "development" {
		ldb.logger.Info(
			"SQL QueryRow",
			log.String("query", query),
			log.Any("args", args),
			log.Duration("duration", time.Since(start)),
		)
	}

	return row
}
