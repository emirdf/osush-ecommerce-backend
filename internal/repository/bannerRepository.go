package repository

import (
	"context"
	"errors"
	"osush/internal/db"
	"osush/internal/log"
	"osush/internal/models"
	"time"
)

type BannerRepository struct {
	db     *db.LoggedDB
	logger log.Logger
}

func NewBannerRepository(db *db.LoggedDB, logger log.Logger) *BannerRepository {
	return &BannerRepository{
		db:     db,
		logger: logger,
	}
}

func (r *BannerRepository) CreateBanner(
	ctx context.Context,
	banner models.Banner,
) (*int, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("creating banner")

	titleDictionaryID, err := CreateDictionary(ctx, r.db, banner.TitleTm, banner.TitleRu)
	if err != nil {
		r.logger.Error("failed to create banner title dictionary",
			log.Error(err),
		)

		return nil, errors.New("failed to create banner title")
	}

	r.logger.Debug("banner title dictionary created",
		log.Int("dictionary_id", *titleDictionaryID),
	)

	var descriptionDictionaryID *int

	if banner.DescriptionTm != "" && banner.DescriptionRu != "" {

		descriptionDictionaryID, err = CreateDictionary(
			ctx,
			r.db,
			banner.DescriptionTm,
			banner.DescriptionRu,
		)
		if err != nil {
			r.logger.Error(
				"failed to create banner description dictionary",
				log.Error(err),
			)

			return nil, errors.New("failed to create banner description")
		}

		r.logger.Debug("banner description dictionary created",
			log.Int("description_dictionary_id", *descriptionDictionaryID),
		)
	}

	var bannerID int

	query := `
		INSERT INTO banners (
			title_dictionary_id,
			url,
			description_dictionary_id,
			phone_number,
			display_from,
			display_to
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err = r.db.QueryRowContext(
		ctx,
		query,
		titleDictionaryID,
		banner.URL,
		descriptionDictionaryID,
		banner.PhoneNumber,
		banner.DisplayFrom,
		banner.DisplayTo,
	).Scan(&bannerID)

	if err != nil {
		r.logger.Error("failed to create banner",
			log.Error(err),
		)

		return nil, errors.New("failed to create banner")
	}

	r.logger.Info("banner created successfully",
		log.Int("banner_id", bannerID),
	)

	return &bannerID, nil
}

func (r *BannerRepository) CreateBannerImages(
	ctx context.Context,
	images []models.BannerImage,
) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if len(images) == 0 {
		r.logger.Warn("no banner images to create")
		return nil
	}

	bannerID := images[0].BannerID

	r.logger.Debug("creating banner images",
		log.Int("banner_id", bannerID),
		log.Int("count", len(images)),
	)

	query := `
		INSERT INTO image_banners (
			banner_id,
			image_path,
			is_main
		)
		VALUES ($1, $2, $3)
	`

	for _, image := range images {

		r.logger.Debug("creating banner image",
			log.Int("banner_id", image.BannerID),
			log.String("image_path", image.ImagePath),
			log.Bool("is_main", image.IsMain),
		)

		_, err := r.db.ExecContext(
			ctx,
			query,
			image.BannerID,
			image.ImagePath,
			image.IsMain,
		)

		if err != nil {
			r.logger.Error("failed to create banner image",
				log.Int("banner_id", image.BannerID),
				log.String("image_path", image.ImagePath),
				log.Bool("is_main", image.IsMain),
				log.Error(err),
			)

			return errors.New("failed to create banner image")
		}
	}

	r.logger.Info("banner images created successfully",
		log.Int("banner_id", bannerID),
		log.Int("count", len(images)),
	)

	return nil
}

func (r *BannerRepository) GetBannerList(
	ctx context.Context,
) ([]models.Banner, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("getting banners")

	query := `
		SELECT
			b.id,
			b.title_dictionary_id,
			b.url,
			b.description_dictionary_id,
			b.phone_number,
			b.display_from,
			b.display_to
		FROM banners b
		WHERE b.display_from <= CURRENT_TIMESTAMP AND CURRENT_TIMESTAMP <= b.display_to
		ORDER BY b.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		r.logger.Error("failed to get banners",
			log.Error(err),
		)

		return nil, errors.New("failed to get banners")
	}
	defer rows.Close()

	var banners []models.Banner

	for rows.Next() {
		var banner models.Banner

		err := rows.Scan(
			&banner.ID,
			&banner.TitleDictionaryID,
			&banner.URL,
			&banner.DescriptionDictionaryID,
			&banner.PhoneNumber,
			&banner.DisplayFrom,
			&banner.DisplayTo,
		)

		if err != nil {
			r.logger.Error("failed to scan banner",
				log.Error(err),
			)

			return nil, errors.New("failed to scan banner")
		}

		titleTm, titleRu, err := GetDictionaryByID(
			ctx,
			r.db,
			banner.TitleDictionaryID,
		)
		if err != nil {
			r.logger.Error("failed to get banner title dictionary",
				log.Error(err),
			)

			return nil, errors.New("failed to get banner title")
		}
		banner.TitleTm = *titleTm
		banner.TitleRu = *titleRu

		if banner.DescriptionDictionaryID != nil {
			descriptionTm, descriptionRu, err := GetDictionaryByID(
				ctx,
				r.db,
				*banner.DescriptionDictionaryID,
			)
			if err != nil {
				r.logger.Error("failed to get banner description dictionary",
					log.String("banner_title_tm", banner.TitleTm),
					log.String("banner_title_ru", banner.TitleRu),
					log.Error(err),
				)

				return nil, errors.New("failed to get banner description")
			}
			banner.DescriptionTm = *descriptionTm
			banner.DescriptionRu = *descriptionRu
		}

		banners = append(banners, banner)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("failed while iterating banners",
			log.Error(err),
		)

		return nil, errors.New("failed to get banners")
	}

	r.logger.Info("banners retrieved successfully",
		log.Int("count", len(banners)),
	)

	return banners, nil
}

func (r *BannerRepository) GetBannerImages(
	ctx context.Context,
	bannerID int,
) ([]models.BannerImage, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("getting banner images",
		log.Int("banner_id", bannerID),
	)

	query := `
		SELECT
			id,
			banner_id,
			image_path,
			is_main
		FROM image_banners
		WHERE banner_id = $1
		ORDER BY is_main DESC, id ASC
	`

	rows, err := r.db.QueryContext(ctx, query, bannerID)
	if err != nil {
		r.logger.Error("failed to get banner images",
			log.Int("banner_id", bannerID),
			log.Error(err),
		)

		return nil, errors.New("failed to get banner images")
	}
	defer rows.Close()

	var images []models.BannerImage

	for rows.Next() {
		var image models.BannerImage

		if err := rows.Scan(
			&image.ID,
			&image.BannerID,
			&image.ImagePath,
			&image.IsMain,
		); err != nil {
			r.logger.Error("failed to scan banner image",
				log.Int("banner_id", bannerID),
				log.Error(err),
			)

			return nil, errors.New("failed to scan banner image")
		}

		images = append(images, image)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("failed while iterating banner images",
			log.Int("banner_id", bannerID),
			log.Error(err),
		)

		return nil, errors.New("failed to get banner images")
	}

	r.logger.Debug("banner images retrieved",
		log.Int("banner_id", bannerID),
		log.Int("count", len(images)),
	)

	return images, nil
}
