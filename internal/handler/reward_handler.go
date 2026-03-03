package handler

import (
	"net/http"
	"strconv"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/service"
	"github.com/gin-gonic/gin"
)

type RewardHandler struct {
	rewardService service.RewardService
}

func NewRewardHandler(rewardService service.RewardService) *RewardHandler {
	return &RewardHandler{
		rewardService: rewardService,
	}
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
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: "ID hadiah tidak valid",
		})
		return
	}

	reward, err := h.rewardService.GetRewardByID(ctx, uint(id))
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
// @Success 200 {object} dto.Response{data=service.RewardClaimResult}
// @Router /rewards/{id}/claim [post]
func (h *RewardHandler) ClaimReward(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetUint("user_id")
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: "ID hadiah tidak valid",
		})
		return
	}

	result, err := h.rewardService.ClaimReward(ctx, userID, uint(id))
	if err != nil {
		status := http.StatusInternalServerError
		message := "Gagal mengklaim hadiah"

		switch err {
		case service.ErrInsufficientCoins:
			status = http.StatusBadRequest
			message = "Koin emas tidak cukup"
		case service.ErrRewardUnavailable:
			status = http.StatusBadRequest
			message = "Hadiah tidak tersedia"
		case service.ErrRewardOutOfStock:
			status = http.StatusBadRequest
			message = "Stok hadiah habis"
		case service.ErrRewardNotFound:
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
// @Success 200 {object} dto.Response{data=service.RewardClaimListResult}
// @Router /rewards/my-claims [get]
func (h *RewardHandler) GetMyClaims(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetUint("user_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

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
	userID := c.GetUint("user_id")

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
	if err := c.ShouldBindJSON(&reward); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: "Data hadiah tidak valid",
		})
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
// @Param reward body service.UpdateRewardInput true "Update data"
// @Success 200 {object} dto.Response{data=model.Reward}
// @Router /admin/rewards/{id} [put]
func (h *RewardHandler) AdminUpdateReward(c *gin.Context) {
	ctx := c.Request.Context()
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: "ID hadiah tidak valid",
		})
		return
	}

	var input service.UpdateRewardInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: "Data update tidak valid",
		})
		return
	}

	reward, err := h.rewardService.UpdateReward(ctx, uint(id), input)
	if err != nil {
		if err == service.ErrRewardNotFound {
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
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: "ID hadiah tidak valid",
		})
		return
	}

	if err := h.rewardService.DeleteReward(ctx, uint(id)); err != nil {
		if err == service.ErrRewardNotFound {
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
// @Success 200 {object} dto.Response{data=service.RewardClaimListResult}
// @Router /admin/rewards/claims [get]
func (h *RewardHandler) AdminGetAllClaims(c *gin.Context) {
	ctx := c.Request.Context()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

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
