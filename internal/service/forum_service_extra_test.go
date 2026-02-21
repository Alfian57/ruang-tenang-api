package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type mockForumRepoExtra struct {
	*mockForumRepo

	forumBySlug      *model.Forum
	forumBySlugErr   error
	toggleLikeResult bool
	toggleLikeErr    error
	hasUserLiked     bool
	hasUserLikedErr  error

	postByID    *model.ForumPost
	postByIDErr error

	hasReported         bool
	createPostReportErr error
	createPostReportHit int
	createdReport       *model.ForumPostReport

	postReportByID    *model.ForumPostReport
	postReportByIDErr error
	updatePostReport  *model.ForumPostReport

	voteCount    int64
	burstCount   int64
	authorVotes  int64
	votePostErr  error
	unmarkAccErr error

	createPostErr   error
	createPostCalls int

	forumPosts    []model.ForumPost
	forumPostsTot int64
	forumPostsErr error

	deletePostErr   error
	deletePostCalls int

	userVote    *model.ForumPostVote
	userVoteErr error

	removeVoteErr     error
	removeVoteCalls   int
	markAcceptedErr   error
	unmarkAcceptedErr error
	acceptedAnswer    *model.ForumPost
	acceptedAnswerErr error

	pendingReports    []model.ForumPostReport
	pendingReportsTot int64
	pendingReportsErr error
}

func (m *mockForumRepoExtra) GetForumBySlug(_ context.Context, _ string) (*model.Forum, error) {
	if m.forumBySlugErr != nil {
		return nil, m.forumBySlugErr
	}
	return m.forumBySlug, nil
}

func (m *mockForumRepoExtra) ToggleLike(_ context.Context, _, _ uint) (bool, error) {
	if m.toggleLikeErr != nil {
		return false, m.toggleLikeErr
	}
	return m.toggleLikeResult, nil
}

func (m *mockForumRepoExtra) HasUserLiked(_ context.Context, _, _ uint) (bool, error) {
	if m.hasUserLikedErr != nil {
		return false, m.hasUserLikedErr
	}
	return m.hasUserLiked, nil
}

func (m *mockForumRepoExtra) GetForumPostByID(_ context.Context, _ uint) (*model.ForumPost, error) {
	if m.postByIDErr != nil {
		return nil, m.postByIDErr
	}
	return m.postByID, nil
}

func (m *mockForumRepoExtra) HasUserReportedPost(_ context.Context, _, _ uint) (bool, error) {
	return m.hasReported, nil
}

func (m *mockForumRepoExtra) CreatePostReport(_ context.Context, report *model.ForumPostReport) error {
	m.createPostReportHit++
	m.createdReport = report
	return m.createPostReportErr
}

func (m *mockForumRepoExtra) GetPostReportByID(_ context.Context, _ uint) (*model.ForumPostReport, error) {
	if m.postReportByIDErr != nil {
		return nil, m.postReportByIDErr
	}
	return m.postReportByID, nil
}

func (m *mockForumRepoExtra) UpdatePostReport(_ context.Context, report *model.ForumPostReport) error {
	m.updatePostReport = report
	return nil
}

func (m *mockForumRepoExtra) CountUserVotesInLastHour(_ context.Context, _ uint) (int64, error) {
	return m.voteCount, nil
}

func (m *mockForumRepoExtra) CountUserVotesInLastMinutes(_ context.Context, _ uint, _ int) (int64, error) {
	return m.burstCount, nil
}

func (m *mockForumRepoExtra) CountUserVotesOnAuthorInLastHour(_ context.Context, _, _ uint) (int64, error) {
	return m.authorVotes, nil
}

func (m *mockForumRepoExtra) VotePost(_ context.Context, _, _ uint, _ model.VoteType) error {
	return m.votePostErr
}

func (m *mockForumRepoExtra) CreateForumPost(_ context.Context, _ *model.ForumPost) error {
	m.createPostCalls++
	return m.createPostErr
}

func (m *mockForumRepoExtra) GetForumPosts(_ context.Context, _ uint, _, _ int) ([]model.ForumPost, int64, error) {
	if m.forumPostsErr != nil {
		return nil, 0, m.forumPostsErr
	}
	return m.forumPosts, m.forumPostsTot, nil
}

func (m *mockForumRepoExtra) DeleteForumPost(_ context.Context, _ uint) error {
	m.deletePostCalls++
	return m.deletePostErr
}

