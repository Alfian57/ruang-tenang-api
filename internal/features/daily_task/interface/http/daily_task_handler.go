package handler

import (
	"net/http"
	"strconv"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/gin-gonic/gin"

	"github.com/Alfian57/ruang-tenang-api/internal/features/daily_task/application")

type DailyTaskHandler struct {
	dailyTaskService application.DailyTaskService
}

func NewDailyTaskHandler(dailyTaskService application.DailyTaskService) *DailyTaskHandler {
	return &DailyTaskHandler{
		dailyTaskService: dailyTaskService,
	}
}

// GetDailyTasks godoc
// @Summary Get today's daily tasks
// @Description Get all daily tasks for the authenticated user for today
// @Tags Daily Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response{data=model.DailyTaskSummary}
// @Failure 401 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /daily-tasks [get]
func (h *DailyTaskHandler) GetDailyTasks(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetUint("user_id")

	summary, err := h.dailyTaskService.GetTodayTasks(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Message: "Gagal mengambil daily tasks",
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Message: "Berhasil mengambil daily tasks",
		Data:    summary,
	})
}

// ClaimDailyLogin godoc
// @Summary Claim daily login reward
// @Description Process daily login and claim login task reward
// @Tags Daily Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response{data=DailyLoginResult}
// @Failure 401 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /daily-tasks/login [post]
func (h *DailyTaskHandler) ClaimDailyLogin(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetUint("user_id")

	result, err := h.dailyTaskService.ProcessDailyLogin(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Message: "Gagal memproses daily login",
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Message: result.Message,
		Data:    result,
	})
}

// ClaimTaskReward godoc
// @Summary Claim reward for a completed task
// @Description Claim XP reward for a completed daily task
// @Tags Daily Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Task ID"
// @Success 200 {object} dto.Response{data=ClaimResult}
// @Failure 400 {object} dto.Response
// @Failure 401 {object} dto.Response
// @Failure 404 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /daily-tasks/{id}/claim [post]
func (h *DailyTaskHandler) ClaimTaskReward(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetUint("user_id")
	taskIDStr := c.Param("id")

	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: "ID task tidak valid",
		})
		return
	}

	result, err := h.dailyTaskService.ClaimTaskReward(ctx, userID, uint(taskID))
	if err != nil {
		switch err {
		case application.ErrTaskNotFound:
			c.JSON(http.StatusNotFound, dto.Response{
				Success: false,
				Message: "Task tidak ditemukan",
			})
		case application.ErrTaskNotCompleted:
			c.JSON(http.StatusBadRequest, dto.Response{
				Success: false,
				Message: "Task belum selesai",
			})
		case application.ErrTaskAlreadyClaimed:
			c.JSON(http.StatusBadRequest, dto.Response{
				Success: false,
				Message: "Reward sudah diklaim",
			})
		default:
			c.JSON(http.StatusInternalServerError, dto.Response{
				Success: false,
				Message: "Gagal mengklaim reward",
			})
		}
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Message: "Berhasil mengklaim reward: +" + strconv.Itoa(result.XPEarned) + " XP",
		Data:    result,
	})
}

// ClaimAllRewards godoc
// @Summary Claim all completed task rewards
// @Description Claim XP rewards for all completed but unclaimed daily tasks
// @Tags Daily Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response{data=ClaimAllResult}
// @Failure 401 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /daily-tasks/claim-all [post]
func (h *DailyTaskHandler) ClaimAllRewards(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetUint("user_id")

	result, err := h.dailyTaskService.ClaimAllRewards(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Message: "Gagal mengklaim rewards",
		})
		return
	}

	message := "Tidak ada reward yang bisa diklaim"
	if result.TotalClaimed > 0 {
		message = "Berhasil mengklaim " + strconv.Itoa(result.TotalClaimed) + " rewards: +" + strconv.Itoa(result.TotalXPEarned) + " XP"
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Message: message,
		Data:    result,
	})
}

// GetTaskHistory godoc
// @Summary Get daily task history
// @Description Get historical daily task summaries for the authenticated user
// @Tags Daily Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(7)
// @Success 200 {object} dto.Response{data=TaskHistoryResult}
// @Failure 401 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /daily-tasks/history [get]
func (h *DailyTaskHandler) GetTaskHistory(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetUint("user_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "7"))

	result, err := h.dailyTaskService.GetTaskHistory(ctx, userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Message: "Gagal mengambil history daily tasks",
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Message: "Berhasil mengambil history daily tasks",
		Data:    result,
	})
}
