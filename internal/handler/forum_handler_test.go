package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/service"
	"github.com/gin-gonic/gin"
)

type mockForumServiceForHandler struct {
	createForumFn             func(ctx context.Context, userID uint, title, content string, categoryID *uint) (*model.Forum, error)
	deleteForumBySlugFn       func(ctx context.Context, userID uint, userRole string, slug string) error
	createForumPostBySlugFn   func(ctx context.Context, userID uint, forumSlug string, content string) error
	getForumsFn               func(ctx context.Context, limit, offset int, search string, categoryID *uint) ([]model.Forum, int64, error)
	getForumBySlugFn          func(ctx context.Context, userID uint, slug string) (*model.Forum, error)
	getForumPostsSortedByFn   func(ctx context.Context, forumSlug string, limit, offset int, sort string, userID uint) ([]model.ForumPost, int64, error)
	deleteForumPostFn         func(ctx context.Context, userID uint, userRole string, postID uint) error
	toggleLikeBySlugFn        func(ctx context.Context, userID uint, forumSlug string) (bool, error)
	votePostFn                func(ctx context.Context, userID, postID uint, voteType string) error
	removePostVoteFn          func(ctx context.Context, userID, postID uint) error
	markAsAcceptedAnswerFn    func(ctx context.Context, userID uint, userRole string, postID uint) error
	unmarkAcceptedAnswerFn    func(ctx context.Context, userID uint, userRole string, forumID uint) error
	unmarkAcceptedBySlugFn    func(ctx context.Context, userID uint, userRole string, forumSlug string) error
	getAcceptedAnswerBySlugFn func(ctx context.Context, forumSlug string) (*model.ForumPost, error)
	reportPostFn              func(ctx context.Context, userID, postID uint, reason, description string) error
	getPendingReportsFn       func(ctx context.Context, limit, offset int) ([]model.ForumPostReport, int64, error)
	reviewPostReportFn        func(ctx context.Context, reviewerID, reportID uint, status string, notes string) error

	lastCreatePostUserID uint
	lastCreatePostSlug   string
	lastCreatePostBody   string
}