func (m *mockForumRepoExtra) GetUserPostVote(_ context.Context, _, _ uint) (*model.ForumPostVote, error) {
	if m.userVoteErr != nil {
		return nil, m.userVoteErr
	}
	return m.userVote, nil
}

func (m *mockForumRepoExtra) RemovePostVote(_ context.Context, _, _ uint) error {
	m.removeVoteCalls++
	return m.removeVoteErr
}

func (m *mockForumRepoExtra) MarkAsAcceptedAnswer(_ context.Context, _ uint) error {
	return m.markAcceptedErr
}

func (m *mockForumRepoExtra) UnmarkAcceptedAnswer(_ context.Context, _ uint) error {
	if m.unmarkAcceptedErr != nil {
		return m.unmarkAcceptedErr
	}
	return m.unmarkAccErr
}

func (m *mockForumRepoExtra) GetAcceptedAnswer(_ context.Context, _ uint) (*model.ForumPost, error) {
	if m.acceptedAnswerErr != nil {
		return nil, m.acceptedAnswerErr
	}
	return m.acceptedAnswer, nil
}

func (m *mockForumRepoExtra) GetPendingPostReports(_ context.Context, _, _ int) ([]model.ForumPostReport, int64, error) {
	if m.pendingReportsErr != nil {
		return nil, 0, m.pendingReportsErr
	}
	return m.pendingReports, m.pendingReportsTot, nil
}

func newTestUserRepo(t *testing.T) *repository.UserRepository {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("migrate user: %v", err)
	}
	return repository.NewUserRepository(db)
}

func TestForumService_GetByIDSlug_DeleteAndLikeBySlug(t *testing.T) {
	repo := &mockForumRepoExtra{
		mockForumRepo:    &mockForumRepo{forumByID: &model.Forum{ID: 1, UserID: 7}, likesMap: map[uint]int64{1: 3}, repliesMap: map[uint]int64{1: 2}},
		forumBySlug:      &model.Forum{ID: 2, UserID: 9},
		toggleLikeResult: true,
		hasUserLiked:     true,
	}
	svc := &forumService{repo: repo}

	byID, err := svc.GetForumByID(context.Background(), 99, 1)
	if err != nil || byID.LikesCount != 3 || byID.RepliesCount != 2 || !byID.IsLiked {
		t.Fatalf("unexpected get by id result: forum=%+v err=%v", byID, err)
	}

	bySlug, err := svc.GetForumBySlug(context.Background(), 99, "any")
	if err != nil || bySlug.ID != 2 || !bySlug.IsLiked {
		t.Fatalf("unexpected get by slug result: forum=%+v err=%v", bySlug, err)
	}

	if err := svc.DeleteForumBySlug(context.Background(), 1, "member", "any"); err == nil {
		t.Fatalf("expected unauthorized delete")
	}
	if err := svc.DeleteForumBySlug(context.Background(), 9, "member", "any"); err != nil {
		t.Fatalf("expected owner can delete, got %v", err)
	}

	liked, err := svc.ToggleLikeBySlug(context.Background(), 1, "any")
	if err != nil || !liked {
		t.Fatalf("unexpected toggle like by slug: liked=%v err=%v", liked, err)
	}
}

func TestForumService_SlugPostAndReportPaths(t *testing.T) {
	repo := &mockForumRepoExtra{mockForumRepo: &mockForumRepo{postsResult: []model.ForumPost{{ID: 10}}, postsTotal: 1}, forumBySlug: &model.Forum{ID: 4}, postByID: &model.ForumPost{ID: 11, UserID: 100, ForumID: 4}}
	svc := &forumService{repo: repo}

	posts, total, err := svc.GetForumPostsSortedBySlug(context.Background(), "slug", 10, 0, "oldest", 5)
	if err != nil || len(posts) != 1 || total != 1 || repo.lastSort != repository.SortByOldest {
		t.Fatalf("unexpected sorted by slug result: posts=%d total=%d sort=%s err=%v", len(posts), total, repo.lastSort, err)
	}

	repo.forumBySlugErr = errors.New("missing")
	if _, _, err := svc.GetForumPostsBySlug(context.Background(), "missing", 10, 0); err == nil {
		t.Fatalf("expected get posts by slug error")
	}
	repo.forumBySlugErr = nil

	if err := svc.ReportPost(context.Background(), 1, 11, "invalid", "desc"); err == nil {
		t.Fatalf("expected invalid reason error")
	}

	repo.postByIDErr = errors.New("not found")
	if err := svc.ReportPost(context.Background(), 1, 11, string(model.PostReportReasonSpam), "desc"); err == nil {
		t.Fatalf("expected post not found error")
	}
	repo.postByIDErr = nil

	repo.postByID = &model.ForumPost{ID: 11, UserID: 1, ForumID: 4}
	if err := svc.ReportPost(context.Background(), 1, 11, string(model.PostReportReasonSpam), "desc"); err == nil {
		t.Fatalf("expected own post report error")
	}

	repo.postByID = &model.ForumPost{ID: 11, UserID: 2, ForumID: 4}
	repo.hasReported = true
	if err := svc.ReportPost(context.Background(), 1, 11, string(model.PostReportReasonSpam), "desc"); err == nil {
		t.Fatalf("expected already reported error")
	}

	repo.hasReported = false
	if err := svc.ReportPost(context.Background(), 1, 11, string(model.PostReportReasonSpam), "desc"); err != nil {
		t.Fatalf("expected report creation success, got %v", err)
	}
	if repo.createPostReportHit != 1 || repo.createdReport == nil {
		t.Fatalf("expected one post report creation")
	}
}

