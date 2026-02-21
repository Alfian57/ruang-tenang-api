package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
)

type mockForumRepo struct {
	forums        []model.Forum
	total         int64
	getForumsErr  error
	forumByID     *model.Forum
	forumByIDErr  error
	deleteErr     error
	deleteCalls   int
	likesMap      map[uint]int64
	repliesMap    map[uint]int64
	lastSort      repository.PostSortOption
	postsResult   []model.ForumPost
	postsTotal    int64
	postsSortErr  error
	createCalls   int
	createForumID uint
}

func (m *mockForumRepo) CreateForum(_ context.Context, forum *model.Forum) error {
	m.createCalls++
	m.createForumID = forum.UserID
	forum.ID = 99
	return nil
}
func (m *mockForumRepo) GetForums(_ context.Context, _, _ int, _ string, _ *uint) ([]model.Forum, int64, error) {
	if m.getForumsErr != nil {
		return nil, 0, m.getForumsErr
	}
	return m.forums, m.total, nil
}
func (m *mockForumRepo) GetForumByID(_ context.Context, _ uint) (*model.Forum, error) {
	if m.forumByIDErr != nil {
		return nil, m.forumByIDErr
	}
	return m.forumByID, nil
}
func (m *mockForumRepo) GetForumBySlug(_ context.Context, _ string) (*model.Forum, error) {
	return nil, errors.New("not implemented")
}
func (m *mockForumRepo) DeleteForum(_ context.Context, _ uint) error {
	m.deleteCalls++
	return m.deleteErr
}
func (m *mockForumRepo) UpdateForum(_ context.Context, _ *model.Forum) error         { return nil }
func (m *mockForumRepo) CreateForumPost(_ context.Context, _ *model.ForumPost) error { return nil }
func (m *mockForumRepo) GetForumPosts(_ context.Context, _ uint, _, _ int) ([]model.ForumPost, int64, error) {
	return nil, 0, nil
}
func (m *mockForumRepo) GetForumPostsSorted(_ context.Context, _ uint, _, _ int, sort repository.PostSortOption, _ uint) ([]model.ForumPost, int64, error) {
	m.lastSort = sort
	if m.postsSortErr != nil {
		return nil, 0, m.postsSortErr
	}
	return m.postsResult, m.postsTotal, nil
}
func (m *mockForumRepo) DeleteForumPost(_ context.Context, _ uint) error { return nil }
func (m *mockForumRepo) GetForumPostByID(_ context.Context, _ uint) (*model.ForumPost, error) {
	return nil, nil
}
func (m *mockForumRepo) UpdateForumPost(_ context.Context, _ *model.ForumPost) error { return nil }
func (m *mockForumRepo) ToggleLike(_ context.Context, _, _ uint) (bool, error)       { return false, nil }
func (m *mockForumRepo) GetLikesCount(_ context.Context, forumID uint) (int64, error) {
	if m.likesMap == nil {
		return 0, nil
	}
	return m.likesMap[forumID], nil
}
func (m *mockForumRepo) GetRepliesCount(_ context.Context, forumID uint) (int64, error) {
	if m.repliesMap == nil {
		return 0, nil
	}
	return m.repliesMap[forumID], nil
}
func (m *mockForumRepo) HasUserLiked(_ context.Context, _, _ uint) (bool, error) { return false, nil }
func (m *mockForumRepo) FindUpdatedSince(_ context.Context, _ time.Time) ([]model.Forum, error) {
	return nil, nil
}
func (m *mockForumRepo) VotePost(_ context.Context, _, _ uint, _ model.VoteType) error { return nil }
func (m *mockForumRepo) RemovePostVote(_ context.Context, _, _ uint) error             { return nil }
func (m *mockForumRepo) GetUserPostVote(_ context.Context, _, _ uint) (*model.ForumPostVote, error) {
	return nil, nil
}
func (m *mockForumRepo) GetPostVotesCount(_ context.Context, _ uint, _ model.VoteType) (int64, error) {
	return 0, nil
}
func (m *mockForumRepo) UpdatePostVoteCounts(_ context.Context, _ uint) error { return nil }
func (m *mockForumRepo) MarkAsAcceptedAnswer(_ context.Context, _ uint) error { return nil }
func (m *mockForumRepo) UnmarkAcceptedAnswer(_ context.Context, _ uint) error { return nil }
func (m *mockForumRepo) GetAcceptedAnswer(_ context.Context, _ uint) (*model.ForumPost, error) {
	return nil, nil
}
func (m *mockForumRepo) UpdateCommunityFavorite(_ context.Context, _ uint) error { return nil }
func (m *mockForumRepo) CreatePostReport(_ context.Context, _ *model.ForumPostReport) error {
	return nil
}
func (m *mockForumRepo) GetPostReports(_ context.Context, _ uint) ([]model.ForumPostReport, error) {
	return nil, nil
}
func (m *mockForumRepo) GetPendingPostReports(_ context.Context, _, _ int) ([]model.ForumPostReport, int64, error) {
	return nil, 0, nil
}
func (m *mockForumRepo) GetPostReportByID(_ context.Context, _ uint) (*model.ForumPostReport, error) {
	return nil, nil
}
func (m *mockForumRepo) UpdatePostReport(_ context.Context, _ *model.ForumPostReport) error {
	return nil
}
func (m *mockForumRepo) HasUserReportedPost(_ context.Context, _, _ uint) (bool, error) {
	return false, nil
}
func (m *mockForumRepo) GetPostReportsCount(_ context.Context, _ uint) (int64, error) { return 0, nil }
func (m *mockForumRepo) GetUserVotesForPosts(_ context.Context, _ uint, _ []uint) (map[uint]*model.ForumPostVote, error) {
	return nil, nil
}
func (m *mockForumRepo) CountUserVotesInLastHour(_ context.Context, _ uint) (int64, error) {
	return 0, nil
}
func (m *mockForumRepo) CountUserVotesInLastMinutes(_ context.Context, _ uint, _ int) (int64, error) {
	return 0, nil
}
func (m *mockForumRepo) CountUserVotesOnAuthorInLastHour(_ context.Context, _, _ uint) (int64, error) {
	return 0, nil
}