func (m *mockForumServiceForHandler) CreateForum(ctx context.Context, userID uint, title, content string, categoryID *uint) (*model.Forum, error) {
	if m.createForumFn != nil {
		return m.createForumFn(ctx, userID, title, content, categoryID)
	}
	return &model.Forum{ID: 1, UserID: userID, Title: title, Content: content}, nil
}
func (m *mockForumServiceForHandler) GetForums(ctx context.Context, limit, offset int, search string, categoryID *uint) ([]model.Forum, int64, error) {
	if m.getForumsFn != nil {
		return m.getForumsFn(ctx, limit, offset, search, categoryID)
	}
	return []model.Forum{}, 0, nil
}
func (m *mockForumServiceForHandler) GetForumByID(_ context.Context, _, id uint) (*model.Forum, error) {
	return &model.Forum{ID: id}, nil
}
func (m *mockForumServiceForHandler) GetForumBySlug(ctx context.Context, userID uint, slug string) (*model.Forum, error) {
	if m.getForumBySlugFn != nil {
		return m.getForumBySlugFn(ctx, userID, slug)
	}
	return &model.Forum{ID: 1, Title: slug}, nil
}
func (m *mockForumServiceForHandler) DeleteForum(_ context.Context, _ uint, _ string, _ uint) error {
	return nil
}
func (m *mockForumServiceForHandler) DeleteForumBySlug(ctx context.Context, userID uint, userRole string, slug string) error {
	if m.deleteForumBySlugFn != nil {
		return m.deleteForumBySlugFn(ctx, userID, userRole, slug)
	}
	return nil
}
func (m *mockForumServiceForHandler) CreateForumPost(_ context.Context, _ uint, _ uint, _ string) error {
	return nil
}
func (m *mockForumServiceForHandler) CreateForumPostBySlug(ctx context.Context, userID uint, forumSlug string, content string) error {
	m.lastCreatePostUserID = userID
	m.lastCreatePostSlug = forumSlug
	m.lastCreatePostBody = content
	if m.createForumPostBySlugFn != nil {
		return m.createForumPostBySlugFn(ctx, userID, forumSlug, content)
	}
	return nil
}
func (m *mockForumServiceForHandler) GetForumPosts(_ context.Context, _ uint, _, _ int) ([]model.ForumPost, int64, error) {
	return nil, 0, nil
}
func (m *mockForumServiceForHandler) GetForumPostsSorted(_ context.Context, _ uint, _, _ int, _ string, _ uint) ([]model.ForumPost, int64, error) {
	return nil, 0, nil
}
func (m *mockForumServiceForHandler) GetForumPostsSortedBySlug(ctx context.Context, forumSlug string, limit, offset int, sort string, userID uint) ([]model.ForumPost, int64, error) {
	if m.getForumPostsSortedByFn != nil {
		return m.getForumPostsSortedByFn(ctx, forumSlug, limit, offset, sort, userID)
	}
	return []model.ForumPost{}, 0, nil
}
func (m *mockForumServiceForHandler) DeleteForumPost(ctx context.Context, userID uint, userRole string, postID uint) error {
	if m.deleteForumPostFn != nil {
		return m.deleteForumPostFn(ctx, userID, userRole, postID)
	}
	return nil
}
func (m *mockForumServiceForHandler) ToggleLike(_ context.Context, _, _ uint) (bool, error) {
	return false, nil
}
func (m *mockForumServiceForHandler) ToggleLikeBySlug(ctx context.Context, userID uint, forumSlug string) (bool, error) {
	if m.toggleLikeBySlugFn != nil {
		return m.toggleLikeBySlugFn(ctx, userID, forumSlug)
	}
	return true, nil
}
func (m *mockForumServiceForHandler) GetForumStats(_ context.Context, _ uint) (int64, error) {
	return 0, nil
}
func (m *mockForumServiceForHandler) VotePost(ctx context.Context, userID, postID uint, voteType string) error {
	if m.votePostFn != nil {
		return m.votePostFn(ctx, userID, postID, voteType)
	}
	return nil
}
func (m *mockForumServiceForHandler) RemovePostVote(ctx context.Context, userID, postID uint) error {
	if m.removePostVoteFn != nil {
		return m.removePostVoteFn(ctx, userID, postID)
	}
	return nil
}
func (m *mockForumServiceForHandler) GetPostVoteStatus(_ context.Context, _, _ uint) (*model.ForumPostVote, error) {
	return nil, nil
}
func (m *mockForumServiceForHandler) MarkAsAcceptedAnswer(ctx context.Context, userID uint, userRole string, postID uint) error {
	if m.markAsAcceptedAnswerFn != nil {
		return m.markAsAcceptedAnswerFn(ctx, userID, userRole, postID)
	}
	return nil
}
func (m *mockForumServiceForHandler) UnmarkAcceptedAnswer(ctx context.Context, userID uint, userRole string, forumID uint) error {
	if m.unmarkAcceptedAnswerFn != nil {
		return m.unmarkAcceptedAnswerFn(ctx, userID, userRole, forumID)
	}
	return nil
}
func (m *mockForumServiceForHandler) UnmarkAcceptedAnswerBySlug(ctx context.Context, userID uint, userRole string, forumSlug string) error {
	if m.unmarkAcceptedBySlugFn != nil {
		return m.unmarkAcceptedBySlugFn(ctx, userID, userRole, forumSlug)
	}
	return nil
}
func (m *mockForumServiceForHandler) GetAcceptedAnswer(_ context.Context, _ uint) (*model.ForumPost, error) {
	return nil, nil
}
func (m *mockForumServiceForHandler) GetAcceptedAnswerBySlug(ctx context.Context, forumSlug string) (*model.ForumPost, error) {
	if m.getAcceptedAnswerBySlugFn != nil {
		return m.getAcceptedAnswerBySlugFn(ctx, forumSlug)
	}
	return nil, nil
}
func (m *mockForumServiceForHandler) ReportPost(ctx context.Context, userID, postID uint, reason, description string) error {
	if m.reportPostFn != nil {
		return m.reportPostFn(ctx, userID, postID, reason, description)
	}
	return nil
}
func (m *mockForumServiceForHandler) GetPendingPostReports(ctx context.Context, limit, offset int) ([]model.ForumPostReport, int64, error) {
	if m.getPendingReportsFn != nil {
		return m.getPendingReportsFn(ctx, limit, offset)
	}
	return nil, 0, nil
}
func (m *mockForumServiceForHandler) ReviewPostReport(ctx context.Context, reviewerID, reportID uint, status string, notes string) error {
	if m.reviewPostReportFn != nil {
		return m.reviewPostReportFn(ctx, reviewerID, reportID, status, notes)
	}
	return nil
}