func TestForumService_ReviewPostReportAndVoteValidation(t *testing.T) {
	repo := &mockForumRepoExtra{
		mockForumRepo:  &mockForumRepo{forumByID: &model.Forum{ID: 4, UserID: 2}},
		postByID:       &model.ForumPost{ID: 11, UserID: 3, ForumID: 4},
		postReportByID: &model.ForumPostReport{ID: 77, PostID: 11, Reason: model.PostReportReasonSpam},
	}
	svc := &forumService{repo: repo}

	if err := svc.ReviewPostReport(context.Background(), 9, 77, "wrong", "n"); err == nil {
		t.Fatalf("expected invalid status error")
	}

	repo.postReportByIDErr = errors.New("missing")
	if err := svc.ReviewPostReport(context.Background(), 9, 77, "reviewed", "n"); err == nil {
		t.Fatalf("expected report not found error")
	}
	repo.postReportByIDErr = nil

	if err := svc.ReviewPostReport(context.Background(), 9, 77, "actioned", "danger"); err != nil {
		t.Fatalf("expected actioned review success, got %v", err)
	}
	if repo.updatePostReport == nil || repo.updatePostReport.ReviewedBy == nil || *repo.updatePostReport.ReviewedBy != 9 {
		t.Fatalf("expected reviewed report to be updated with reviewer")
	}

	repo.postByIDErr = errors.New("missing")
	if err := svc.VotePost(context.Background(), 1, 11, "upvote"); err == nil {
		t.Fatalf("expected post not found error")
	}
	repo.postByIDErr = nil

	repo.postByID = &model.ForumPost{ID: 11, UserID: 1, ForumID: 4}
	if err := svc.VotePost(context.Background(), 1, 11, "upvote"); err == nil {
		t.Fatalf("expected self vote error")
	}

	repo.postByID = &model.ForumPost{ID: 11, UserID: 2, ForumID: 4}
	repo.voteCount = 30
	if err := svc.VotePost(context.Background(), 1, 11, "upvote"); err == nil {
		t.Fatalf("expected hourly rate limit error")
	}
	repo.voteCount = 0

	repo.burstCount = 10
	if err := svc.VotePost(context.Background(), 1, 11, "upvote"); err == nil {
		t.Fatalf("expected burst voting error")
	}
	repo.burstCount = 0

	repo.authorVotes = 5
	if err := svc.VotePost(context.Background(), 1, 11, "upvote"); err == nil {
		t.Fatalf("expected same-author target error")
	}

	if err := svc.VotePost(context.Background(), 1, 11, "invalid"); err == nil {
		t.Fatalf("expected invalid vote type error")
	}
}

