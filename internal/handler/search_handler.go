package handler

import (
	"net/http"
	"sync"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	articleRepo *repository.ArticleRepository
	songRepo    *repository.SongRepository
}

func NewSearchHandler(articleRepo *repository.ArticleRepository, songRepo *repository.SongRepository) *SearchHandler {
	return &SearchHandler{
		articleRepo: articleRepo,
		songRepo:    songRepo,
	}
}

// GlobalSearch godoc
// @Summary Global search
// @Description Search for articles and songs
// @Tags Search
// @Accept json
// @Produce json
// @Query q query string true "Search query"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Router /search [get]
func (h *SearchHandler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
			"articles": []model.Article{},
			"songs":    []model.Song{},
		}, "Query empty"))
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
		articles, _, articleErr = h.articleRepo.FindPublished(0, query, 1, 5)
	}()

	// Search Songs
	go func() {
		defer wg.Done()
		songs, songErr = h.songRepo.Search(query)
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
	}, "Search successful"))
}
