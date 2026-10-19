package usecase

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"osush/internal/config"
	customError "osush/internal/errors"
	"osush/internal/handlers/request"
	"osush/internal/handlers/response"
	"osush/internal/log"
	"osush/internal/models"
	"osush/internal/repository"
	minioRepo "osush/internal/repository/minio"
	"strings"

	"github.com/bytedance/gopkg/util/logger"
	"golang.org/x/crypto/bcrypt"
)

type ClientUseCase struct {
	repo      *repository.ClientRepository
	minioRepo *minioRepo.MinioRepository
	orderRepo *repository.OrderRepository
	logger    log.Logger
}

func NewClientUseCase(
	repo *repository.ClientRepository,
	minioRepo *minioRepo.MinioRepository,
	orderRepo *repository.OrderRepository,
	logger log.Logger,
) *ClientUseCase {
	return &ClientUseCase{
		repo:      repo,
		minioRepo: minioRepo,
		orderRepo: orderRepo,
		logger:    logger,
	}
}

func (u *ClientUseCase) Register(
	ctx context.Context,
	req request.RegisterRequest,
	image *multipart.FileHeader,
	baseURL string,
) (*response.ProfileResponse, error) {

	logger.Info("starting save upload client avatar from minio",
		log.String("name", req.UserName),
	)

	var imagePath *string
	if image != nil {
		var err error

		imagePath, err = GenerateFileName(image, "client")
		if err != nil {
			u.logger.Error("failed to save client avatar from minio",
				log.String("name", req.UserName),
				log.Error(err),
			)
			return nil, errors.New("failed to save client avatar from minio")
		}

		file, err := image.Open()
		if err != nil {
			u.logger.Error(
				"failed to open client avatar",
				log.String("name", req.UserName),
				log.Error(err),
			)

			return nil, errors.New("failed to open client avatar")
		}
		defer file.Close()

		contentType := image.Header.Get("Content-Type")

		if err := u.minioRepo.Upload(
			ctx,
			*imagePath,
			file,
			image.Size,
			contentType,
		); err != nil {

			u.logger.Error(
				"failed to upload client avatar to minio",
				log.String("name", req.UserName),
				log.String("object_name", *imagePath),
				log.Error(err),
			)

			return nil, errors.New("failed to upload client avatar to minio")
		}

		logger.Debug("client avatar saved from minio",
			log.String("name", req.UserName),
			log.String("image_path", *imagePath),
		)
	}

	u.logger.Info("starting client registration",
		log.String("user_name", req.UserName),
	)

	exists, err := u.repo.ExistsByPhoneNumber(ctx, req.PhoneNumber)

	if err != nil {
		u.logger.Error("failed to check client phone_number",
			log.String("phone_number", req.PhoneNumber),
			log.Error(err),
		)

		return nil, err
	}

	if exists {
		u.logger.Warn("client phone_number already exists",
			log.String("phone_number", req.PhoneNumber),
		)

		return nil, customError.ErrAlreadyData("phone_number")
	}

	passwordHash, err := HashPassword(req.Password)

	if err != nil {
		u.logger.Error("failed to hash client password",
			log.Error(err),
		)

		return nil, errors.New("failed to create password")
	}

	client := models.Client{
		UserName:    req.UserName,
		Surname:     req.Surname,
		Address:     req.Address,
		Password:    passwordHash,
		ImagePath:   imagePath,
		PhoneNumber: req.PhoneNumber,
	}

	clientID, err := u.repo.CreateClient(
		ctx,
		client,
	)

	if err != nil {
		u.logger.Error("failed to register client",
			log.String("user_name", req.UserName),
			log.Error(err),
		)

		return nil, err
	}

	u.logger.Info("client registered successfully",
		log.Int("client_id", *clientID),
	)

	if baseURL != "" && client.ImagePath != nil {
		cleanPath := strings.ReplaceAll(*client.ImagePath, "\\", "/")
		imageCleanPath := fmt.Sprintf("%s/%s", baseURL, cleanPath)
		client.ImagePath = &imageCleanPath
	}

	return &response.ProfileResponse{
		ID:          *clientID,
		UserName:    client.UserName,
		Surname:     client.Surname,
		Address:     client.Address,
		ImagePath:   client.ImagePath,
		PhoneNumber: client.PhoneNumber,
	}, nil
}