func TestForumService_CheckForumBlockedAndAccountAge(t *testing.T) {
	ctx := context.Background()
	userRepo := newTestUserRepo(t)

	oldUser := &model.User{Name: "Old", Username: "olduser", Email: "old@x.id", Password: "x", CreatedAt: time.Now().Add(-48 * time.Hour)}
	newUser := &model.User{Name: "New", Username: "newuser", Email: "new@x.id", Password: "x", CreatedAt: time.Now().Add(-2 * time.Hour)}
	blocked := &model.User{Name: "Blocked", Username: "blockeduser", Email: "blocked@x.id", Password: "x", IsForumBlocked: true, CreatedAt: time.Now().Add(-72 * time.Hour)}

	if err := userRepo.Create(ctx, oldUser); err != nil {
		t.Fatalf("create old user: %v", err)
	}
	if err := userRepo.Create(ctx, newUser); err != nil {
		t.Fatalf("create new user: %v", err)
	}
	if err := userRepo.Create(ctx, blocked); err != nil {
		t.Fatalf("create blocked user: %v", err)
	}

	repo := &mockForumRepoExtra{mockForumRepo: &mockForumRepo{}, postByID: &model.ForumPost{ID: 11, UserID: oldUser.ID + 100, ForumID: 4}}
	svc := &forumService{repo: repo, userRepo: userRepo}

	if _, err := svc.CreateForum(ctx, blocked.ID, "t", "c", nil); !errors.Is(err, ErrForumBlocked) {
		t.Fatalf("expected forum blocked error, got %v", err)
	}

	if err := svc.VotePost(ctx, newUser.ID, 11, "upvote"); err == nil {
		t.Fatalf("expected account age restriction error")
	}
}

func TestForumService_PostOpsAndStats(t *testing.T) {
	ctx := context.Background()
	repo := &mockForumRepoExtra{
		mockForumRepo: &mockForumRepo{likesMap: map[uint]int64{4: 9}},
		forumBySlug:   &model.Forum{ID: 4, UserID: 7},
		postByID:      &model.ForumPost{ID: 11, UserID: 7, ForumID: 4},
		forumPosts:    []model.ForumPost{{ID: 1}, {ID: 2}},
		forumPostsTot: 2,
	}

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	svc := &forumService{repo: repo, gamificationService: NewGamificationService(db)}

	if err := svc.CreateForumPost(ctx, 1, 4, "hello"); err != nil {
		t.Fatalf("create forum post expected success: %v", err)
	}
	if repo.createPostCalls != 1 {
		t.Fatalf("expected one create post call, got %d", repo.createPostCalls)
	}

	repo.createPostErr = errors.New("create failed")
	if err := svc.CreateForumPost(ctx, 1, 4, "hello"); err == nil {
		t.Fatalf("expected create forum post error")
	}
	repo.createPostErr = nil

	if err := svc.CreateForumPostBySlug(ctx, 1, "slug", "hi"); err != nil {
		t.Fatalf("create forum post by slug expected success: %v", err)
	}

	posts, total, err := svc.GetForumPosts(ctx, 4, 10, 0)
	if err != nil || len(posts) != 2 || total != 2 {
		t.Fatalf("unexpected forum posts result: len=%d total=%d err=%v", len(posts), total, err)
	}

	repo.postByID = &model.ForumPost{ID: 11, UserID: 2, ForumID: 4}
	if err := svc.DeleteForumPost(ctx, 1, "member", 11); err == nil || err.Error() != "unauthorized" {
		t.Fatalf("expected unauthorized delete forum post, got %v", err)
	}
	if err := svc.DeleteForumPost(ctx, 2, "member", 11); err != nil {
		t.Fatalf("expected owner delete forum post success: %v", err)
	}

	liked, err := svc.ToggleLike(ctx, 1, 4)
	if err != nil || liked {
		t.Fatalf("expected default toggle like false,nil got liked=%v err=%v", liked, err)
	}

	likes, err := svc.GetForumStats(ctx, 4)
	if err != nil || likes != 9 {
		t.Fatalf("expected likes=9, got %d err=%v", likes, err)
	}
}

