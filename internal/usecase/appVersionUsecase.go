package usecase

import (
	"context"
	"errors"
	customError "osush/internal/errors"
	"osush/internal/handlers/request"
	"osush/internal/handlers/response"
	"osush/internal/log"
	"osush/internal/models"
	"osush/internal/repository"
	"osush/internal/version"
	"time"
)

type AppVersionUseCase struct {
	repo   *repository.AppVersionRepository
	logger log.Logger
}

func NewAppVersionUseCase(repo *repository.AppVersionRepository, logger log.Logger) *AppVersionUseCase {
	return &AppVersionUseCase{repo: repo, logger: logger}
}

func toAppVersionResponse(v models.AppVersion) response.AppVersionResponse {
	return response.AppVersionResponse{
		Platform:      v.Platform,
		MinVersion:    v.MinVersion,
		LatestVersion: v.LatestVersion,
		StoreURL:      v.StoreURL,
		// Stored as UTC wall time; RFC3339 ("...Z") lets clients show local time.
		UpdatedAt: v.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// Check returns the versions of platform. With a non-empty current version it
// also says what that app must do ("force", "optional" or "none").
func (u *AppVersionUseCase) Check(ctx context.Context, platform, current string) (*response.AppVersionResponse, error) {
	if !version.IsValidPlatform(platform) {
		return nil, customError.ErrBadRequest("platform must be android or ios")
	}
	if current != "" && !version.IsValid(current) {
		return nil, customError.ErrBadRequest("version: expected like 1.2.3 or 1.2.3+7")
	}

	v, err := u.repo.GetByPlatform(ctx, platform)
	if errors.Is(err, repository.ErrAppVersionNotFound) {
		return nil, customError.ErrNotFound("app version for " + platform)
	}
	if err != nil {
		return nil, err
	}

	res := toAppVersionResponse(*v)
	if current != "" {
		res.Update = version.UpdateType(current, v.MinVersion, v.LatestVersion)
	}
	return &res, nil
}

func (u *AppVersionUseCase) GetAll(ctx context.Context) (*response.AppVersionListResponse, error) {
	versions, err := u.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	res := &response.AppVersionListResponse{
		AppVersions: make([]response.AppVersionResponse, 0, len(versions)),
	}
	for _, v := range versions {
		res.AppVersions = append(res.AppVersions, toAppVersionResponse(v))
	}
	return res, nil
}

// Update sets the versions of one platform. The request must be validated.
func (u *AppVersionUseCase) Update(ctx context.Context, req request.AppVersionRequest) (*response.AppVersionResponse, error) {
	u.logger.Info("starting update app version",
		log.String("platform", req.Platform),
		log.String("min_version", req.MinVersion),
		log.String("latest_version", req.LatestVersion),
	)

	v, err := u.repo.Save(ctx, models.AppVersion{
		Platform:      req.Platform,
		MinVersion:    req.MinVersion,
		LatestVersion: req.LatestVersion,
		StoreURL:      req.StoreURL,
	})
	if err != nil {
		return nil, err
	}

	res := toAppVersionResponse(*v)
	return &res, nil
}
