package repository

import (
	"context"
	"errors"
	"osush/internal/db"
	"osush/internal/log"
	"osush/internal/models"
	"time"

	"github.com/jackc/pgx/v4"
)

var ErrAppVersionNotFound = errors.New("app version not found")

type AppVersionRepository struct {
	db     *db.LoggedDB
	logger log.Logger
}

func NewAppVersionRepository(db *db.LoggedDB, logger log.Logger) *AppVersionRepository {
	return &AppVersionRepository{db: db, logger: logger}
}

func (r *AppVersionRepository) GetByPlatform(ctx context.Context, platform string) (*models.AppVersion, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var v models.AppVersion
	err := r.db.QueryRowContext(ctx, `
		SELECT platform, min_version, latest_version, store_url, updated_at
		FROM app_versions
		WHERE platform = $1
	`, platform).Scan(&v.Platform, &v.MinVersion, &v.LatestVersion, &v.StoreURL, &v.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAppVersionNotFound
	}
	if err != nil {
		r.logger.Error("failed to get app version",
			log.String("platform", platform),
			log.Error(err),
		)
		return nil, errors.New("failed to get app version")
	}

	return &v, nil
}

func (r *AppVersionRepository) GetAll(ctx context.Context) ([]models.AppVersion, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := r.db.QueryContext(ctx, `
		SELECT platform, min_version, latest_version, store_url, updated_at
		FROM app_versions
		ORDER BY platform
	`)
	if err != nil {
		r.logger.Error("failed to get app versions", log.Error(err))
		return nil, errors.New("failed to get app versions")
	}
	defer rows.Close()

	versions := make([]models.AppVersion, 0, 2)
	for rows.Next() {
		var v models.AppVersion
		if err := rows.Scan(&v.Platform, &v.MinVersion, &v.LatestVersion, &v.StoreURL, &v.UpdatedAt); err != nil {
			r.logger.Error("failed to scan app version", log.Error(err))
			return nil, errors.New("failed to scan app version")
		}
		versions = append(versions, v)
	}
	if err := rows.Err(); err != nil {
		r.logger.Error("failed while iterating app versions", log.Error(err))
		return nil, errors.New("failed to get app versions")
	}

	return versions, nil
}

// Save inserts or updates the versions of one platform.
func (r *AppVersionRepository) Save(ctx context.Context, v models.AppVersion) (*models.AppVersion, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var out models.AppVersion
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO app_versions (platform, min_version, latest_version, store_url)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (platform) DO UPDATE
		SET min_version = EXCLUDED.min_version,
			latest_version = EXCLUDED.latest_version,
			store_url = EXCLUDED.store_url,
			updated_at = (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
		RETURNING platform, min_version, latest_version, store_url, updated_at
	`, v.Platform, v.MinVersion, v.LatestVersion, v.StoreURL).
		Scan(&out.Platform, &out.MinVersion, &out.LatestVersion, &out.StoreURL, &out.UpdatedAt)
	if err != nil {
		r.logger.Error("failed to save app version",
			log.String("platform", v.Platform),
			log.Error(err),
		)
		return nil, errors.New("failed to save app version")
	}

	r.logger.Info("app version saved",
		log.String("platform", out.Platform),
		log.String("min_version", out.MinVersion),
		log.String("latest_version", out.LatestVersion),
	)

	return &out, nil
}
