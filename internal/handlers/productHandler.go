package handlers

import (
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"osush/internal/config"
	customError "osush/internal/errors"
	"osush/internal/handlers/request"
	"osush/internal/handlers/response"
	"osush/internal/log"
	"osush/internal/middleware"
	"osush/internal/usecase"
	"osush/internal/validation"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	cfg            *config.Config
	productUseCase *usecase.ProductUseCase
	logger         log.Logger
}

func NewProductHandler(cfg *config.Config, u *usecase.ProductUseCase, logger log.Logger) *ProductHandler {
	return &ProductHandler{
		cfg:            cfg,
		productUseCase: u,
		logger:         logger,
	}
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	var (
		additionalImages []*multipart.FileHeader
		req              request.ProductRequest
	)

	logger.Info("handling CreateProduct request")

	jsonData := c.PostForm("data")
	if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
		logger.Error("failed to parse product request data",
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	logger.Debug("product request data parsed",
		log.String("name", req.Name),
	)

	if err := req.Validate(); err != nil {
		logger.Warn("product request validation failed",
			log.String("name", req.Name),
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	logger.Info("product request is valid",
		log.String("name", req.Name),
	)

	mainImage, err := c.FormFile("mainImage")
	if err != nil {
		logger.Error("failed to get product main image",
			log.String("name", req.Name),
			log.Error(err),
		)

		if errors.Is(err, http.ErrMissingFile) {
			logger.Warn("product main image is missing",
				log.String("name", req.Name),
			)

			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Error: "mainImage is required",
			})
			return
		}

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	logger.Debug("product main image received",
		log.String("name", req.Name),
		log.String("file_name", mainImage.Filename),
		log.Int64("size", mainImage.Size),
	)

	if err := validation.ValidateImageSize(mainImage.Size); err != nil {
		logger.Warn("product main image size validation failed",
			log.String("name", req.Name),
			log.Int64("size", mainImage.Size),
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	contentType := mainImage.Header.Get("Content-Type")

	if err := validation.ValidateImageFormat(contentType); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	logger.Debug("product main image size is valid",
		log.String("name", req.Name),
		log.Int64("size", mainImage.Size),
	)

	form, err := c.MultipartForm()
	if err != nil {
		logger.Warn("failed to get multipart form for additional images",
			log.Error(err),
		)
	} else {
		additionalFiles := form.File["additionImages"]

		logger.Debug("additional product images received",
			log.Int("count", len(additionalFiles)),
		)

		for _, image := range additionalFiles {
			logger.Debug("saving product additional image",
				log.String("file_name", image.Filename),
				log.Int64("size", image.Size),
			)

			if err := validation.ValidateImageSize(image.Size); err != nil {
				logger.Warn("product addition images size validation failed",
					log.String("name", req.Name),
					log.Int64("size", image.Size),
					log.Error(err),
				)

				c.JSON(http.StatusBadRequest, response.ErrorResponse{
					Error: err.Error(),
				})
				return
			}

			contentType := image.Header.Get("Content-Type")

			if err := validation.ValidateImageFormat(contentType); err != nil {
				c.JSON(http.StatusBadRequest, response.ErrorResponse{
					Error: err.Error(),
				})
				return
			}

			additionalImages = append(additionalImages, image)
		}
	}

	productID, err := h.productUseCase.CreateProduct(ctx, req)
	if err != nil {
		logger.Error("failed to create product",
			log.String("name", req.Name),
			log.Error(err),
		)

		customError.HandleError(c, err)
		return
	}

	logger.Info("product created successfully",
		log.Int("product_id", *productID),
		log.String("name", req.Name),
	)

	logger.Debug("product upload directory prepared",
		log.Int("product_id", *productID),
	)

	logger.Info("starting save product main image",
		log.Int("product_id", *productID),
		log.String("file_name", mainImage.Filename),
	)

	err = h.productUseCase.AddImages(
		ctx,
		*productID,
		mainImage,
		additionalImages,
	)

	if err != nil {
		logger.Error("failed to save product images",
			log.Int("product_id", *productID),
			log.Error(err),
		)

		customError.HandleError(c, err)
		return
	}

	logger.Info("product created successfully",
		log.Int("product_id", *productID),
		log.String("name", req.Name),
	)

	baseURL := h.cfg.MinioBaseURL()

	product, err := h.productUseCase.GetProductByID(ctx, *productID, -1, baseURL)
	if err != nil {
		logger.Error("failed to get product",
			log.Int("id", *productID),
			log.Error(err),
		)

		customError.HandleError(c, err)
		return
	}

	logger.Info("GetProductByID request completed",
		log.Int("product_id", *productID),
	)

	c.JSON(http.StatusCreated, product)
}

func (h *ProductHandler) GetProduct(c *gin.Context) {

	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling GetProduct request")

	clientID := c.MustGet("clientId").(int)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Warn("invalid product id",
			log.String("id", c.Param("id")),
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "invalid product id",
		})
		return
	}

	logger.Debug("product id parsed",
		log.Int("product_id", id),
	)

	baseURL := h.cfg.MinioBaseURL()

	product, err := h.productUseCase.GetProductByID(ctx, id, clientID, baseURL)
	if err != nil {
		logger.Error("failed to get product",
			log.Int("product_id", id),
			log.Error(err),
		)

		customError.HandleError(c, err)
		return
	}

	logger.Info("GetProductByID request completed",
		log.Int("product_id", id),
	)

	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) GetProductList(c *gin.Context) {

	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling GetAllProducts request")

	var filter request.ProductFilter

	if err := c.ShouldBindQuery(&filter); err != nil {
		logger.Warn("invalid product filter",
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	logger.Debug("product filter parsed",
		log.String("search", filter.Search),
		log.String("sort_by_value", filter.SortByValue),
		log.String("limit", filter.Limit),
		log.String("offset", filter.Offset),
		log.Int("subcategory_id", filter.SubcategoryID),
		log.Int("category_id", filter.CategoryID),
	)

	offsetInt, limitInt, err := validation.ValidatePagination(filter.Offset, filter.Limit)
	if err != nil {
		logger.Error("failed validate pagination",
			log.String("limit", filter.Limit),
			log.String("offset", filter.Offset),
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	filter.OffsetInt = *offsetInt
	filter.LimitInt = *limitInt

	err = validation.ValidateSearchQuery(filter.Search)
	if err != nil {
		logger.Error("failed validate search",
			log.String("search", filter.Search),
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}
	clientID := c.MustGet("clientId").(int)

	baseURL := h.cfg.MinioBaseURL()

	products, err := h.productUseCase.GetProductList(ctx, filter, clientID, baseURL)
	if err != nil {
		logger.Error("failed to get products",
			log.Error(err),
		)

		customError.HandleError(c, err)
		return
	}

	logger.Info("GetAllProducts request completed",
		log.Int("count", len(products.Products)),
	)

	c.JSON(http.StatusOK, products)
}

func (h *ProductHandler) GetIndexProducts(c *gin.Context) {

	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling GetIndexProducts request")

	clientID := c.MustGet("clientId").(int)

	baseURL := h.cfg.MinioBaseURL()

	result, err := h.productUseCase.GetIndexProducts(ctx, baseURL, clientID)
	if err != nil {
		logger.Error("failed to get index products",
			log.Error(err),
		)

		customError.HandleError(c, err)
		return
	}

	logger.Info("GetIndexProducts request completed")

	c.JSON(http.StatusOK, result)
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	logger.Info("handling UpdateProduct request")

	id := c.Param("id")

	productID, err := strconv.Atoi(id)
	if err != nil {
		logger.Warn("invalid product id",
			log.String("id", id),
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "invalid product id",
		})
		return
	}

	var req request.UpdateProductRequest

	jsonData := c.PostForm("data")

	if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
		logger.Error("failed to parse update product request data",
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	logger.Debug("update product request data parsed",
		log.Int("product_id", productID),
		log.String("name", req.Name),
	)

	if err := req.Validate(); err != nil {
		logger.Warn("update product request validation failed",
			log.Int("product_id", productID),
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	mainImage, err := c.FormFile("mainImage")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			logger.Warn("product main image is missing",
				log.Int("product_id", productID),
			)

			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Error: "mainImage is required",
			})
			return
		}

		logger.Error("failed to get product main image",
			log.Int("product_id", productID),
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	if err := validation.ValidateImageSize(mainImage.Size); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	if err := validation.ValidateImageFormat(
		mainImage.Header.Get("Content-Type"),
	); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	var additionalImages []*multipart.FileHeader

	form, err := c.MultipartForm()
	if err == nil {
		additionalImages = form.File["additionImages"]
	}

	logger.Debug("additional product images received",
		log.Int("product_id", productID),
		log.Int("count", len(additionalImages)),
	)

	for _, image := range additionalImages {

		if err := validation.ValidateImageSize(image.Size); err != nil {
			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Error: err.Error(),
			})
			return
		}

		if err := validation.ValidateImageFormat(
			image.Header.Get("Content-Type"),
		); err != nil {
			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Error: err.Error(),
			})
			return
		}
	}

	if err := h.productUseCase.UpdateProduct(
		ctx,
		productID,
		req,
		mainImage,
		additionalImages,
	); err != nil {

		logger.Error("failed to update product",
			log.Int("product_id", productID),
			log.Error(err),
		)

		customError.HandleError(c, err)
		return
	}

	logger.Info("product updated successfully",
		log.Int("product_id", productID),
	)

	baseURL := h.cfg.MinioBaseURL()

	product, err := h.productUseCase.GetProductByID(
		ctx,
		productID,
		-1,
		baseURL,
	)
	if err != nil {
		logger.Error("failed to get updated product",
			log.Int("product_id", productID),
			log.Error(err),
		)

		customError.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	logger := middleware.GetLogger(c)
	ctx := c.Request.Context()

	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Error("invalid product id",
			log.Error(err),
		)

		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "invalid product id",
		})

		return
	}

	logger.Info("handling DeleteProduct request",
		log.Int("product_id", productID),
	)

	if err := h.productUseCase.DeleteProduct(
		ctx,
		productID,
	); err != nil {
		customError.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "product deleted successfully",
	})
}