type mockDailyTaskProgressForForumHandler struct {
	progressCalls int
	lastUserID    uint
	lastTaskType  model.DailyTaskType
}

func (m *mockDailyTaskProgressForForumHandler) InitializeDailyTasks(_ context.Context, _ uint) error {
	return nil
}
func (m *mockDailyTaskProgressForForumHandler) ProcessDailyLogin(_ context.Context, _ uint) (*service.DailyLoginResult, error) {
	return nil, nil
}
func (m *mockDailyTaskProgressForForumHandler) UpdateTaskProgress(_ context.Context, userID uint, taskType model.DailyTaskType) error {
	m.progressCalls++
	m.lastUserID = userID
	m.lastTaskType = taskType
	return nil
}
func (m *mockDailyTaskProgressForForumHandler) GetTodayTasks(_ context.Context, _ uint) (*model.DailyTaskSummary, error) {
	return nil, nil
}
func (m *mockDailyTaskProgressForForumHandler) ClaimTaskReward(_ context.Context, _ uint, _ uint) (*service.ClaimResult, error) {
	return nil, nil
}
func (m *mockDailyTaskProgressForForumHandler) ClaimAllRewards(_ context.Context, _ uint) (*service.ClaimAllResult, error) {
	return nil, nil
}
func (m *mockDailyTaskProgressForForumHandler) GetTaskHistory(_ context.Context, _ uint, _, _ int) (*service.TaskHistoryResult, error) {
	return nil, nil
}

func TestForumHandler_CreateForumAndDeleteForum(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("create forum success", func(t *testing.T) {
		svc := &mockForumServiceForHandler{}
		h := NewForumHandler(svc)
		r := gin.New()
		r.POST("/forums", func(c *gin.Context) {
			c.Set("user_id", uint(11))
			h.CreateForum(c)
		})

		body := bytes.NewBufferString(`{"title":"Topik Baru","content":"Isi"}`)
		req := httptest.NewRequest(http.MethodPost, "/forums", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", w.Code)
		}
	})

	t.Run("delete forum unauthorized", func(t *testing.T) {
		svc := &mockForumServiceForHandler{
			deleteForumBySlugFn: func(_ context.Context, _ uint, _ string, _ string) error {
				return errors.New("unauthorized")
			},
		}
		h := NewForumHandler(svc)
		r := gin.New()
		r.DELETE("/forums/:slug", func(c *gin.Context) {
			c.Set("user_id", uint(11))
			c.Set("user_role", "member")
			h.DeleteForum(c)
		})

		req := httptest.NewRequest(http.MethodDelete, "/forums/abc", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w.Code)
		}
	})
}

