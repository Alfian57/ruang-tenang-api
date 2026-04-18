package handler

import (
	"net/http"
	"strconv"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/gin-gonic/gin"

	"github.com/Alfian57/ruang-tenang-api/internal/features/reward/application"
	"github.com/Alfian57/ruang-tenang-api/internal/features/reward/infrastructure"
)

type RewardHandler struct {
	rewardService application.RewardService
}

func NewRewardHandler(rewardService application.RewardService) *RewardHandler {
	return &RewardHandler{
		rewardService: rewardService,
	}
}

func (h *RewardHandler) requireUserID(c *gin.Context) (uint, bool) {
	rawUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Success: false,
			Message: "Unauthorized",
		})
		return 0, false
	}

	userID, ok := rawUserID.(uint)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Success: false,
			Message: "Unauthorized",
		})
		return 0, false
	}

	return userID, true
}

func (h *RewardHandler) parseRewardID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: "ID hadiah tidak valid",
		})
		return 0, false
	}

	return uint(id), true
}

func (h *RewardHandler) bindJSON(c *gin.Context, target interface{}, invalidMessage string) bool {
	if err := c.ShouldBindJSON(target); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: invalidMessage,
		})
		return false
	}

	return true
}

func (h *RewardHandler) parsePageAndSize(c *gin.Context) (int, int) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	return page, pageSize
}

// ============ Member Endpoints ============

// GetAvailableRewards godoc
// @Summary Get available rewards
// @Description Get all active rewards that can be claimed with gold coins
// @Tags Rewards
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response{data=[]model.Reward}
// @Router /rewards [get]
func (h *RewardHandler) GetAvailableRewards(c *gin.Context) {
	ctx := c.Request.Context()

	rewards, err := h.rewardService.GetAvailableRewards(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Message: "Gagal mengambil daftar hadiah",
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Message: "Berhasil mengambil daftar hadiah",
		Data:    rewards,
	})
}

// GetRewardDetail godoc
// @Summary Get reward detail
// @Description Get detail of a specific reward
// @Tags Rewards
// @Produce json
// @Security BearerAuth
// @Param id path int true "Reward ID"
// @Success 200 {object} dto.Response{data=model.Reward}
// @Router /rewards/{id} [get]
func (h *RewardHandler) GetRewardDetail(c *gin.Context) {
	ctx := c.Request.Context()
	id, ok := h.parseRewardID(c)
	if !ok {
		return
	}

	reward, err := h.rewardService.GetRewardByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.Response{
			Success: false,
			Message: "Hadiah tidak ditemukan",
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Message: "Berhasil mengambil detail hadiah",
		Data:    reward,
	})
}

// ClaimReward godoc
// @Summary Claim a reward
// @Description Claim a reward by spending gold coins
// @Tags Rewards
// @Produce json
// @Security BearerAuth
// @Param id path int true "Reward ID"
// @Success 200 {object} dto.Response{data=application.RewardClaimResult}
// @Router /rewards/{id}/claim [post]
func (h *RewardHandler) ClaimReward(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := h.requireUserID(c)
	if !ok {
		return
	}

	id, ok := h.parseRewardID(c)
	if !ok {
		return
	}

	result, err := h.rewardService.ClaimReward(ctx, userID, id)
	if err != nil {
		status := http.StatusInternalServerError
		message := "Gagal mengklaim hadiah"

		switch err {
		case infrastructure.ErrInsufficientCoins:
			status = http.StatusBadRequest
			message = "Koin emas tidak cukup"
		case infrastructure.ErrRewardUnavailable:
			status = http.StatusBadRequest
			message = "Hadiah tidak tersedia"
		case infrastructure.ErrRewardOutOfStock:
			status = http.StatusBadRequest
			message = "Stok hadiah habis"
		case infrastructure.ErrRewardAlreadyOwned:
			status = http.StatusConflict
			message = "Anda sudah memiliki hadiah tema ini"
		case application.ErrRewardNotFound:
			status = http.StatusNotFound
			message = "Hadiah tidak ditemukan"
		}

		c.JSON(status, dto.Response{
			Success: false,
			Message: message,
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Message: "Berhasil mengklaim hadiah!",
		Data:    result,
	})
}

// GetMyClaims godoc
// @Summary Get my reward claims
// @Description Get paginated list of rewards claimed by the authenticated user
// @Tags Rewards
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} dto.Response{data=application.RewardClaimListResult}
// @Router /rewards/my-claims [get]
func (h *RewardHandler) GetMyClaims(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := h.requireUserID(c)
	if !ok {
		return
	}

	page, pageSize := h.parsePageAndSize(c)

	result, err := h.rewardService.GetUserClaims(ctx, userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Message: "Gagal mengambil riwayat klaim",
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Message: "Berhasil mengambil riwayat klaim",
		Data:    result,
	})
}

// GetCoinBalance godoc
// @Summary Get coin balance
// @Description Get the authenticated user's gold coin balance
// @Tags Rewards
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response{data=map[string]int64}
// @Router /rewards/balance [get]
func (h *RewardHandler) GetCoinBalance(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := h.requireUserID(c)
	if !ok {
		return
	}

	balance, err := h.rewardService.GetCoinBalance(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Message: "Gagal mengambil saldo koin",
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Message: "Berhasil mengambil saldo koin",
		Data: map[string]int64{
			"gold_coins": balance,
		},
	})
}

// ============ Admin Endpoints ============