func TestForumService_MoreErrorAndSuccessBranches(t *testing.T) {
	ctx := context.Background()
	repo := &mockForumRepoExtra{
		mockForumRepo: &mockForumRepo{forumByID: &model.Forum{ID: 4, UserID: 10}},
		forumBySlug:   &model.Forum{ID: 4, UserID: 10},
		postByID:      &model.ForumPost{ID: 11, UserID: 20, ForumID: 4},
	}
	svc := &forumService{repo: repo}

	repo.forumByIDErr = errors.New("missing forum")
	if err := svc.DeleteForum(ctx, 10, "admin", 4); err == nil {
		t.Fatal("expected delete forum by id to fail when forum missing")
	}
	repo.forumByIDErr = nil

	repo.forumBySlugErr = errors.New("missing slug")
	if err := svc.DeleteForumBySlug(ctx, 10, "admin", "slug"); err == nil {
		t.Fatal("expected delete forum by slug to fail when forum missing")
	}
	if _, err := svc.ToggleLikeBySlug(ctx, 1, "slug"); err == nil {
		t.Fatal("expected toggle like by slug to fail when forum missing")
	}
	if _, _, err := svc.GetForumPostsSortedBySlug(ctx, "slug", 10, 0, "top", 1); err == nil {
		t.Fatal("expected sorted by slug to fail when forum missing")
	}
	if _, err := svc.GetAcceptedAnswerBySlug(ctx, "slug"); err == nil {
		t.Fatal("expected accepted answer by slug to fail when forum missing")
	}
	if err := svc.UnmarkAcceptedAnswerBySlug(ctx, 1, "member", "slug"); err == nil {
		t.Fatal("expected unmark by slug to fail when forum missing")
	}
	repo.forumBySlugErr = nil

	repo.forumBySlugErr = errors.New("post missing")
	if err := svc.CreateForumPostBySlug(ctx, 1, "slug", "hello"); err == nil {
		t.Fatal("expected create post by slug to fail when create fails")
	}
	repo.forumBySlugErr = nil

	repo.createPostErr = errors.New("create fail")
	if err := svc.CreateForumPostBySlug(ctx, 1, "slug", "hello"); err == nil {
		t.Fatal("expected create forum post by slug create error")
	}
	repo.createPostErr = nil

	repo.votePostErr = errors.New("vote fail")
	if err := svc.VotePost(ctx, 1, 11, "downvote"); err == nil {
		t.Fatal("expected downvote to surface repo vote error")
	}
	repo.votePostErr = nil

	repo.userVote = nil
	if err := svc.RemovePostVote(ctx, 1, 11); err == nil {
		t.Fatal("expected no vote to remove error")
	}

	repo.userVote = &model.ForumPostVote{UserID: 1, PostID: 11, VoteType: model.VoteTypeUpvote}
	repo.postByIDErr = errors.New("missing post")
	if err := svc.RemovePostVote(ctx, 1, 11); err == nil {
		t.Fatal("expected post not found on remove vote")
	}
	repo.postByIDErr = nil

	repo.removeVoteErr = errors.New("remove fail")
	if err := svc.RemovePostVote(ctx, 1, 11); err == nil {
		t.Fatal("expected remove vote repo error")
	}
	repo.removeVoteErr = nil
	if err := svc.RemovePostVote(ctx, 1, 11); err != nil {
		t.Fatalf("expected remove vote success, got %v", err)
	}
}

func TestForumService_AcceptedAnswerAdditionalBranches(t *testing.T) {
	ctx := context.Background()
	repo := &mockForumRepoExtra{
		mockForumRepo: &mockForumRepo{forumByID: &model.Forum{ID: 4, UserID: 10}},
		forumBySlug:   &model.Forum{ID: 4, UserID: 10},
		postByID:      &model.ForumPost{ID: 11, UserID: 20, ForumID: 4, IsAcceptedAnswer: false},
	}
	svc := &forumService{repo: repo}

	repo.postByIDErr = errors.New("missing")
	if err := svc.MarkAsAcceptedAnswer(ctx, 10, "admin", 11); err == nil {
		t.Fatal("expected mark accepted to fail when post missing")
	}
	repo.postByIDErr = nil

	repo.forumByIDErr = errors.New("missing forum")
	if err := svc.MarkAsAcceptedAnswer(ctx, 10, "admin", 11); err == nil {
		t.Fatal("expected mark accepted to fail when forum missing")
	}
	repo.forumByIDErr = nil

	if err := svc.MarkAsAcceptedAnswer(ctx, 99, "member", 11); err == nil {
		t.Fatal("expected unauthorized mark accepted")
	}

	repo.postByID = &model.ForumPost{ID: 11, UserID: 10, ForumID: 4}
	if err := svc.MarkAsAcceptedAnswer(ctx, 10, "member", 11); err == nil {
		t.Fatal("expected cannot mark own answer")
	}

	repo.postByID = &model.ForumPost{ID: 11, UserID: 20, ForumID: 4}
	repo.markAcceptedErr = errors.New("mark fail")
	if err := svc.MarkAsAcceptedAnswer(ctx, 10, "admin", 11); err == nil {
		t.Fatal("expected mark accepted repo error")
	}
	repo.markAcceptedErr = nil
	repo.postByID.IsAcceptedAnswer = true
	if err := svc.MarkAsAcceptedAnswer(ctx, 10, "admin", 11); err != nil {
		t.Fatalf("expected mark accepted success, got %v", err)
	}

	repo.forumByIDErr = errors.New("missing forum")
	if err := svc.UnmarkAcceptedAnswer(ctx, 10, "admin", 4); err == nil {
		t.Fatal("expected unmark accepted fail when forum missing")
	}
	repo.forumByIDErr = nil

	if err := svc.UnmarkAcceptedAnswer(ctx, 99, "member", 4); err == nil {
		t.Fatal("expected unauthorized unmark accepted")
	}

	repo.unmarkAcceptedErr = errors.New("unmark fail")
	if err := svc.UnmarkAcceptedAnswer(ctx, 10, "member", 4); err == nil {
		t.Fatal("expected unmark accepted repo error")
	}
	repo.unmarkAcceptedErr = nil

	repo.acceptedAnswer = &model.ForumPost{ID: 1}
	if _, err := svc.GetAcceptedAnswerBySlug(ctx, "slug-ok"); err != nil {
		t.Fatalf("expected get accepted answer by slug success, got %v", err)
	}

	if err := svc.UnmarkAcceptedAnswerBySlug(ctx, 10, "member", "slug-ok"); err != nil {
		t.Fatalf("expected unmark accepted by slug success, got %v", err)
	}
}