func (u *ClientUseCase) UpdateProfile(
	ctx context.Context,
	clientID int,
	req request.UpdateRequest,
	image *multipart.FileHeader,
	baseURL string,
) (*response.ProfileResponse, error) {

	u.logger.Info("starting client profile update",
		log.Int("client_id", clientID),
	)

	currentClient, err := u.repo.GetByID(ctx, clientID)
	if err != nil {
		u.logger.Error("failed to get current client profile",
			log.Int("client_id", clientID),
			log.Error(err),
		)

		return nil, err
	}

	var imagePath *string
	if image != nil {
		logger.Info("starting save upload client avatar from minio",
			log.String("name", req.UserName),
		)

		imagePath, err = GenerateFileName(image, "client")
		if err != nil {
			u.logger.Error("failed to save client avatar from minio",
				log.String("name", req.UserName),
				log.Error(err),
			)
			return nil, errors.New("failed to save client avatar from minio")
		}

		file, err := image.Open()
		if err != nil {
			u.logger.Error(
				"failed to open client avatar",
				log.String("name", req.UserName),
				log.Error(err),
			)

			return nil, errors.New("failed to open client avatar")
		}
		defer file.Close()

		contentType := image.Header.Get("Content-Type")

		if err := u.minioRepo.Upload(
			ctx,
			*imagePath,
			file,
			image.Size,
			contentType,
		); err != nil {

			u.logger.Error(
				"failed to upload client avatar to minio",
				log.String("name", req.UserName),
				log.String("object_name", *imagePath),
				log.Error(err),
			)

			return nil, errors.New("failed to upload client avatar to minio")
		}

		logger.Debug("client avatar saved from minio",
			log.String("name", req.UserName),
			log.String("image_path", *imagePath),
		)

	}

	exists, err := u.repo.ExistsByPhoneNumberExceptID(
		ctx,
		req.PhoneNumber,
		clientID,
	)

	if err != nil {
		u.logger.Error("failed to check phone_number",
			log.Int("client_id", clientID),
			log.Error(err),
		)

		return nil, err
	}

	if exists {
		u.logger.Warn("phone_number already belongs to another client",
			log.Int("client_id", clientID),
			log.String("phone_number", req.PhoneNumber),
		)

		return nil, customError.ErrAlreadyData("phone_number")
	}

	client := models.Client{
		ID:          clientID,
		UserName:    req.UserName,
		Surname:     req.Surname,
		Address:     req.Address,
		PhoneNumber: req.PhoneNumber,
		ImagePath:   imagePath,
	}

	err = u.repo.UpdateProfile(ctx, client)
	if err != nil {
		u.logger.Error("failed to update client profile",
			log.Int("client_id", clientID),
			log.Error(err),
		)

		if imagePath != nil {
			if removeErr := u.minioRepo.Delete(ctx, *imagePath); removeErr != nil {
				u.logger.Warn("failed to remove new client avatar from minio",
					log.String("image_path", *imagePath),
					log.Error(removeErr),
				)
			}
		}

		return nil, err
	}

	if currentClient.ImagePath != nil {
		if err := u.minioRepo.Delete(ctx, *currentClient.ImagePath); err != nil {
			u.logger.Warn("failed to remove old client avatar",
				log.String("image_path", *currentClient.ImagePath),
				log.Error(err),
			)
		} else {
			u.logger.Info("old client avatar removed",
				log.String("image_path", *currentClient.ImagePath),
			)
		}
	}

	profile, err := u.GetProfile(ctx, clientID, baseURL)
	if err != nil {
		return nil, err
	}

	return profile, nil
}

