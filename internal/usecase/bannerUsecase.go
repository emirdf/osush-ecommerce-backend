package usecase

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"osush/internal/handlers/request"
	"osush/internal/handlers/response"
	"osush/internal/log"
	"osush/internal/models"
	"osush/internal/repository"
	minioRepo "osush/internal/repository/minio"
	"strings"
	"time"
)

type BannerUseCase struct {
	repo      *repository.BannerRepository
	minioRepo *minioRepo.MinioRepository
	logger    log.Logger
}

func NewBannerUseCase(
	repo *repository.BannerRepository,
	minioRepo *minioRepo.MinioRepository,
	logger log.Logger) *BannerUseCase {
	return &BannerUseCase{
		repo:      repo,
		minioRepo: minioRepo,
		logger:    logger,
	}
}

func (u *BannerUseCase) CreateBanner(
	ctx context.Context,
	req request.BannerRequest,
) (*int, error) {

	u.logger.Info("starting create banner")

	displayFrom, err := time.Parse("02-01-2006", req.DisplayFrom)
	if err != nil {
		return nil, errors.New("failed parsed display_from")
	}

	displayTo, err := time.Parse("02-01-2006", req.DisplayTo)
	if err != nil {
		return nil, errors.New("failed parsed display_to")
	}

	banner := models.Banner{
		TitleTm:       req.Title.Tm,
		TitleRu:       req.Title.Ru,
		DescriptionTm: req.Description.Tm,
		DescriptionRu: req.Description.Ru,
		URL:           req.URL,
		PhoneNumber:   req.PhoneNumber,
		DisplayFrom:   displayFrom,
		DisplayTo:     displayTo,
	}

	u.logger.Debug("creating banner in repository")

	bannerID, err := u.repo.CreateBanner(ctx, banner)
	if err != nil {
		u.logger.Error("failed to create banner",
			log.Error(err),
		)

		return nil, err
	}

	u.logger.Info("banner created successfully",
		log.Int("banner_id", *bannerID),
	)

	return bannerID, nil
}

func (u *BannerUseCase) AddImages(
	ctx context.Context,
	bannerID int,
	mainImage *multipart.FileHeader,
	images []*multipart.FileHeader,
) error {

	u.logger.Info("starting create banner images",
		log.Int("banner_id", bannerID),
		log.Int("additional_images_count", len(images)),
	)

	var imageBanners []models.BannerImage

	mainImagePath, err := GenerateFileName(mainImage, "main_img_banner")
	if err != nil {
		u.logger.Error("failed to generate banner main image name",
			log.String("name", mainImage.Filename),
			log.Error(err),
		)

		return errors.New("failed to generate banner main image name")
	}

	file, err := mainImage.Open()
	if err != nil {
		u.logger.Error("failed to open banner main image",
			log.String("name", mainImage.Filename),
			log.Error(err),
		)

		return errors.New("failed to open banner main image")
	}

	contentType := mainImage.Header.Get("Content-Type")

	if err := u.minioRepo.Upload(
		ctx,
		*mainImagePath,
		file,
		mainImage.Size,
		contentType,
	); err != nil {

		file.Close()

		u.logger.Error("failed to upload banner main image to minio",
			log.String("name", mainImage.Filename),
			log.String("object_name", *mainImagePath),
			log.Error(err),
		)

		return errors.New("failed to upload banner main image to minio")
	}

	file.Close()

	u.logger.Debug("banner main image saved to minio",
		log.String("name", mainImage.Filename),
		log.String("object_name", *mainImagePath),
	)

	imageBanners = append(imageBanners, models.BannerImage{
		BannerID:  bannerID,
		ImagePath: *mainImagePath,
		IsMain:    true,
	})

	for _, image := range images {

		imagePath, err := GenerateFileName(image, "banner")
		if err != nil {
			u.logger.Error("failed to generate banner image name",
				log.String("name", image.Filename),
				log.Error(err),
			)

			return errors.New("failed to generate banner image name")
		}

		file, err := image.Open()
		if err != nil {
			u.logger.Error("failed to open banner image",
				log.String("name", image.Filename),
				log.Error(err),
			)

			return errors.New("failed to open banner image")
		}

		contentType := image.Header.Get("Content-Type")

		if err := u.minioRepo.Upload(
			ctx,
			*imagePath,
			file,
			image.Size,
			contentType,
		); err != nil {

			file.Close()

			u.logger.Error("failed to upload banner image to minio",
				log.String("name", image.Filename),
				log.String("object_name", *imagePath),
				log.Error(err),
			)

			return errors.New("failed to upload banner image to minio")
		}

		file.Close()

		imageBanners = append(imageBanners, models.BannerImage{
			BannerID:  bannerID,
			ImagePath: *imagePath,
			IsMain:    false,
		})
	}

	u.logger.Debug("banner images prepared",
		log.Int("banner_id", bannerID),
		log.Int("total_images_count", len(imageBanners)),
		log.Int("additional_images_count", len(images)),
	)

	if err := u.repo.CreateBannerImages(ctx, imageBanners); err != nil {
		u.logger.Error("failed to create banner images",
			log.Int("banner_id", bannerID),
			log.Int("total_images_count", len(imageBanners)),
			log.Error(err),
		)

		return err
	}

	u.logger.Info("banner images created successfully",
		log.Int("banner_id", bannerID),
		log.Int("total_images_count", len(imageBanners)),
	)

	return nil
}

