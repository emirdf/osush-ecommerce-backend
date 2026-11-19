package router

import (
	"osush/internal/config"
	"osush/internal/handlers"
	"osush/internal/log"
	InternalMiddlewarePkg "osush/internal/middleware"
	"osush/internal/repository"
	"osush/internal/telemetry"

	"github.com/gin-gonic/gin"
)

type RouterDeps struct {
	AuthHandler         *handlers.AuthHandler
	ClientHandler       *handlers.ClientHandler
	CategoryHandler     *handlers.CategoryHandler
	SubcategoryHandler  *handlers.SubcategoryHandler
	ProductHandler      *handlers.ProductHandler
	BasketHandler       *handlers.BasketHandler
	OrderHandler        *handlers.OrderHandler
	ProductLikeHandler  *handlers.ProductLikeHandler
	BannerHandler       *handlers.BannerHandler
	DeviceTokenHandler  *handlers.DeviceTokenHandler
	NotificationHandler *handlers.NotificationHandler
	FeedbackHandler     *handlers.FeedbackHandler
	HealthChecker       *telemetry.HealthChecker
	ClientRepository    *repository.ClientRepository
}

func NewRouter(cfg *config.Config, logger log.Logger, deps *RouterDeps) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(InternalMiddlewarePkg.CORSMiddleware())
	r.Use(InternalMiddlewarePkg.LoggingMiddleware(logger))

	// r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	if deps.HealthChecker != nil {
		r.GET("/health", deps.HealthChecker.HealthCheck)
		r.GET("/health/ready", deps.HealthChecker.ReadinessCheck)
		r.GET("/health/live", deps.HealthChecker.LivenessCheck)
	} else {
		r.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})
	}

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		auth.POST("/login", InternalMiddlewarePkg.BodySizeLimit(1*1024*1024), deps.AuthHandler.Login)
		auth.PUT(
			"",
			InternalMiddlewarePkg.AdminOnly(cfg, deps.ClientRepository),
			InternalMiddlewarePkg.BodySizeLimit(10*1024*1024),
			deps.AuthHandler.UpdateProfile)
		auth.GET(
			"",
			InternalMiddlewarePkg.AdminOnly(cfg, deps.ClientRepository),
			InternalMiddlewarePkg.BodySizeLimit(1*1024*1024),
			deps.AuthHandler.GetProfile)
		auth.GET(
			"/index",
			InternalMiddlewarePkg.AdminOnly(cfg, deps.ClientRepository),
			InternalMiddlewarePkg.BodySizeLimit(1*1024*1024),
			deps.AuthHandler.GetAdminIndexPage)
	}

	{
		client := v1.Group("/client")
		client.POST("/register", InternalMiddlewarePkg.BodySizeLimit(10*1024*1024), deps.ClientHandler.Register)
		client.POST("/login", InternalMiddlewarePkg.BodySizeLimit(1*1024*1024), deps.ClientHandler.Login)
		client.POST(
			"/logout",
			InternalMiddlewarePkg.BodySizeLimit(1*1024*1024),
			InternalMiddlewarePkg.AuthMiddleware(cfg, deps.ClientRepository),
			deps.ClientHandler.Logout)
		client.PUT(
			"",
			InternalMiddlewarePkg.BodySizeLimit(1*1024*1024),
			InternalMiddlewarePkg.AuthMiddleware(cfg, deps.ClientRepository),
			deps.ClientHandler.UpdateProfile)
		client.GET(
			"",
			InternalMiddlewarePkg.BodySizeLimit(1*1024*1024),
			InternalMiddlewarePkg.AuthMiddleware(cfg, deps.ClientRepository),
			deps.ClientHandler.GetProfile)
		client.DELETE(
			"/delete",
			InternalMiddlewarePkg.BodySizeLimit(1*1024*1024),
			InternalMiddlewarePkg.AuthMiddleware(cfg, deps.ClientRepository),
			deps.ClientHandler.DeleteAccount)
		client.GET(
			"/list",
			InternalMiddlewarePkg.BodySizeLimit(1*1024*1024),
			InternalMiddlewarePkg.AdminOnly(cfg, deps.ClientRepository),
			deps.ClientHandler.GetClientList)
	}

	{
		category := v1.Group("/category")
		category.POST(
			"",
			InternalMiddlewarePkg.AdminOnly(cfg, deps.ClientRepository),
			InternalMiddlewarePkg.BodySizeLimit(10*1024*1024),
			deps.CategoryHandler.CreateCategory)
		category.GET("/:id", InternalMiddlewarePkg.BodySizeLimit(1*1024*1024), deps.CategoryHandler.GetCategory)
		category.GET("", InternalMiddlewarePkg.BodySizeLimit(1*1024*1024), deps.CategoryHandler.GetCategoryList)
		category.PUT(
			"/:id",
			InternalMiddlewarePkg.AdminOnly(cfg, deps.ClientRepository),
			InternalMiddlewarePkg.BodySizeLimit(10*1024*1024),
			deps.CategoryHandler.UpdateCategory)
		category.DELETE(
			"/:id",
			InternalMiddlewarePkg.AdminOnly(cfg, deps.ClientRepository),
			InternalMiddlewarePkg.BodySizeLimit(1*1024*1024),
			deps.CategoryHandler.DeleteCategory)
	}

	{
		subcategory := v1.Group("/subcategory")
		subcategory.POST(
			"",
			InternalMiddlewarePkg.AdminOnly(cfg, deps.ClientRepository),
			InternalMiddlewarePkg.BodySizeLimit(10*1024*1024),
			deps.SubcategoryHandler.CreateSubcategory)
		subcategory.GET("/:id", InternalMiddlewarePkg.BodySizeLimit(1*1024*1024), deps.SubcategoryHandler.GetSubcategory)
		subcategory.GET("", InternalMiddlewarePkg.BodySizeLimit(1*1024*1024), deps.SubcategoryHandler.GetSubcategoryList)
		subcategory.PUT(
			"/:id",
			InternalMiddlewarePkg.AdminOnly(cfg, deps.ClientRepository),
			InternalMiddlewarePkg.BodySizeLimit(10*1024*1024),
			deps.SubcategoryHandler.UpdateSubcategory)
		subcategory.DELETE(
			"/:id",
			InternalMiddlewarePkg.AdminOnly(cfg, deps.ClientRepository),
			InternalMiddlewarePkg.BodySizeLimit(1*1024*1024),
			deps.SubcategoryHandler.DeleteSubcategory)
	}

	{
		product := v1.Group("/product")
		product.POST(
			"",
			InternalMiddlewarePkg.AdminOnly(cfg, deps.ClientRepository),
			InternalMiddlewarePkg.BodySizeLimit(10*1024*1024),
			deps.ProductHandler.CreateProduct)
		product.GET(
			"/:id",
			InternalMiddlewarePkg.BodySizeLimit(1*1024*1024),
			InternalMiddlewarePkg.OptionalAuthMiddleware(cfg, deps.ClientRepository),
			deps.ProductHandler.GetProduct)
		product.GET(
			"",
			InternalMiddlewarePkg.BodySizeLimit(1*1024*1024),
			InternalMiddlewarePkg.OptionalAuthMiddleware(cfg, deps.ClientRepository),
			deps.ProductHandler.GetProductList)
		product.GET(
			"/index",
			InternalMiddlewarePkg.BodySizeLimit(1*1024*1024),
			InternalMiddlewarePkg.OptionalAuthMiddleware(cfg, deps.ClientRepository),
			deps.ProductHandler.GetIndexProducts)
		product.PUT(
			"/:id",
			InternalMiddlewarePkg.AdminOnly(cfg, deps.ClientRepository),
			InternalMiddlewarePkg.BodySizeLimit(10*1024*1024),
			deps.ProductHandler.UpdateProduct)
		product.DELETE(
			"/:id",
			InternalMiddlewarePkg.AdminOnly(cfg, deps.ClientRepository),
			InternalMiddlewarePkg.BodySizeLimit(1*1024*1024),
			deps.ProductHandler.DeleteProduct)
	}

	{
		productLike := v1.Group("/product-like")
		productLike.PUT(
			"/:id",
			InternalMiddlewarePkg.BodySizeLimit(1*1024*1024),
			InternalMiddlewarePkg.AuthMiddleware(cfg, deps.ClientRepository),
			deps.ProductLikeHandler.ToggleLikeProduct)
		productLike.GET(
			"",
			InternalMiddlewarePkg.BodySizeLimit(1*1024*1024),
			InternalMiddlewarePkg.AuthMiddleware(cfg, deps.ClientRepository),
			deps.ProductLikeHandler.GetLikedProducts)
	}

	{
		basket := v1.Group("/basket")
		basket.POST(
			"",
			InternalMiddlewarePkg.BodySizeLimit(1*1024*1024),
			InternalMiddlewarePkg.AuthMiddleware(cfg, deps.ClientRepository),
			deps.BasketHandler.AddProductFromBasket)
		basket.DELETE(
			"",
			InternalMiddlewarePkg.BodySizeLimit(1*1024*1024),
			InternalMiddlewarePkg.AuthMiddleware(cfg, deps.ClientRepository),
			deps.BasketHandler.DeleteProductFromBasket)
		basket.DELETE(
			"/product/:id",
			InternalMiddlewarePkg.BodySizeLimit(1*1024*1024),
			InternalMiddlewarePkg.AuthMiddleware(cfg, deps.ClientRepository),
			deps.BasketHandler.CleanProductFromBasket)
		basket.GET(
			"",
			InternalMiddlewarePkg.BodySizeLimit(1*1024*1024),
			InternalMiddlewarePkg.AuthMiddleware(cfg, deps.ClientRepository),
			deps.BasketHandler.GetAllBasketProducts)

	}

	{
		order := v1.Group("/product/order")
		order.POST(
			"",
			InternalMiddlewarePkg.BodySizeLimit(10*1024*1024),
			InternalMiddlewarePkg.AuthMiddleware(cfg, deps.ClientRepository),
			deps.OrderHandler.CreateOrder)
		order.GET(
			"/:id",
			InternalMiddlewarePkg.BodySizeLimit(1*1024*1024),
			InternalMiddlewarePkg.AdminOnly(cfg, deps.ClientRepository),
			deps.OrderHandler.GetOrder)
		order.GET(
			"/:id/client",
			InternalMiddlewarePkg.BodySizeLimit(1*1024*1024),
			InternalMiddlewarePkg.AuthMiddleware(cfg, deps.ClientRepository),
			deps.OrderHandler.GetOrder)
		order.GET(
			"",
			InternalMiddlewarePkg.BodySizeLimit(1*1024*1024),
			InternalMiddlewarePkg.AdminOnly(cfg, deps.ClientRepository),
			deps.OrderHandler.GetOrderList)
		order.GET(
			"/client",
			InternalMiddlewarePkg.BodySizeLimit(1*1024*1024),
			InternalMiddlewarePkg.AuthMiddleware(cfg, deps.ClientRepository),
			deps.OrderHandler.GetOrderClientList)
		order.PUT(
			"/send/:id",
			InternalMiddlewarePkg.BodySizeLimit(1*1024*1024),
			InternalMiddlewarePkg.AdminOnly(cfg, deps.ClientRepository),
			deps.OrderHandler.MarkOrderAsSend)
	}

	{
		banner := v1.Group("/banner")
		banner.POST(
			"",
			InternalMiddlewarePkg.AdminOnly(cfg, deps.ClientRepository),
			InternalMiddlewarePkg.BodySizeLimit(10*1024*1024),
			deps.BannerHandler.CreateBanner)
		banner.GET(
			"",
			InternalMiddlewarePkg.BodySizeLimit(1*1024*1024),
			deps.BannerHandler.GetBannerList)
	}

	{
		deviceToken := v1.Group("/device-token")
		deviceToken.POST(
			"",
			InternalMiddlewarePkg.OptionalAuthMiddleware(cfg, deps.ClientRepository),
			InternalMiddlewarePkg.BodySizeLimit(2*1024*1024),
			deps.DeviceTokenHandler.CreateDeviceToken)
	}

	{
		notification := v1.Group("/notification")
		notification.POST(
			"/admin",
			InternalMiddlewarePkg.AdminOnly(cfg, deps.ClientRepository),
			InternalMiddlewarePkg.BodySizeLimit(2*1024*1024),
			deps.NotificationHandler.CreateNotification)
		notification.GET(
			"",
			InternalMiddlewarePkg.OptionalAuthMiddleware(cfg, deps.ClientRepository),
			InternalMiddlewarePkg.BodySizeLimit(2*1024*1024),
			deps.NotificationHandler.GetAllNotifications)
		notification.GET(
			"/admin",
			InternalMiddlewarePkg.AdminOnly(cfg, deps.ClientRepository),
			InternalMiddlewarePkg.BodySizeLimit(2*1024*1024),
			deps.NotificationHandler.GetAllNotificationsForAdmin)
		notification.DELETE(
			"/admin/:id",
			InternalMiddlewarePkg.AdminOnly(cfg, deps.ClientRepository),
			InternalMiddlewarePkg.BodySizeLimit(2*1024*1024),
			deps.NotificationHandler.DeleteNotification)
	}

	{
		feedback := v1.Group("/feedback")
		feedback.POST(
			"",
			InternalMiddlewarePkg.OptionalAuthMiddleware(cfg, deps.ClientRepository),
			InternalMiddlewarePkg.BodySizeLimit(1*1024*1024),
			deps.FeedbackHandler.CreateFeedback)
		feedback.GET(
			"/admin",
			InternalMiddlewarePkg.AdminOnly(cfg, deps.ClientRepository),
			InternalMiddlewarePkg.BodySizeLimit(1*1024*1024),
			deps.FeedbackHandler.GetFeedbackList)
		feedback.DELETE(
			"/admin/:id",
			InternalMiddlewarePkg.AdminOnly(cfg, deps.ClientRepository),
			InternalMiddlewarePkg.BodySizeLimit(1*1024*1024),
			deps.FeedbackHandler.DeleteFeedback)
	}

	return r
}
