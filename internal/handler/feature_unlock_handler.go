package handler

import (
	"net/http"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/service"
	"github.com/gin-gonic/gin"
)

type FeatureUnlockHandler struct {
	featureService *service.FeatureUnlockService
}

func NewFeatureUnlockHandler(featureService *service.FeatureUnlockService) *FeatureUnlockHandler {
	return &FeatureUnlockHandler{
		featureService: featureService,
	}
}

// GetAllFeatures godoc
// @Summary Get all features
// @Description Get all feature definitions grouped by level
// @Tags features
// @Accept json
// @Produce json
// @Success 200 {object} []dto.FeaturesByLevelResponse
// @Router /api/v1/features [get]
func (h *FeatureUnlockHandler) GetAllFeatures(c *gin.Context) {
	ctx := c.Request.Context()
	features, err := h.featureService.GetAllFeatures(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil daftar fitur"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(features, "Daftar fitur berhasil diambil"))
}

// GetFeaturesByCategory godoc
// @Summary Get features by category
// @Description Get features grouped by category
// @Tags features
// @Accept json
// @Produce json
// @Param category path string true "Category key"
// @Success 200 {object} []dto.FeatureUnlockResponse
// @Router /api/v1/features/category/{category} [get]
func (h *FeatureUnlockHandler) GetFeaturesByCategory(c *gin.Context) {
	ctx := c.Request.Context()
	category := c.Param("category")

	features, err := h.featureService.GetFeaturesByCategory(ctx, category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil fitur berdasarkan kategori"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(features, "Fitur berdasarkan kategori berhasil diambil"))
}

// GetFeatureCategories godoc
// @Summary Get feature categories
// @Description Get all feature categories
// @Tags features
// @Accept json
// @Produce json
// @Success 200 {object} []dto.FeatureCategoryInfo
// @Router /api/v1/features/categories [get]
func (h *FeatureUnlockHandler) GetFeatureCategories(c *gin.Context) {
	ctx := c.Request.Context()
	categories := h.featureService.GetFeatureCategories(ctx)

	c.JSON(http.StatusOK, dto.SuccessResponse(categories, "Kategori fitur berhasil diambil"))
}

// GetUserFeatures godoc
// @Summary Get user features
// @Description Get current user's feature unlock status
// @Tags features
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.UserFeaturesResponse
// @Router /api/v1/features/my-features [get]
func (h *FeatureUnlockHandler) GetUserFeatures(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	features, err := h.featureService.GetUserFeatures(ctx, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil fitur user"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(features, "Fitur user berhasil diambil"))
}

// CheckFeatureAccess godoc
// @Summary Check feature access
// @Description Check if user has access to a specific feature
// @Tags features
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param featureKey path string true "Feature key"
// @Success 200 {object} dto.FeatureAccessResponse
// @Router /api/v1/features/check/{featureKey} [get]
func (h *FeatureUnlockHandler) CheckFeatureAccess(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	featureKey := c.Param("featureKey")

	access, err := h.featureService.CheckFeatureAccess(ctx, userID.(uint), featureKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal memeriksa akses fitur"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(access, "Status akses fitur berhasil diperiksa"))
}

// GetUpcomingFeatures godoc
// @Summary Get upcoming features
// @Description Get features that will be unlocked at upcoming levels
// @Tags features
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} []dto.LockedFeatureResponse
// @Router /api/v1/features/upcoming [get]
func (h *FeatureUnlockHandler) GetUpcomingFeatures(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	features, err := h.featureService.GetUpcomingFeatures(ctx, userID.(uint), 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil fitur mendatang"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(features, "Fitur mendatang berhasil diambil"))
}
