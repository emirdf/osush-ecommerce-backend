package app

import (
	"context"
	"os"
	"os/signal"
	"osush/internal/config"
	"osush/internal/db"
	"osush/internal/firebase"
	"osush/internal/handlers"
	"osush/internal/job"
	"osush/internal/log"
	"osush/internal/minio"
	"osush/internal/repository"
	minioRepo "osush/internal/repository/minio"
	"osush/internal/router"
	"osush/internal/telemetry"
	"osush/internal/usecase"
	"syscall"

	"github.com/robfig/cron/v3"
)

func Run(logger log.Logger) error {
	cfg := config.Load("railwayConfig") //railwayConfig    development

	logger.Info("loading configuration",
		log.String("db_host", cfg.DBHost),
		log.String("server_port", cfg.ServerPort),
	)

	database, err := db.NewDB(cfg, logger)
	if err != nil {
		return err
	}
	defer database.Close()

	minioClient, err := minio.NewClient(
		cfg.MinioEndpoint,
		cfg.MinioAccessKey,
		cfg.MinioSecretKey,
		cfg.MinioUseSSl,
	)
	if err != nil {
		logger.Error(
			"failed to create MinIO client",
			log.Error(err),
		)

		return err
	}

	firebaseService, err := firebase.NewFirebaseService(context.TODO())
	if err != nil {
		logger.Error(
			"failed to create firebase client",
			log.Error(err),
		)
		return err
	}

	healthChecker := telemetry.NewHealthChecker(database.GetDB())

	authRepo := repository.NewUserRepository(database, logger)
	clientRepo := repository.NewClientRepository(database, logger)
	categoryRepo := repository.NewCategoryRepository(database, logger)
	subcategoryRepo := repository.NewSubcategoryRepository(database, logger)
	productRepo := repository.NewProductRepository(database, logger)
	basketRepo := repository.NewBasketRepository(database, logger)
	orderProductsRepo := repository.NewOrderProductRepository(database, logger)
	orderRepo := repository.NewOrderRepository(database, logger)
	minioRepo := minioRepo.NewMinioRepository(minioClient, cfg.MinioBucket)
	productLikeRepo := repository.NewProductLikeRepository(database, logger)
	bannerRepo := repository.NewBannerRepository(database, logger)
	deviceTokenRepo := repository.NewDeviceTokenRepository(database, logger)
	notificationRepo := repository.NewNotificationRepository(database, logger)
	feedbackRepo := repository.NewFeedbackRepository(database, logger)
	appVersionRepo := repository.NewAppVersionRepository(database, logger)

	authUseCase := usecase.NewAuthUseCase(
		authRepo, minioRepo, categoryRepo, productRepo, clientRepo, orderProductsRepo, logger)
	clientUseCase := usecase.NewClientUseCase(clientRepo, minioRepo, orderRepo, logger)
	categoryUseCase := usecase.NewCategoryUseCase(categoryRepo, subcategoryRepo, minioRepo, logger)
	subcategoryUseCase := usecase.NewSubcategoryUseCase(subcategoryRepo, categoryRepo, productRepo, logger)
	productUseCase := usecase.NewProductUseCase(
		productRepo, subcategoryRepo, minioRepo, clientRepo,
		productLikeRepo, orderProductsRepo, logger)
	basketUseCase := usecase.NewBasketUseCase(basketRepo, productRepo, clientRepo, logger)
	orderUseCase := usecase.NewOrderUseCase(orderRepo, clientRepo, basketRepo, orderProductsRepo, productRepo, deviceTokenRepo, firebaseService, logger)
	productLikeUseCase := usecase.NewProductLikeUseCase(productLikeRepo, clientRepo, productRepo, logger)
	bannerUseCase := usecase.NewBannerUseCase(bannerRepo, minioRepo, logger)
	deviceTokenUseCase := usecase.NewDeviceTokenUseCase(deviceTokenRepo, logger)
	notificationUseCase := usecase.NewNotificationUseCase(notificationRepo, deviceTokenRepo, firebaseService, logger)
	feedbackUseCase := usecase.NewFeedbackUseCase(feedbackRepo, logger)
	appVersionUseCase := usecase.NewAppVersionUseCase(appVersionRepo, logger)

	authHandler := handlers.NewAuthHandler(cfg, authUseCase, logger)
	clientHandler := handlers.NewClientHandler(cfg, clientUseCase, logger)
	categoryHandler := handlers.NewCategoryHandler(cfg, categoryUseCase, logger)
	subcategoryHandler := handlers.NewSubcategoryHandler(subcategoryUseCase, logger)
	productHandler := handlers.NewProductHandler(cfg, productUseCase, logger)
	basketHandler := handlers.NewBasketHandler(cfg, basketUseCase, logger)
	orderHandler := handlers.NewOrderHandler(cfg, orderUseCase, logger)
	productLikeHandler := handlers.NewProductLikeHandler(cfg, productLikeUseCase, logger)
	bannerHandler := handlers.NewBannerHandler(cfg, bannerUseCase, logger)
	deviceTokenHandler := handlers.NewDeviceTokenHandler(cfg, deviceTokenUseCase, logger)
	notificationHandler := handlers.NewNotificationHandler(cfg, notificationUseCase, logger)
	feedbackHandler := handlers.NewFeedbackHandler(cfg, feedbackUseCase, logger)
	appVersionHandler := handlers.NewAppVersionHandler(cfg, appVersionUseCase, logger)

	cleanNotificationJob := job.NewCleanNotificationJob(notificationUseCase, logger)

	r := router.NewRouter(cfg, logger, &router.RouterDeps{
		AuthHandler:         authHandler,
		ClientHandler:       clientHandler,
		CategoryHandler:     categoryHandler,
		SubcategoryHandler:  subcategoryHandler,
		ProductHandler:      productHandler,
		BasketHandler:       basketHandler,
		OrderHandler:        orderHandler,
		ProductLikeHandler:  productLikeHandler,
		BannerHandler:       bannerHandler,
		DeviceTokenHandler:  deviceTokenHandler,
		NotificationHandler: notificationHandler,
		FeedbackHandler:     feedbackHandler,
		AppVersionHandler:   appVersionHandler,
		HealthChecker:       healthChecker,

		ClientRepository: clientRepo,
	})

	logger.Info("starting server",
		log.String("port", cfg.ServerPort),
	)

	go func() {
		if err := r.Run(":" + cfg.ServerPort); err != nil {
			logger.Error(
				"Server failed",
				log.Error(err),
			)
		}
	}()

	c := cron.New()

	_, err = c.AddFunc(
		"0 0 * * *",
		cleanNotificationJob.Run,
	)
	if err != nil {
		logger.Error(
			"failed to register clean notification job",
			log.Error(err),
		)

		return err
	}

	c.Start()

	quit := make(chan os.Signal, 1)

	signal.Notify(
		quit,
		os.Interrupt,
		syscall.SIGTERM,
	)

	<-quit

	logger.Info("Shutting down application...")

	return nil
}