func TestForumHandler_CreateForumPostUpdatesDailyTask(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockForumServiceForHandler{}
	h := NewForumHandler(svc)

	dailyTask := &mockDailyTaskProgressForForumHandler{}
	h.SetDailyTaskService(dailyTask)

	r := gin.New()
	r.POST("/forums/:slug", func(c *gin.Context) {
		c.Set("user_id", uint(20))
		h.CreateForumPost(c)
	})

	body := bytes.NewBufferString(`{"content":"Halo forum"}`)
	req := httptest.NewRequest(http.MethodPost, "/forums/general", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	if svc.lastCreatePostUserID != 20 || svc.lastCreatePostSlug != "general" || svc.lastCreatePostBody != "Halo forum" {
		t.Fatalf("unexpected forum post call args: user=%d slug=%s body=%s", svc.lastCreatePostUserID, svc.lastCreatePostSlug, svc.lastCreatePostBody)
	}
	if dailyTask.progressCalls != 1 || dailyTask.lastUserID != 20 || dailyTask.lastTaskType != model.TaskTypeCommentForum {
		t.Fatalf("unexpected daily task update state: calls=%d user=%d task=%s", dailyTask.progressCalls, dailyTask.lastUserID, dailyTask.lastTaskType)
	}

	t.Run("create forum post invalid payload and internal error", func(t *testing.T) {
		svc2 := &mockForumServiceForHandler{
			createForumPostBySlugFn: func(_ context.Context, _ uint, _ string, _ string) error {
				return errors.New("boom")
			},
		}
		h2 := NewForumHandler(svc2)
		r2 := gin.New()
		r2.POST("/forums/:slug", func(c *gin.Context) {
			c.Set("user_id", uint(21))
			h2.CreateForumPost(c)
		})

		wBad := httptest.NewRecorder()
		reqBad := httptest.NewRequest(http.MethodPost, "/forums/general", bytes.NewBufferString("{"))
		reqBad.Header.Set("Content-Type", "application/json")
		r2.ServeHTTP(wBad, reqBad)
		if wBad.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", wBad.Code)
		}

		wErr := httptest.NewRecorder()
		reqErr := httptest.NewRequest(http.MethodPost, "/forums/general", bytes.NewBufferString(`{"content":"x"}`))
		reqErr.Header.Set("Content-Type", "application/json")
		r2.ServeHTTP(wErr, reqErr)
		if wErr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", wErr.Code)
		}
	})
}

