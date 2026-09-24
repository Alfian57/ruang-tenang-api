package handler

import (
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	articleinfra "github.com/Alfian57/ruang-tenang-api/internal/features/article/infrastructure"
	songinfra "github.com/Alfian57/ruang-tenang-api/internal/features/song/infrastructure"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	articleRepo *articleinfra.ArticleRepository
	songRepo    *songinfra.SongRepository
}

func NewSearchHandler(articleRepo *articleinfra.ArticleRepository, songRepo *songinfra.SongRepository) *SearchHandler {
	return &SearchHandler{
		articleRepo: articleRepo,
		songRepo:    songRepo,
	}
}

// GlobalSearch godoc
// @Summary Global search
// @Description Search for published articles and songs for regular users
// @Tags Search
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param q query string true "Search query"
// @Param type query string false "Use songs for paginated song-only search"
// @Param page query int false "Song search page"
// @Param limit query int false "Song results per page"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Router /search [get]
func (h *SearchHandler) Search(c *gin.Context) {
	ctx := c.Request.Context()
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
			"articles": []model.Article{},
			"songs":    []model.Song{},
			"total":    0,
		}, "Query empty"))
		return
	}
	if c.Query("type") == "songs" {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "12"))
		if page < 1 {
			page = 1
		}
		if limit < 1 || limit > 50 {
			limit = 12
		}
		songs, total, err := h.songRepo.SearchPage(ctx, query, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to search songs"))
			return
		}
		c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"articles": []model.Article{}, "songs": songs, "total": total, "page": page, "limit": limit, "total_pages": (total + int64(limit) - 1) / int64(limit)}, "Search successful"))
		return
	}

	var wg sync.WaitGroup
	var articles []model.Article
	var songs []model.Song
	var articleErr, songErr error

	wg.Add(2)

	// Search Articles
	go func() {
		defer wg.Done()
		// Search published articles only, limited to 5
		articles, _, articleErr = h.articleRepo.FindPublished(ctx, 0, query, 1, 5)
	}()

	// Search Songs
	go func() {
		defer wg.Done()
		songs, songErr = h.songRepo.Search(ctx, query)
	}()

	wg.Wait()

	if articleErr != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to search articles"))
		return
	}

	if songErr != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to search songs"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"articles": articles,
		"songs":    songs,
		"total":    len(articles) + len(songs),
	}, "Search successful"))
}