func TestForumService_VoteStatusAcceptedAnswerAndReports(t *testing.T) {
	ctx := context.Background()
	repo := &mockForumRepoExtra{
		mockForumRepo:     &mockForumRepo{forumByID: &model.Forum{ID: 4, UserID: 10}},
		forumBySlug:       &model.Forum{ID: 4, UserID: 10},
		postByID:          &model.ForumPost{ID: 21, UserID: 33, ForumID: 4},
		userVote:          &model.ForumPostVote{UserID: 1, PostID: 21, VoteType: model.VoteTypeUpvote},
		acceptedAnswer:    &model.ForumPost{ID: 21, ForumID: 4, IsAcceptedAnswer: true},
		pendingReports:    []model.ForumPostReport{{ID: 1}, {ID: 2}},
		pendingReportsTot: 2,
	}

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	svc := &forumService{repo: repo, gamificationService: NewGamificationService(db)}

	if err := svc.RemovePostVote(ctx, 1, 21); err != nil {
		t.Fatalf("expected remove post vote success: %v", err)
	}
	if repo.removeVoteCalls != 1 {
		t.Fatalf("expected one remove vote call, got %d", repo.removeVoteCalls)
	}

	repo.userVote = nil
	if err := svc.RemovePostVote(ctx, 1, 21); err == nil || err.Error() != "no vote to remove" {
		t.Fatalf("expected no vote to remove, got %v", err)
	}
	repo.userVote = &model.ForumPostVote{UserID: 1, PostID: 21, VoteType: model.VoteTypeUpvote}

	status, err := svc.GetPostVoteStatus(ctx, 1, 21)
	if err != nil || status == nil || status.PostID != 21 {
		t.Fatalf("unexpected vote status: %+v err=%v", status, err)
	}

	if err := svc.MarkAsAcceptedAnswer(ctx, 99, "member", 21); err == nil {
		t.Fatalf("expected unauthorized mark accepted answer")
	}
	if err := svc.MarkAsAcceptedAnswer(ctx, 10, "member", 21); err != nil {
		t.Fatalf("expected owner mark accepted answer success: %v", err)
	}

	if err := svc.UnmarkAcceptedAnswer(ctx, 99, "member", 4); err == nil {
		t.Fatalf("expected unauthorized unmark accepted answer")
	}
	if err := svc.UnmarkAcceptedAnswerBySlug(ctx, 10, "member", "slug"); err != nil {
		t.Fatalf("expected unmark by slug success: %v", err)
	}

	accepted, err := svc.GetAcceptedAnswer(ctx, 4)
	if err != nil || accepted == nil || accepted.ID != 21 {
		t.Fatalf("unexpected accepted answer: %+v err=%v", accepted, err)
	}
	acceptedBySlug, err := svc.GetAcceptedAnswerBySlug(ctx, "slug")
	if err != nil || acceptedBySlug == nil || acceptedBySlug.ID != 21 {
		t.Fatalf("unexpected accepted answer by slug: %+v err=%v", acceptedBySlug, err)
	}

	reports, total, err := svc.GetPendingPostReports(ctx, 10, 0)
	if err != nil || len(reports) != 2 || total != 2 {
		t.Fatalf("unexpected pending reports result: len=%d total=%d err=%v", len(reports), total, err)
	}
}