func TestForumHandler_GetForumsAndToggleLike(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("GetForums success with query params", func(t *testing.T) {
		svc := &mockForumServiceForHandler{
			getForumsFn: func(_ context.Context, limit, offset int, search string, categoryID *uint) ([]model.Forum, int64, error) {
				if limit != 5 || offset != 10 || search != "abc" {
					t.Fatalf("unexpected query mapping: limit=%d offset=%d search=%s", limit, offset, search)
				}
				if categoryID == nil || *categoryID != 2 {
					t.Fatalf("expected category id 2, got %v", categoryID)
				}
				return []model.Forum{{ID: 1, Title: "A"}}, 1, nil
			},
		}
		h := NewForumHandler(svc)
		r := gin.New()
		r.GET("/forums", h.GetForums)

		req := httptest.NewRequest(http.MethodGet, "/forums?limit=5&offset=10&search=abc&category_id=2", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("ToggleLike error", func(t *testing.T) {
		svc := &mockForumServiceForHandler{
			toggleLikeBySlugFn: func(_ context.Context, _ uint, _ string) (bool, error) {
				return false, errors.New("boom")
			},
		}
		h := NewForumHandler(svc)
		r := gin.New()
		r.PUT("/forums/:slug/like", func(c *gin.Context) {
			c.Set("user_id", uint(10))
			h.ToggleLike(c)
		})

		req := httptest.NewRequest(http.MethodPut, "/forums/general/like", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("ToggleLike unliked branch", func(t *testing.T) {
		svc := &mockForumServiceForHandler{
			toggleLikeBySlugFn: func(_ context.Context, _ uint, _ string) (bool, error) { return false, nil },
		}
		h := NewForumHandler(svc)
		r := gin.New()
		r.PUT("/forums/:slug/like", func(c *gin.Context) {
			c.Set("user_id", uint(10))
			h.ToggleLike(c)
		})

		req := httptest.NewRequest(http.MethodPut, "/forums/general/like", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})
}

func TestForumHandler_ForumAndPostEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("get forum by slug not found", func(t *testing.T) {
		svc := &mockForumServiceForHandler{
			getForumBySlugFn: func(_ context.Context, _ uint, _ string) (*model.Forum, error) {
				return nil, errors.New("not found")
			},
		}
		h := NewForumHandler(svc)
		r := gin.New()
		r.GET("/forums/:slug", func(c *gin.Context) {
			c.Set("user_id", uint(10))
			h.GetForumByID(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/forums/slug-1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})

	t.Run("get forum posts success and error", func(t *testing.T) {
		svc := &mockForumServiceForHandler{
			getForumPostsSortedByFn: func(_ context.Context, _ string, _, _ int, _ string, _ uint) ([]model.ForumPost, int64, error) {
				return []model.ForumPost{{ID: 1}}, 1, nil
			},
		}
		h := NewForumHandler(svc)
		r := gin.New()
		r.GET("/forums/:slug/posts", func(c *gin.Context) {
			c.Set("user_id", uint(10))
			h.GetForumPosts(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/forums/slug-1/posts?limit=5&offset=0&sort=newest", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}

		svc.getForumPostsSortedByFn = func(_ context.Context, _ string, _, _ int, _ string, _ uint) ([]model.ForumPost, int64, error) {
			return nil, 0, errors.New("boom")
		}
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, req)
		if w2.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w2.Code)
		}
	})

	t.Run("delete post unauthorized", func(t *testing.T) {
		svc := &mockForumServiceForHandler{
			deleteForumPostFn: func(_ context.Context, _ uint, _ string, _ uint) error {
				return errors.New("unauthorized")
			},
		}
		h := NewForumHandler(svc)
		r := gin.New()
		r.DELETE("/posts/:id", func(c *gin.Context) {
			c.Set("user_id", uint(8))
			c.Set("user_role", "member")
			h.DeleteForumPost(c)
		})

		req := httptest.NewRequest(http.MethodDelete, "/posts/11", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w.Code)
		}
	})
}

func TestForumHandler_VoteAcceptReportAndUtilityEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("create forum invalid and internal error", func(t *testing.T) {
		svc := &mockForumServiceForHandler{
			createForumFn: func(_ context.Context, _ uint, _, _ string, _ *uint) (*model.Forum, error) {
				return nil, errors.New("create failed")
			},
		}
		h := NewForumHandler(svc)
		r := gin.New()
		r.POST("/forums", func(c *gin.Context) {
			c.Set("user_id", uint(9))
			h.CreateForum(c)
		})

		w1 := httptest.NewRecorder()
		req1 := httptest.NewRequest(http.MethodPost, "/forums", bytes.NewBufferString("{"))
		req1.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w1, req1)
		if w1.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w1.Code)
		}

		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodPost, "/forums", bytes.NewBufferString(`{"title":"Topik","content":"Isi"}`))
		req2.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w2, req2)
		if w2.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w2.Code)
		}
	})

	t.Run("vote endpoints", func(t *testing.T) {
		svc := &mockForumServiceForHandler{}
		h := NewForumHandler(svc)
		r := gin.New()
		r.PUT("/posts/:id/upvote", func(c *gin.Context) { c.Set("user_id", uint(1)); h.UpvotePost(c) })
		r.PUT("/posts/:id/downvote", func(c *gin.Context) { c.Set("user_id", uint(1)); h.DownvotePost(c) })
		r.DELETE("/posts/:id/vote", func(c *gin.Context) { c.Set("user_id", uint(1)); h.RemovePostVote(c) })

		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, httptest.NewRequest(http.MethodPut, "/posts/2/upvote", nil))
		if w1.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w1.Code)
		}

		svc.votePostFn = func(_ context.Context, _, _ uint, _ string) error { return errors.New("cannot vote on your own post") }
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, httptest.NewRequest(http.MethodPut, "/posts/2/downvote", nil))
		if w2.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w2.Code)
		}

		w2b := httptest.NewRecorder()
		r.ServeHTTP(w2b, httptest.NewRequest(http.MethodPut, "/posts/2/upvote", nil))
		if w2b.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 upvote own-post, got %d", w2b.Code)
		}

		svc.removePostVoteFn = func(_ context.Context, _, _ uint) error { return errors.New("no vote") }
		w3 := httptest.NewRecorder()
		r.ServeHTTP(w3, httptest.NewRequest(http.MethodDelete, "/posts/2/vote", nil))
		if w3.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w3.Code)
		}

		svc.votePostFn = func(_ context.Context, _, _ uint, _ string) error { return errors.New("db down") }
		w4 := httptest.NewRecorder()
		r.ServeHTTP(w4, httptest.NewRequest(http.MethodPut, "/posts/2/upvote", nil))
		if w4.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w4.Code)
		}
	})

	t.Run("accepted answer endpoints", func(t *testing.T) {
		svc := &mockForumServiceForHandler{}
		h := NewForumHandler(svc)
		r := gin.New()
		r.PUT("/posts/:id/accept", func(c *gin.Context) { c.Set("user_id", uint(1)); c.Set("user_role", "member"); h.MarkAcceptedAnswer(c) })
		r.DELETE("/forums/:slug/accepted-answer", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			c.Set("user_role", "member")
			h.UnmarkAcceptedAnswer(c)
		})
		r.GET("/forums/:slug/accepted-answer", h.GetAcceptedAnswer)

		svc.markAsAcceptedAnswerFn = func(_ context.Context, _ uint, _ string, _ uint) error {
			return errors.New("only the thread creator can mark an accepted answer")
		}
		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, httptest.NewRequest(http.MethodPut, "/posts/3/accept", nil))
		if w1.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w1.Code)
		}

		svc.unmarkAcceptedBySlugFn = func(_ context.Context, _ uint, _ string, _ string) error {
			return errors.New("only the thread creator can unmark an accepted answer")
		}
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, httptest.NewRequest(http.MethodDelete, "/forums/slug-1/accepted-answer", nil))
		if w2.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w2.Code)
		}

		svc.getAcceptedAnswerBySlugFn = func(_ context.Context, _ string) (*model.ForumPost, error) { return nil, nil }
		w3 := httptest.NewRecorder()
		r.ServeHTTP(w3, httptest.NewRequest(http.MethodGet, "/forums/slug-1/accepted-answer", nil))
		if w3.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w3.Code)
		}

		svc.getAcceptedAnswerBySlugFn = func(_ context.Context, _ string) (*model.ForumPost, error) { return nil, errors.New("boom") }
		w4 := httptest.NewRecorder()
		r.ServeHTTP(w4, httptest.NewRequest(http.MethodGet, "/forums/slug-1/accepted-answer", nil))
		if w4.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w4.Code)
		}

		svc.getAcceptedAnswerBySlugFn = func(_ context.Context, _ string) (*model.ForumPost, error) { return &model.ForumPost{ID: 88}, nil }
		w5 := httptest.NewRecorder()
		r.ServeHTTP(w5, httptest.NewRequest(http.MethodGet, "/forums/slug-1/accepted-answer", nil))
		if w5.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w5.Code)
		}
	})

	t.Run("downvote internal error branch", func(t *testing.T) {
		svc := &mockForumServiceForHandler{
			votePostFn: func(_ context.Context, _, _ uint, voteType string) error {
				if voteType != "downvote" {
					t.Fatalf("unexpected vote type %s", voteType)
				}
				return errors.New("db fail")
			},
		}
		h := NewForumHandler(svc)
		r := gin.New()
		r.PUT("/posts/:id/downvote", func(c *gin.Context) { c.Set("user_id", uint(1)); h.DownvotePost(c) })

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/posts/2/downvote", nil))
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("report and moderation endpoints", func(t *testing.T) {
		svc := &mockForumServiceForHandler{}
		h := NewForumHandler(svc)
		r := gin.New()
		r.POST("/posts/:id/report", func(c *gin.Context) { c.Set("user_id", uint(2)); h.ReportPost(c) })
		r.GET("/moderation/post-reports", h.GetPendingPostReports)
		r.PUT("/moderation/post-reports/:id", func(c *gin.Context) { c.Set("user_id", uint(3)); h.ReviewPostReport(c) })

		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, httptest.NewRequest(http.MethodPost, "/posts/5/report", bytes.NewBufferString("{")))
		if w1.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w1.Code)
		}

		svc.getPendingReportsFn = func(_ context.Context, _, _ int) ([]model.ForumPostReport, int64, error) {
			return nil, 0, errors.New("boom")
		}
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/moderation/post-reports", nil))
		if w2.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w2.Code)
		}

		w3 := httptest.NewRecorder()
		req3 := httptest.NewRequest(http.MethodPut, "/moderation/post-reports/7", bytes.NewBufferString("{"))
		req3.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w3, req3)
		if w3.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w3.Code)
		}

		svc.reportPostFn = func(_ context.Context, _ uint, _ uint, _, _ string) error { return errors.New("invalid report reason") }
		w4 := httptest.NewRecorder()
		req4 := httptest.NewRequest(http.MethodPost, "/posts/5/report", bytes.NewBufferString(`{"reason":"spam","description":"x"}`))
		req4.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w4, req4)
		if w4.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 invalid report reason, got %d", w4.Code)
		}

		svc.reportPostFn = func(_ context.Context, _ uint, _ uint, _, _ string) error { return errors.New("boom") }
		w5 := httptest.NewRecorder()
		req5 := httptest.NewRequest(http.MethodPost, "/posts/5/report", bytes.NewBufferString(`{"reason":"spam","description":"x"}`))
		req5.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w5, req5)
		if w5.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 report internal, got %d", w5.Code)
		}

		svc.reportPostFn = nil
		w6 := httptest.NewRecorder()
		req6 := httptest.NewRequest(http.MethodPost, "/posts/5/report", bytes.NewBufferString(`{"reason":"spam","description":"x"}`))
		req6.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w6, req6)
		if w6.Code != http.StatusCreated {
			t.Fatalf("expected 201 report success, got %d", w6.Code)
		}

		svc.reviewPostReportFn = func(_ context.Context, _, _ uint, _ string, _ string) error {
			return errors.New("invalid report status")
		}
		w7 := httptest.NewRecorder()
		req7 := httptest.NewRequest(http.MethodPut, "/moderation/post-reports/7", bytes.NewBufferString(`{"status":"bad","notes":"x"}`))
		req7.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w7, req7)
		if w7.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 invalid status, got %d", w7.Code)
		}

		svc.reviewPostReportFn = func(_ context.Context, _, _ uint, _ string, _ string) error { return errors.New("boom") }
		w8 := httptest.NewRecorder()
		req8 := httptest.NewRequest(http.MethodPut, "/moderation/post-reports/7", bytes.NewBufferString(`{"status":"reviewed","notes":"x"}`))
		req8.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w8, req8)
		if w8.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 review internal, got %d", w8.Code)
		}

		svc.reviewPostReportFn = nil
		w9 := httptest.NewRecorder()
		req9 := httptest.NewRequest(http.MethodPut, "/moderation/post-reports/7", bytes.NewBufferString(`{"status":"reviewed","notes":"ok"}`))
		req9.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w9, req9)
		if w9.Code != http.StatusOK {
			t.Fatalf("expected 200 review success, got %d", w9.Code)
		}
	})

	t.Run("static option endpoints", func(t *testing.T) {
		h := NewForumHandler(&mockForumServiceForHandler{})
		r := gin.New()
		r.GET("/forums/report-reasons", h.GetReportReasons)
		r.GET("/forums/sort-options", h.GetSortOptions)

		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, httptest.NewRequest(http.MethodGet, "/forums/report-reasons", nil))
		if w1.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w1.Code)
		}

		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/forums/sort-options", nil))
		if w2.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w2.Code)
		}
	})
}