// AdminGetAllRewards godoc
// @Summary Get all rewards (admin)
// @Description Get all rewards including inactive ones
// @Tags Admin Rewards
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response{data=[]model.Reward}
// @Router /admin/rewards [get]
func (h *RewardHandler) AdminGetAllRewards(c *gin.Context) {
	ctx := c.Request.Context()

	rewards, err := h.rewardService.GetAllRewards(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Message: "Gagal mengambil daftar hadiah",
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Message: "Berhasil mengambil daftar hadiah",
		Data:    rewards,
	})
}

// AdminCreateReward godoc
// @Summary Create a reward (admin)
// @Description Create a new reward that can be claimed with gold coins
// @Tags Admin Rewards
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param reward body object true "Reward data"
// @Success 201 {object} dto.Response{data=model.Reward}
// @Router /admin/rewards [post]
func (h *RewardHandler) AdminCreateReward(c *gin.Context) {
	ctx := c.Request.Context()

	var reward model.Reward
	if !h.bindJSON(c, &reward, "Data hadiah tidak valid") {
		return
	}

	if reward.Name == "" || reward.CoinCost <= 0 {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: "Nama dan harga koin harus diisi",
		})
		return
	}

	if err := h.rewardService.CreateReward(ctx, &reward); err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Message: "Gagal membuat hadiah",
		})
		return
	}

	c.JSON(http.StatusCreated, dto.Response{
		Success: true,
		Message: "Berhasil membuat hadiah",
		Data:    reward,
	})
}

// AdminUpdateReward godoc
// @Summary Update a reward (admin)
// @Description Update an existing reward
// @Tags Admin Rewards
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Reward ID"
// @Param reward body application.UpdateRewardInput true "Update data"
// @Success 200 {object} dto.Response{data=model.Reward}
// @Router /admin/rewards/{id} [put]
func (h *RewardHandler) AdminUpdateReward(c *gin.Context) {
	ctx := c.Request.Context()
	id, ok := h.parseRewardID(c)
	if !ok {
		return
	}

	var input application.UpdateRewardInput
	if !h.bindJSON(c, &input, "Data update tidak valid") {
		return
	}

	reward, err := h.rewardService.UpdateReward(ctx, id, input)
	if err != nil {
		if err == application.ErrRewardNotFound {
			c.JSON(http.StatusNotFound, dto.Response{
				Success: false,
				Message: "Hadiah tidak ditemukan",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Message: "Gagal mengupdate hadiah",
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Message: "Berhasil mengupdate hadiah",
		Data:    reward,
	})
}

// AdminDeleteReward godoc
// @Summary Delete a reward (admin)
// @Description Delete a reward
// @Tags Admin Rewards
// @Produce json
// @Security BearerAuth
// @Param id path int true "Reward ID"
// @Success 200 {object} dto.Response
// @Router /admin/rewards/{id} [delete]
func (h *RewardHandler) AdminDeleteReward(c *gin.Context) {
	ctx := c.Request.Context()
	id, ok := h.parseRewardID(c)
	if !ok {
		return
	}

	if err := h.rewardService.DeleteReward(ctx, id); err != nil {
		if err == application.ErrRewardNotFound {
			c.JSON(http.StatusNotFound, dto.Response{
				Success: false,
				Message: "Hadiah tidak ditemukan",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Message: "Gagal menghapus hadiah",
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Message: "Berhasil menghapus hadiah",
	})
}

// AdminGetAllClaims godoc
// @Summary Get all reward claims (admin)
// @Description Get paginated list of all reward claims with user info
// @Tags Admin Rewards
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} dto.Response{data=application.RewardClaimListResult}
// @Router /admin/rewards/claims [get]
func (h *RewardHandler) AdminGetAllClaims(c *gin.Context) {
	ctx := c.Request.Context()

	page, pageSize := h.parsePageAndSize(c)

	result, err := h.rewardService.GetAllClaims(ctx, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Message: "Gagal mengambil daftar klaim",
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Message: "Berhasil mengambil daftar klaim",
		Data:    result,
	})
}

// ============ Theme Endpoints ============

// GetOwnedThemes godoc
// @Summary Get owned themes
// @Description Get list of themes owned by the authenticated user
// @Tags Rewards
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response
// @Router /rewards/themes [get]
func (h *RewardHandler) GetOwnedThemes(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := h.requireUserID(c)
	if !ok {
		return
	}

	themes, activeTheme, err := h.rewardService.GetOwnedThemes(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Message: "Gagal mengambil daftar tema",
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Message: "Berhasil mengambil daftar tema",
		Data: map[string]interface{}{
			"owned_themes": themes,
			"active_theme": activeTheme,
		},
	})
}

// ActivateTheme godoc
// @Summary Activate a theme
// @Description Set a previously claimed theme as the active dashboard theme
// @Tags Rewards
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response
// @Router /rewards/themes/activate [put]
func (h *RewardHandler) ActivateTheme(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := h.requireUserID(c)
	if !ok {
		return
	}

	var req struct {
		Theme string `json:"theme" binding:"required"`
	}
	if !h.bindJSON(c, &req, "Tema tidak valid") {
		return
	}

	if err := h.rewardService.ActivateTheme(ctx, userID, req.Theme); err != nil {
		status := http.StatusInternalServerError
		message := "Gagal mengaktifkan tema"

		if err == application.ErrThemeNotOwned {
			status = http.StatusForbidden
			message = "Anda belum memiliki tema ini"
		}

		c.JSON(status, dto.Response{
			Success: false,
			Message: message,
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Message: "Tema berhasil diaktifkan!",
		Data: map[string]string{
			"active_theme": req.Theme,
		},
	})
}