func TestForumService_GetForumsPopulatesCounts(t *testing.T) {
	repo := &mockForumRepo{
		forums: []model.Forum{{ID: 1, Title: "A"}, {ID: 2, Title: "B"}},
		total:  2,
		likesMap: map[uint]int64{
			1: 5,
			2: 3,
		},
		repliesMap: map[uint]int64{
			1: 2,
			2: 4,
		},
	}
	svc := &forumService{repo: repo}

	forums, total, err := svc.GetForums(context.Background(), 10, 0, "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 || len(forums) != 2 {
		t.Fatalf("unexpected result size: total=%d len=%d", total, len(forums))
	}
	if forums[0].LikesCount != 5 || forums[0].RepliesCount != 2 {
		t.Fatalf("unexpected forum[0] counts: %+v", forums[0])
	}
	if forums[1].LikesCount != 3 || forums[1].RepliesCount != 4 {
		t.Fatalf("unexpected forum[1] counts: %+v", forums[1])
	}
}

func TestForumService_DeleteForumAuthorization(t *testing.T) {
	repo := &mockForumRepo{forumByID: &model.Forum{ID: 10, UserID: 7}}
	svc := &forumService{repo: repo}

	err := svc.DeleteForum(context.Background(), 8, "member", 10)
	if err == nil || err.Error() != "unauthorized" {
		t.Fatalf("expected unauthorized error, got %v", err)
	}
	if repo.deleteCalls != 0 {
		t.Fatalf("delete should not be called on unauthorized request, got %d", repo.deleteCalls)
	}

	err = svc.DeleteForum(context.Background(), 7, "member", 10)
	if err != nil {
		t.Fatalf("expected owner can delete, got %v", err)
	}
	if repo.deleteCalls != 1 {
		t.Fatalf("expected one delete call, got %d", repo.deleteCalls)
	}
}

func TestForumService_GetForumPostsSortedMapsSortOption(t *testing.T) {
	repo := &mockForumRepo{postsResult: []model.ForumPost{{ID: 1}}, postsTotal: 1}
	svc := &forumService{repo: repo}

	_, _, err := svc.GetForumPostsSorted(context.Background(), 10, 20, 0, "newest", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.lastSort != repository.SortByNewest {
		t.Fatalf("expected sort newest, got %s", repo.lastSort)
	}

	_, _, _ = svc.GetForumPostsSorted(context.Background(), 10, 20, 0, "oldest", 3)
	if repo.lastSort != repository.SortByOldest {
		t.Fatalf("expected sort oldest, got %s", repo.lastSort)
	}

	_, _, _ = svc.GetForumPostsSorted(context.Background(), 10, 20, 0, "invalid", 3)
	if repo.lastSort != repository.SortByTop {
		t.Fatalf("expected default sort top, got %s", repo.lastSort)
	}
}

func TestForumService_CreateForumAndVoteValidation(t *testing.T) {
	repo := &mockForumRepo{}
	svc := &forumService{repo: repo}

	forum, err := svc.CreateForum(context.Background(), 5, "Topik", "Isi", nil)
	if err != nil {
		t.Fatalf("unexpected create forum error: %v", err)
	}
	if forum == nil || forum.UserID != 5 {
		t.Fatalf("unexpected forum result: %+v", forum)
	}
	if repo.createCalls != 1 || repo.createForumID != 5 {
		t.Fatalf("unexpected create calls state: calls=%d userID=%d", repo.createCalls, repo.createForumID)
	}

	if err := svc.VotePost(context.Background(), 5, 99, "invalid"); err == nil || err.Error() != "invalid vote type" {
		t.Fatalf("expected invalid vote type error, got %v", err)
	}
}