func TestForumHandler_DeleteAndAcceptedAnswer_AdditionalBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("delete forum success and internal error", func(t *testing.T) {
		svc := &mockForumServiceForHandler{}
		h := NewForumHandler(svc)
		r := gin.New()
		r.DELETE("/forums/:slug", func(c *gin.Context) {
			c.Set("user_id", uint(11))
			c.Set("user_role", "member")
			h.DeleteForum(c)
		})

		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, httptest.NewRequest(http.MethodDelete, "/forums/topic-1", nil))
		if w1.Code != http.StatusOK {
			t.Fatalf("expected 200 delete forum success, got %d", w1.Code)
		}

		svc.deleteForumBySlugFn = func(_ context.Context, _ uint, _ string, _ string) error { return errors.New("db fail") }
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, httptest.NewRequest(http.MethodDelete, "/forums/topic-1", nil))
		if w2.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 delete forum internal, got %d", w2.Code)
		}
	})

	t.Run("delete post success and internal error", func(t *testing.T) {
		svc := &mockForumServiceForHandler{}
		h := NewForumHandler(svc)
		r := gin.New()
		r.DELETE("/posts/:id", func(c *gin.Context) {
			c.Set("user_id", uint(8))
			c.Set("user_role", "member")
			h.DeleteForumPost(c)
		})

		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, httptest.NewRequest(http.MethodDelete, "/posts/11", nil))
		if w1.Code != http.StatusOK {
			t.Fatalf("expected 200 delete post success, got %d", w1.Code)
		}

		svc.deleteForumPostFn = func(_ context.Context, _ uint, _ string, _ uint) error { return errors.New("db fail") }
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, httptest.NewRequest(http.MethodDelete, "/posts/11", nil))
		if w2.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 delete post internal, got %d", w2.Code)
		}
	})

	t.Run("accepted answer success and internal errors", func(t *testing.T) {
		svc := &mockForumServiceForHandler{}
		h := NewForumHandler(svc)
		r := gin.New()
		r.PUT("/posts/:id/accept", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			c.Set("user_role", "member")
			h.MarkAcceptedAnswer(c)
		})
		r.DELETE("/forums/:slug/accepted-answer", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			c.Set("user_role", "member")
			h.UnmarkAcceptedAnswer(c)
		})

		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, httptest.NewRequest(http.MethodPut, "/posts/3/accept", nil))
		if w1.Code != http.StatusOK {
			t.Fatalf("expected 200 mark accepted success, got %d", w1.Code)
		}

		svc.markAsAcceptedAnswerFn = func(_ context.Context, _ uint, _ string, _ uint) error { return errors.New("db fail") }
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, httptest.NewRequest(http.MethodPut, "/posts/3/accept", nil))
		if w2.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 mark accepted internal, got %d", w2.Code)
		}

		w3 := httptest.NewRecorder()
		r.ServeHTTP(w3, httptest.NewRequest(http.MethodDelete, "/forums/slug-1/accepted-answer", nil))
		if w3.Code != http.StatusOK {
			t.Fatalf("expected 200 unmark accepted success, got %d", w3.Code)
		}

		svc.unmarkAcceptedBySlugFn = func(_ context.Context, _ uint, _ string, _ string) error { return errors.New("db fail") }
		w4 := httptest.NewRecorder()
		r.ServeHTTP(w4, httptest.NewRequest(http.MethodDelete, "/forums/slug-1/accepted-answer", nil))
		if w4.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 unmark accepted internal, got %d", w4.Code)
		}
	})
}