func (u *ClientUseCase) Login(
	ctx context.Context,
	cfg *config.Config,
	req request.LoginClientRequest,
) (*string, error) {

	u.logger.Info("starting client login",
		log.String("phone_number", req.PhoneNumber),
	)

	client, err := u.repo.GetByPhoneNumber(
		ctx,
		req.PhoneNumber,
	)

	if err != nil {

		if customError.ErrNotFound("client").Is(err) {
			u.logger.Warn("authentication failed",
				log.String("phone_number", req.PhoneNumber),
			)

			return nil, customError.ErrUnauthorized
		}

		u.logger.Error("failed to get client for login",
			log.String("phone_number", req.PhoneNumber),
			log.Error(err),
		)

		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(client.Password),
		[]byte(req.Password),
	); err != nil {
		u.logger.Warn(
			"client authentication failed",
			log.String("phone_number", req.PhoneNumber),
		)

		return nil, customError.ErrUnauthorized
	}

	token, err := GenerateTokenPair(cfg, client.ID, "CLIENT")
	if err != nil {
		u.logger.Error(
			"failed to generate token",
			log.Int("client_id", client.ID),
			log.String("username", client.UserName),
			log.Error(err),
		)

		return nil, err
	}

	u.logger.Info(
		"login successful",
		log.Int("client_id", client.ID),
		log.String("username", client.UserName),
	)

	return &token, nil
}

func (u *ClientUseCase) Logout(
	ctx context.Context,
	token string,
) error {
	u.logger.Info("starting client logout")

	u.logger.Debug("adding client token to blacklist")

	err := u.repo.Logout(ctx, token)
	if err != nil {
		u.logger.Error("failed to add client token to blacklist",
			log.Error(err),
		)

		return err
	}

	u.logger.Info("client logout successful")

	return nil
}

func (u *ClientUseCase) GetProfile(
	ctx context.Context,
	clientID int,
	baseURL string,
) (*response.ProfileResponse, error) {

	u.logger.Info("starting get client profile",
		log.Int("client_id", clientID),
	)

	client, err := u.repo.GetByID(ctx, clientID)
	if err != nil {
		u.logger.Error("failed to get client profile",
			log.Int("client_id", clientID),
			log.Error(err),
		)

		return nil, err
	}

	var imagePath *string

	if client.ImagePath != nil {
		cleanPath := strings.ReplaceAll(*client.ImagePath, "\\", "/")

		fullPath := fmt.Sprintf("%s/%s", baseURL, cleanPath)

		imagePath = &fullPath
	}

	u.logger.Info("client profile retrieved successfully",
		log.Int("client_id", clientID),
	)

	return &response.ProfileResponse{
		ID:          client.ID,
		UserName:    client.UserName,
		Surname:     client.Surname,
		Address:     client.Address,
		PhoneNumber: client.PhoneNumber,
		ImagePath:   imagePath,
	}, nil
}
func (u *ClientUseCase) GetClientList(
	ctx context.Context,
	search string,
	limit int,
	offset int,
) (*response.ClientListResponse, error) {

	u.logger.Debug(
		"getting client list",
		log.String("search", search),
		log.Int("limit", limit),
		log.Int("offset", offset),
	)

	clients, total, err := u.repo.GetClientList(ctx, search, limit, offset)

	if err != nil {
		u.logger.Error(
			"failed to get client list",
			log.Error(err),
		)

		return nil, err
	}

	result := response.ClientListResponse{
		Total:   *total,
		Clients: make([]response.Client, 0, len(*clients)),
	}

	for _, client := range *clients {

		orderCount, totalSpent, err :=
			u.orderRepo.GetClientTotalOrderAndSpend(ctx, client.ID)

		if err != nil {
			u.logger.Error(
				"failed to get client order count and total spent",
				log.Int("client_id", client.ID),
				log.Error(err),
			)

			return nil, err
		}

		clientResp := response.Client{
			ID:          client.ID,
			UserName:    client.UserName,
			Surname:     client.Surname,
			PhoneNumber: client.PhoneNumber,
			OrderCount:  *orderCount,
			TotalSpent:  *totalSpent,
		}

		result.Clients = append(result.Clients, clientResp)
	}

	return &result, nil
}
