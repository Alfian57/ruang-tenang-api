package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupForumRepoDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	schema := []string{
		`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, deleted_at DATETIME)`,
		`CREATE TABLE forum_categories (
			id INTEGER PRIMARY KEY,
			name TEXT,
			slug TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE forums (
			id INTEGER PRIMARY KEY,
			user_id INTEGER,
			category_id INTEGER,
			title TEXT,
			slug TEXT,
			content TEXT,
			trigger_warnings TEXT,
			is_flagged BOOLEAN,
			flagged_reason TEXT,
			has_accepted_answer BOOLEAN,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE forum_posts (
			id INTEGER PRIMARY KEY,
			forum_id INTEGER,
			user_id INTEGER,
			content TEXT,
			is_flagged BOOLEAN,
			flagged_reason TEXT,
			is_accepted_answer BOOLEAN,
			is_community_favorite BOOLEAN,
			upvotes_count INTEGER,
			downvotes_count INTEGER,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE forum_likes (
			id INTEGER PRIMARY KEY,
			forum_id INTEGER,
			user_id INTEGER,
			created_at DATETIME
		)`,
		`CREATE TABLE forum_post_votes (
			id INTEGER PRIMARY KEY,
			post_id INTEGER,
			user_id INTEGER,
			vote_type TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE forum_post_reports (
			id INTEGER PRIMARY KEY,
			post_id INTEGER,
			reporter_id INTEGER,
			reason TEXT,
			description TEXT,
			status TEXT,
			reviewed_by INTEGER,
			reviewed_at DATETIME,
			moderator_notes TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)`,
	}

	for _, q := range schema {
		if err := db.Exec(q).Error; err != nil {
			t.Fatalf("create schema failed: %v", err)
		}
	}

	now := time.Now()
	seed := []string{
		`INSERT INTO users (id, name) VALUES (1, 'u1'), (2, 'u2'), (3, 'u3')`,
		`INSERT INTO forum_categories (id, name, slug, created_at, updated_at) VALUES (1, 'General', 'general', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		`INSERT INTO forums (id, user_id, category_id, title, slug, content, is_flagged, has_accepted_answer, created_at, updated_at) VALUES (1, 1, 1, 'Forum 1', 'forum-1', 'content', 0, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		`INSERT INTO forum_posts (id, forum_id, user_id, content, is_accepted_answer, is_community_favorite, upvotes_count, downvotes_count, created_at, updated_at) VALUES (1, 1, 2, 'post 1', 0, 0, 0, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		`INSERT INTO forum_post_votes (id, post_id, user_id, vote_type, created_at, updated_at) VALUES (1, 1, 3, 'upvote', ?, ?)`,
	}
	for i, q := range seed {
		if i == len(seed)-1 {
			if err := db.Exec(q, now, now).Error; err != nil {
				t.Fatalf("seed failed: %v", err)
			}
			continue
		}
		if err := db.Exec(q).Error; err != nil {
			t.Fatalf("seed failed: %v", err)
		}
	}

	return db
}

func TestForumRepository_BasicAndPostOps(t *testing.T) {
	db := setupForumRepoDB(t)
	repo := NewForumRepository(db)
	ctx := context.Background()

	forum := &model.Forum{UserID: 1, CategoryID: ptrUint(1), Title: "New", Slug: "new-forum", Content: "hello"}
	if err := repo.CreateForum(ctx, forum); err != nil {
		t.Fatalf("create forum: %v", err)
	}

	forums, total, err := repo.GetForums(ctx, 10, 0, "", nil)
	if err != nil || total == 0 || len(forums) == 0 {
		t.Fatalf("get forums: total=%d len=%d err=%v", total, len(forums), err)
	}

	_, _, err = repo.GetForums(ctx, 10, 0, "abc", nil)
	if err == nil {
		t.Fatal("expected sqlite ILIKE error for search branch")
	}

	fByID, err := repo.GetForumByID(ctx, 1)
	if err != nil || fByID.ID != 1 {
		t.Fatalf("get forum by id: %v", err)
	}
	fBySlug, err := repo.GetForumBySlug(ctx, "forum-1")
	if err != nil || fBySlug.ID != 1 {
		t.Fatalf("get forum by slug: %v", err)
	}

	updateForum := &model.Forum{ID: 1, UserID: 1, CategoryID: ptrUint(1), Title: "updated", Slug: "forum-1", Content: "content"}
	if err := repo.UpdateForum(ctx, updateForum); err != nil {
		t.Fatalf("update forum: %v", err)
	}

	post := &model.ForumPost{ForumID: 1, UserID: 1, Content: "new post"}
	if err := repo.CreateForumPost(ctx, post); err != nil {
		t.Fatalf("create post: %v", err)
	}
	posts, pTotal, err := repo.GetForumPosts(ctx, 1, 20, 0)
	if err != nil || pTotal == 0 || len(posts) == 0 {
		t.Fatalf("get posts: total=%d len=%d err=%v", pTotal, len(posts), err)
	}
	postByID, err := repo.GetForumPostByID(ctx, post.ID)
	if err != nil || postByID.ID == 0 {
		t.Fatalf("get post by id: %v", err)
	}
	updatedPost := &model.ForumPost{ID: post.ID, ForumID: 1, UserID: 1, Content: "edited"}
	if err := repo.UpdateForumPost(ctx, updatedPost); err != nil {
		t.Fatalf("update post: %v", err)
	}

	sorted, _, err := repo.GetForumPostsSorted(ctx, 1, 20, 0, SortByTop, 3)
	if err != nil || len(sorted) == 0 {
		t.Fatalf("get sorted top: %v", err)
	}
	_, _, err = repo.GetForumPostsSorted(ctx, 1, 20, 0, SortByNewest, 0)
	if err != nil {
		t.Fatalf("get sorted newest: %v", err)
	}
	_, _, err = repo.GetForumPostsSorted(ctx, 1, 20, 0, SortByOldest, 0)
	if err != nil {
		t.Fatalf("get sorted oldest: %v", err)
	}

	if err := repo.DeleteForumPost(ctx, post.ID); err != nil {
		t.Fatalf("delete post: %v", err)
	}
	if err := repo.DeleteForum(ctx, forum.ID); err != nil {
		t.Fatalf("delete forum: %v", err)
	}
}

func TestForumRepository_LikesVotesReportsAndCounters(t *testing.T) {
	db := setupForumRepoDB(t)
	repo := NewForumRepository(db)
	ctx := context.Background()

	liked, err := repo.ToggleLike(ctx, 1, 1)
	if err != nil || !liked {
		t.Fatalf("toggle like create: liked=%v err=%v", liked, err)
	}
	liked, err = repo.ToggleLike(ctx, 1, 1)
	if err != nil || liked {
		t.Fatalf("toggle like delete: liked=%v err=%v", liked, err)
	}

	_, _ = repo.GetLikesCount(ctx, 1)
	_, _ = repo.GetRepliesCount(ctx, 1)
	_, _ = repo.HasUserLiked(ctx, 1, 1)
	_, _ = repo.FindUpdatedSince(ctx, time.Now().Add(-24*time.Hour))

	if err := repo.VotePost(ctx, 1, 1, model.VoteTypeUpvote); err != nil {
		t.Fatalf("vote upvote: %v", err)
	}
	if err := repo.VotePost(ctx, 1, 1, model.VoteTypeDownvote); err != nil {
		t.Fatalf("vote update: %v", err)
	}
	vote, err := repo.GetUserPostVote(ctx, 1, 1)
	if err != nil || vote == nil {
		t.Fatalf("get user vote: %v", err)
	}
	votesMap, err := repo.GetUserVotesForPosts(ctx, 1, []uint{1})
	if err != nil || len(votesMap) == 0 {
		t.Fatalf("get user votes for posts: %v", err)
	}
	_, _ = repo.GetPostVotesCount(ctx, 1, model.VoteTypeDownvote)
	if err := repo.UpdatePostVoteCounts(ctx, 1); err != nil {
		t.Fatalf("update post vote counts: %v", err)
	}
	if err := repo.RemovePostVote(ctx, 1, 1); err != nil {
		t.Fatalf("remove post vote: %v", err)
	}
	vote, err = repo.GetUserPostVote(ctx, 77, 1)
	if err != nil || vote != nil {
		t.Fatalf("expected nil vote, got %v err=%v", vote, err)
	}

	if err := repo.MarkAsAcceptedAnswer(ctx, 1); err != nil {
		t.Fatalf("mark accepted answer: %v", err)
	}
	accepted, err := repo.GetAcceptedAnswer(ctx, 1)
	if err != nil || accepted == nil {
		t.Fatalf("get accepted answer: %v", err)
	}
	if err := repo.UnmarkAcceptedAnswer(ctx, 1); err != nil {
		t.Fatalf("unmark accepted answer: %v", err)
	}
	accepted, err = repo.GetAcceptedAnswer(ctx, 1)
	if err != nil || accepted != nil {
		t.Fatalf("expected nil accepted answer, got %v err=%v", accepted, err)
	}

	if err := db.Exec(`UPDATE forum_posts SET upvotes_count = 4 WHERE id = 1`).Error; err != nil {
		t.Fatalf("seed upvotes for favorite: %v", err)
	}
	if err := repo.UpdateCommunityFavorite(ctx, 1); err != nil {
		t.Fatalf("update community favorite: %v", err)
	}

	report := &model.ForumPostReport{PostID: 1, ReporterID: 2, Reason: model.PostReportReasonSpam, Description: "spam", Status: model.PostReportStatusPending}
	if err := repo.CreatePostReport(ctx, report); err != nil {
		t.Fatalf("create post report: %v", err)
	}
	reports, err := repo.GetPostReports(ctx, 1)
	if err != nil || len(reports) == 0 {
		t.Fatalf("get post reports: %v", err)
	}
	pending, pTotal, err := repo.GetPendingPostReports(ctx, 10, 0)
	if err != nil || pTotal == 0 || len(pending) == 0 {
		t.Fatalf("get pending reports: total=%d len=%d err=%v", pTotal, len(pending), err)
	}
	reportByID, err := repo.GetPostReportByID(ctx, report.ID)
	if err != nil || reportByID.ID == 0 {
		t.Fatalf("get post report by id: %v", err)
	}
	reviewerID := uint(3)
	now := time.Now()
	updatedReport := &model.ForumPostReport{
		ID:             report.ID,
		PostID:         1,
		ReporterID:     2,
		Reason:         model.PostReportReasonSpam,
		Description:    "spam",
		Status:         model.PostReportStatusReviewed,
		ReviewedBy:     &reviewerID,
		ReviewedAt:     &now,
		ModeratorNotes: "checked",
	}
	if err := repo.UpdatePostReport(ctx, updatedReport); err != nil {
		t.Fatalf("update post report: %v", err)
	}
	_, _ = repo.HasUserReportedPost(ctx, 2, 1)
	_, _ = repo.GetPostReportsCount(ctx, 1)

	_, _ = repo.CountUserVotesInLastHour(ctx, 3)
	_, _ = repo.CountUserVotesInLastMinutes(ctx, 3, 30)
	_, _ = repo.CountUserVotesOnAuthorInLastHour(ctx, 3, 2)
}

func TestForumRepository_AcceptedAnswerAndFavorite_EdgeBranches(t *testing.T) {
	ctx := context.Background()

	t.Run("mark accepted answer not found", func(t *testing.T) {
		db := setupForumRepoDB(t)
		repo := NewForumRepository(db)

		if err := repo.MarkAsAcceptedAnswer(ctx, 9999); err == nil {
			t.Fatal("expected error when post does not exist")
		}
	})

	t.Run("mark accepted answer forum update error", func(t *testing.T) {
		db := setupForumRepoDB(t)
		repo := NewForumRepository(db)

		if err := db.Exec(`DROP TABLE forums`).Error; err != nil {
			t.Fatalf("drop forums: %v", err)
		}

		if err := repo.MarkAsAcceptedAnswer(ctx, 1); err == nil {
			t.Fatal("expected error when forums table is missing")
		}
	})

	t.Run("unmark accepted answer forum update error", func(t *testing.T) {
		db := setupForumRepoDB(t)
		repo := NewForumRepository(db)

		if err := db.Exec(`DROP TABLE forums`).Error; err != nil {
			t.Fatalf("drop forums: %v", err)
		}

		if err := repo.UnmarkAcceptedAnswer(ctx, 1); err == nil {
			t.Fatal("expected error when forums table is missing")
		}
	})

	t.Run("unmark accepted answer post update error", func(t *testing.T) {
		db := setupForumRepoDB(t)
		repo := NewForumRepository(db)

		if err := db.Exec(`DROP TABLE forum_posts`).Error; err != nil {
			t.Fatalf("drop forum_posts: %v", err)
		}

		if err := repo.UnmarkAcceptedAnswer(ctx, 1); err == nil {
			t.Fatal("expected error when forum_posts table is missing")
		}
	})

	t.Run("community favorite no qualifying post", func(t *testing.T) {
		db := setupForumRepoDB(t)
		repo := NewForumRepository(db)

		if err := repo.UpdateCommunityFavorite(ctx, 9999); err != nil {
			t.Fatalf("expected nil when no qualifying post exists, got %v", err)
		}
	})

	t.Run("community favorite reset error", func(t *testing.T) {
		db := setupForumRepoDB(t)
		repo := NewForumRepository(db)

		if err := db.Exec(`DROP TABLE forum_posts`).Error; err != nil {
			t.Fatalf("drop forum_posts: %v", err)
		}

		if err := repo.UpdateCommunityFavorite(ctx, 1); err == nil {
			t.Fatal("expected error when forum_posts table is missing")
		}
	})
}

func TestForumRepository_Getters_NotFoundBranches(t *testing.T) {
	db := setupForumRepoDB(t)
	repo := NewForumRepository(db)
	ctx := context.Background()

	if _, err := repo.GetForumByID(ctx, 9999); err == nil {
		t.Fatal("expected GetForumByID to return error for missing id")
	}

	if _, err := repo.GetForumBySlug(ctx, "missing-slug"); err == nil {
		t.Fatal("expected GetForumBySlug to return error for missing slug")
	}

	if _, err := repo.GetForumPostByID(ctx, 9999); err == nil {
		t.Fatal("expected GetForumPostByID to return error for missing id")
	}

	if _, err := repo.GetPostReportByID(ctx, 9999); err == nil {
		t.Fatal("expected GetPostReportByID to return error for missing id")
	}
}

func ptrUint(v uint) *uint {
	return &v
}