func (u *BannerUseCase) GetBannerList(
	ctx context.Context,
	baseURL string,
) (*response.BannersResponse, error) {

	u.logger.Info("starting get banners list")

	banners, err := u.repo.GetBannerList(ctx)
	if err != nil {
		u.logger.Error("failed to get banners list",
			log.Error(err),
		)

		return nil, err
	}

	result := response.BannersResponse{
		Total:   len(banners),
		Banners: []response.BannerResponse{},
	}

	for _, banner := range banners {

		images, err := u.repo.GetBannerImages(ctx, banner.ID)
		if err != nil {
			return nil, err
		}

		now := time.Now()

		currentDate := time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			0, 0, 0, 0,
			now.Location(),
		)

		displayFrom := time.Date(
			banner.DisplayFrom.Year(),
			banner.DisplayFrom.Month(),
			banner.DisplayFrom.Day(),
			0, 0, 0, 0,
			banner.DisplayFrom.Location(),
		)

		displayTo := time.Date(
			banner.DisplayTo.Year(),
			banner.DisplayTo.Month(),
			banner.DisplayTo.Day(),
			0, 0, 0, 0,
			banner.DisplayTo.Location(),
		)

		isActive := !currentDate.Before(displayFrom) &&
			!currentDate.After(displayTo)

		bannerResponse := response.BannerResponse{
			ID: banner.ID,
			Title: response.Dictionary{
				Tm: banner.TitleTm,
				Ru: banner.TitleRu,
			},
			URL:         banner.URL,
			PhoneNumber: banner.PhoneNumber,
			Images:      make([]string, 0),
			DisplayFrom: banner.DisplayFrom.Format("02-01-2006"),
			DisplayTo:   banner.DisplayTo.Format("02-01-2006"),
			IsActive:    isActive,
		}

		if banner.DescriptionDictionaryID != nil {
			bannerDescripton := response.Dictionary{
				Tm: banner.DescriptionTm,
				Ru: banner.DescriptionRu,
			}

			bannerResponse.Description = &bannerDescripton
		}

		for _, image := range images {

			if baseURL != "" && image.ImagePath != "" {
				cleanPath := strings.ReplaceAll(image.ImagePath, "\\", "/")
				image.ImagePath = fmt.Sprintf("%s/%s", baseURL, cleanPath)
			}

			if image.IsMain {
				bannerResponse.Image = image.ImagePath
			} else {
				bannerResponse.Images = append(
					bannerResponse.Images,
					image.ImagePath,
				)
			}
		}

		result.Banners = append(result.Banners, bannerResponse)
	}

	return &result, nil
}
