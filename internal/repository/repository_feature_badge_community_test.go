package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newExtRepoDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.User{},
		&model.LevelConfig{},
		&model.ExpHistory{},
		&model.Article{},
		&model.ArticleCategory{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	statements := []string{
		`CREATE TABLE IF NOT EXISTS feature_definitions (
			id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
			feature_key TEXT NOT NULL UNIQUE,
			feature_name TEXT NOT NULL,
			description TEXT,
			icon TEXT,
			required_level INTEGER NOT NULL DEFAULT 1,
			category TEXT DEFAULT 'general',
			is_active NUMERIC DEFAULT 1,
			display_order INTEGER DEFAULT 0,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS user_feature_unlocks (
			id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
			user_id INTEGER NOT NULL,
			feature_id TEXT NOT NULL,
			unlocked_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS badge_definitions (
			id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
			badge_key TEXT NOT NULL UNIQUE,
			badge_name TEXT NOT NULL,
			description TEXT,
			icon TEXT,
			category TEXT DEFAULT 'general',
			requirement_type TEXT NOT NULL,
			requirement_value INTEGER DEFAULT 0,
			is_active NUMERIC DEFAULT 1,
			display_order INTEGER DEFAULT 0,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS user_badges (
			id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
			user_id INTEGER NOT NULL,
			badge_id TEXT NOT NULL,
			earned_at DATETIME,
			is_showcased NUMERIC DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS community_stats (
			id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
			month INTEGER NOT NULL,
			year INTEGER NOT NULL,
			total_xp_earned INTEGER DEFAULT 0,
			active_members INTEGER DEFAULT 0,
			total_achievements INTEGER DEFAULT 0,
			new_members INTEGER DEFAULT 0,
			total_stories_published INTEGER DEFAULT 0,
			total_articles_published INTEGER DEFAULT 0,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS monthly_hall_of_fame (
			id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
			user_id INTEGER NOT NULL,
			level INTEGER NOT NULL,
			month INTEGER NOT NULL,
			year INTEGER NOT NULL,
			rank INTEGER NOT NULL,
			monthly_xp INTEGER NOT NULL,
			message TEXT,
			created_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS inspiring_stories (
			id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
			author_id INTEGER NOT NULL,
			title TEXT,
			content TEXT,
			cover_image TEXT,
			is_anonymous NUMERIC,
			has_trigger_warning NUMERIC,
			trigger_warning_text TEXT,
			status TEXT,
			moderator_id INTEGER,
			moderation_feedback TEXT,
			moderated_at DATETIME,
			view_count INTEGER,
			heart_count INTEGER,
			comment_count INTEGER,
			is_featured NUMERIC,
			featured_at DATETIME,
			featured_until DATETIME,
			created_at DATETIME
			,updated_at DATETIME
			,published_at DATETIME
		)`,
	}

	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("exec schema statement failed: %v", err)
		}
	}

	return db
}

func TestFeatureUnlockRepository_FullFlow(t *testing.T) {
	ctx := context.Background()
	db := newExtRepoDB(t)
	r := NewFeatureUnlockRepository(db)
	ur := NewUserRepository(db)

	u := &model.User{Name: "Feature User", Username: "featureuser", Email: "feature@x.id", Password: "x", Exp: 150}
	if err := ur.Create(ctx, u); err != nil {
		t.Fatalf("create user: %v", err)
	}

	f1 := &model.FeatureDefinition{ID: uuid.New(), FeatureKey: "chat_plus", FeatureName: "Chat Plus", RequiredLevel: 1, Category: "ai", IsActive: true}
	f2 := &model.FeatureDefinition{ID: uuid.New(), FeatureKey: "story_write", FeatureName: "Story Write", RequiredLevel: 2, Category: "content", IsActive: true}
	f3 := &model.FeatureDefinition{ID: uuid.New(), FeatureKey: "legacy", FeatureName: "Legacy", RequiredLevel: 3, Category: "special", IsActive: false}

	if err := r.CreateFeatureDefinition(ctx, f1); err != nil {
		t.Fatalf("create f1: %v", err)
	}
	if err := r.CreateFeatureDefinition(ctx, f2); err != nil {
		t.Fatalf("create f2: %v", err)
	}
	if err := r.CreateFeatureDefinition(ctx, f3); err != nil {
		t.Fatalf("create f3: %v", err)
	}
	if err := db.Model(&model.FeatureDefinition{}).Where("id = ?", f3.ID).Update("is_active", false).Error; err != nil {
		t.Fatalf("mark f3 inactive: %v", err)
	}

	all, err := r.GetAllFeatureDefinitions(ctx)
	if err != nil || len(all) != 2 {
		t.Fatalf("get all active features failed: len=%d err=%v", len(all), err)
	}

	byLevel, err := r.GetFeaturesByLevel(ctx, 1)
	if err != nil || len(byLevel) != 1 || byLevel[0].ID != f1.ID {
		t.Fatalf("get features by level failed: %v %v", byLevel, err)
	}

	upTo, err := r.GetFeaturesUpToLevel(ctx, 2)
	if err != nil || len(upTo) != 2 {
		t.Fatalf("get features up to level failed: len=%d err=%v", len(upTo), err)
	}

	if got, err := r.GetFeatureByKey(ctx, "chat_plus"); err != nil || got.ID != f1.ID {
		t.Fatalf("get by key failed: got=%+v err=%v", got, err)
	}
	if got, err := r.GetFeatureByID(ctx, f2.ID); err != nil || got.FeatureKey != "story_write" {
		t.Fatalf("get by id failed: got=%+v err=%v", got, err)
	}

	f2.FeatureName = "Story Write Updated"
	if err := r.UpdateFeatureDefinition(ctx, f2); err != nil {
		t.Fatalf("update feature failed: %v", err)
	}

	if err := r.UnlockFeature(ctx, u.ID, f1.ID); err != nil {
		t.Fatalf("unlock feature failed: %v", err)
	}
	if !r.IsFeatureUnlocked(ctx, u.ID, f1.ID) {
		t.Fatalf("expected feature unlocked")
	}
	if !r.IsFeatureUnlockedByKey(ctx, u.ID, "chat_plus") {
		t.Fatalf("expected unlocked by key")
	}

	unlocked, err := r.GetUserUnlockedFeatures(ctx, u.ID)
	if err != nil || len(unlocked) != 1 {
		t.Fatalf("get user unlocked failed: len=%d err=%v", len(unlocked), err)
	}

	newly, err := r.UnlockFeaturesForLevel(ctx, u.ID, 2)
	if err != nil || len(newly) != 1 || newly[0].ID != f2.ID {
		t.Fatalf("unlock features for level failed: len=%d err=%v", len(newly), err)
	}

	available, err := r.GetNewlyAvailableFeatures(ctx, u.ID, 1, 3)
	if err != nil || len(available) != 1 {
		t.Fatalf("get newly available features failed: len=%d err=%v", len(available), err)
	}

	countUnlocked, countTotal, err := r.GetUserFeatureStats(ctx, u.ID)
	if err != nil || countUnlocked != 2 || countTotal != 2 {
		t.Fatalf("feature stats failed: unlocked=%d total=%d err=%v", countUnlocked, countTotal, err)
	}

	upcoming, err := r.GetUpcomingFeatures(ctx, 1, 10)
	if err != nil || len(upcoming) != 1 || upcoming[0].ID != f2.ID {
		t.Fatalf("upcoming features failed: %+v err=%v", upcoming, err)
	}

	byCategory, err := r.GetFeaturesByCategory(ctx, "content")
	if err != nil || len(byCategory) != 1 || byCategory[0].ID != f2.ID {
		t.Fatalf("features by category failed: %+v err=%v", byCategory, err)
	}
}

func TestBadgeRepository_FullFlow(t *testing.T) {
	ctx := context.Background()
	db := newExtRepoDB(t)
	br := NewBadgeRepository(db)
	ur := NewUserRepository(db)

	u := &model.User{Name: "Badge User", Username: "badgeuser", Email: "badge@x.id", Password: "x", Exp: 300, CurrentStreak: 7}
	if err := ur.Create(ctx, u); err != nil {
		t.Fatalf("create user: %v", err)
	}

	if err := db.Create(&model.LevelConfig{Level: 1, MinExp: 0, BadgeName: "Pemula", BadgeIcon: "🌱"}).Error; err != nil {
		t.Fatalf("create level1: %v", err)
	}
	if err := db.Create(&model.LevelConfig{Level: 2, MinExp: 100, BadgeName: "Naik", BadgeIcon: "⭐"}).Error; err != nil {
		t.Fatalf("create level2: %v", err)
	}
	if err := db.Create(&model.LevelConfig{Level: 3, MinExp: 200, BadgeName: "Maju", BadgeIcon: "🚀"}).Error; err != nil {
		t.Fatalf("create level3: %v", err)
	}

	b1 := &model.BadgeDefinition{ID: uuid.New(), BadgeKey: "streak_7", BadgeName: "Streak 7", Category: "streak", RequirementType: model.BadgeRequirementStreak, RequirementValue: 7, IsActive: true}
	b2 := &model.BadgeDefinition{ID: uuid.New(), BadgeKey: "xp_200", BadgeName: "XP 200", Category: "level", RequirementType: model.BadgeRequirementXP, RequirementValue: 200, IsActive: true}
	b3 := &model.BadgeDefinition{ID: uuid.New(), BadgeKey: "story_1", BadgeName: "Story 1", Category: "contribution", RequirementType: model.BadgeRequirementStory, RequirementValue: 1, IsActive: true}

	if err := br.CreateBadgeDefinition(ctx, b1); err != nil {
		t.Fatalf("create badge1: %v", err)
	}
	if err := br.CreateBadgeDefinition(ctx, b2); err != nil {
		t.Fatalf("create badge2: %v", err)
	}
	if err := br.CreateBadgeDefinition(ctx, b3); err != nil {
		t.Fatalf("create badge3: %v", err)
	}

	if err := db.Create(&model.ExpHistory{UserID: u.ID, ActivityType: "chat_ai", Points: 10}).Error; err != nil {
		t.Fatalf("create exp history: %v", err)
	}
	now := time.Now()
	if err := db.Create(&model.InspiringStory{ID: uuid.New(), AuthorID: u.ID, Title: "My Story", Content: "content", Status: model.StoryStatusApproved, CreatedAt: now}).Error; err != nil {
		t.Fatalf("create inspiring story: %v", err)
	}

	all, err := br.GetAllBadgeDefinitions(ctx)
	if err != nil || len(all) != 3 {
		t.Fatalf("get all badges failed: len=%d err=%v", len(all), err)
	}

	if got, err := br.GetBadgeByKey(ctx, "streak_7"); err != nil || got.ID != b1.ID {
		t.Fatalf("get badge by key failed: got=%+v err=%v", got, err)
	}
	if got, err := br.GetBadgeByID(ctx, b2.ID); err != nil || got.BadgeKey != "xp_200" {
		t.Fatalf("get badge by id failed: got=%+v err=%v", got, err)
	}

	if byCat, err := br.GetBadgesByCategory(ctx, "streak"); err != nil || len(byCat) != 1 {
		t.Fatalf("get badges by category failed: len=%d err=%v", len(byCat), err)
	}
	if byReq, err := br.GetBadgesByRequirementType(ctx, string(model.BadgeRequirementXP)); err != nil || len(byReq) != 1 {
		t.Fatalf("get badges by requirement type failed: len=%d err=%v", len(byReq), err)
	}

	b2.BadgeName = "XP 200 Updated"
	if err := br.UpdateBadgeDefinition(ctx, b2); err != nil {
		t.Fatalf("update badge failed: %v", err)
	}

	if err := br.AwardBadge(ctx, u.ID, b1.ID); err != nil {
		t.Fatalf("award badge failed: %v", err)
	}
	if err := br.AwardBadgeByKey(ctx, u.ID, "xp_200"); err != nil {
		t.Fatalf("award badge by key failed: %v", err)
	}
	if err := br.AwardBadgeByKey(ctx, u.ID, "xp_200"); err != nil {
		t.Fatalf("award duplicate by key should still succeed: %v", err)
	}

	if !br.HasBadge(ctx, u.ID, b1.ID) || !br.HasBadgeByKey(ctx, u.ID, "xp_200") {
		t.Fatalf("expected has badge checks true")
	}

	if userBadges, err := br.GetUserBadges(ctx, u.ID); err != nil || len(userBadges) != 2 {
		t.Fatalf("get user badges failed: len=%d err=%v", len(userBadges), err)
	}
	if userStreak, err := br.GetUserBadgesByCategory(ctx, u.ID, "streak"); err != nil || len(userStreak) == 0 {
		t.Fatalf("get user badges by category failed: len=%d err=%v", len(userStreak), err)
	}

	count, err := br.GetUserBadgeCount(ctx, u.ID)
	if err != nil || count != 2 {
		t.Fatalf("badge count failed: count=%d err=%v", count, err)
	}

	progress, err := br.GetBadgeProgress(ctx, u.ID)
	if err != nil || len(progress) != 3 {
		t.Fatalf("badge progress failed: len=%d err=%v", len(progress), err)
	}

	recent, err := br.GetRecentlyEarnedBadges(ctx, u.ID, time.Now().Add(-time.Hour))
	if err != nil || len(recent) == 0 {
		t.Fatalf("recent badges failed: len=%d err=%v", len(recent), err)
	}

	usersWithBadge, err := br.GetUsersWithBadge(ctx, b1.ID, 10)
	if err != nil || len(usersWithBadge) == 0 {
		t.Fatalf("users with badge failed: len=%d err=%v", len(usersWithBadge), err)
	}

	stats, err := br.GetBadgeCategoryStats(ctx, u.ID)
	if err != nil || len(stats) == 0 {
		t.Fatalf("badge category stats failed: len=%d err=%v", len(stats), err)
	}

	display, err := br.GetDisplayBadges(ctx, u.ID, 1)
	if err != nil || len(display) != 1 {
		t.Fatalf("display badges failed: len=%d err=%v", len(display), err)
	}
}

func TestCommunityProgressRepository_FullFlow(t *testing.T) {
	ctx := context.Background()
	db := newExtRepoDB(t)
	cr := NewCommunityProgressRepository(db)
	ur := NewUserRepository(db)

	if stats, err := cr.GetCommunityStats(ctx); err != nil || stats == nil {
		t.Fatalf("get initial community stats failed: stats=%+v err=%v", stats, err)
	}

	if err := db.Create(&model.LevelConfig{Level: 1, MinExp: 0, BadgeName: "Pemula", BadgeIcon: "🌱"}).Error; err != nil {
		t.Fatalf("create level1: %v", err)
	}
	if err := db.Create(&model.LevelConfig{Level: 2, MinExp: 100, BadgeName: "Naik", BadgeIcon: "⭐"}).Error; err != nil {
		t.Fatalf("create level2: %v", err)
	}

	u1 := &model.User{Name: "U1", Username: "u1", Email: "u1@x.id", Password: "x", Exp: 120}
	u2 := &model.User{Name: "U2", Username: "u2", Email: "u2@x.id", Password: "x", Exp: 80}
	if err := ur.Create(ctx, u1); err != nil {
		t.Fatalf("create user1: %v", err)
	}
	if err := ur.Create(ctx, u2); err != nil {
		t.Fatalf("create user2: %v", err)
	}

	if err := db.Create(&model.ExpHistory{UserID: u1.ID, ActivityType: "chat_ai", Points: 10}).Error; err != nil {
		t.Fatalf("create exp history u1: %v", err)
	}
	if err := db.Create(&model.ExpHistory{UserID: u1.ID, ActivityType: "forum_comment", Points: 5}).Error; err != nil {
		t.Fatalf("create exp history u1 second: %v", err)
	}
	if err := db.Create(&model.ExpHistory{UserID: u2.ID, ActivityType: "chat_ai", Points: 10}).Error; err != nil {
		t.Fatalf("create exp history u2: %v", err)
	}

	badge := model.BadgeDefinition{ID: uuid.New(), BadgeKey: "starter", BadgeName: "Starter", Category: "activity", RequirementType: model.BadgeRequirementManual, RequirementValue: 1, IsActive: true}
	if err := db.Create(&badge).Error; err != nil {
		t.Fatalf("create badge: %v", err)
	}
	if err := db.Create(&model.UserBadge{ID: uuid.New(), UserID: u1.ID, BadgeID: badge.ID, EarnedAt: time.Now()}).Error; err != nil {
		t.Fatalf("create user badge: %v", err)
	}

	if err := db.Create(&model.InspiringStory{ID: uuid.New(), AuthorID: u1.ID, Title: "S", Content: "C", Status: model.StoryStatusApproved}).Error; err != nil {
		t.Fatalf("create inspiring story: %v", err)
	}
	cat := model.ArticleCategory{Name: "General"}
	if err := db.Create(&cat).Error; err != nil {
		t.Fatalf("create article category: %v", err)
	}
	if err := db.Create(&model.Article{Title: "Art", Content: "x", ArticleCategoryID: cat.ID, UserID: u1.ID, Status: model.ArticleStatusPublished}).Error; err != nil {
		t.Fatalf("create article: %v", err)
	}

	recalc, err := cr.RecalculateCommunityStats(ctx)
	if err != nil || recalc == nil {
		t.Fatalf("recalculate stats failed: stats=%+v err=%v", recalc, err)
	}
	if recalc.NewMembers < 2 || recalc.ActiveMembers < 2 {
		t.Fatalf("unexpected recalc stats: %+v", recalc)
	}

	if err := cr.UpdateCommunityStats(ctx, recalc); err != nil {
		t.Fatalf("update community stats create: %v", err)
	}
	recalc.ActiveMembers = 99
	if err := cr.UpdateCommunityStats(ctx, recalc); err != nil {
		t.Fatalf("update community stats update: %v", err)
	}

	stored, err := cr.GetCommunityStats(ctx)
	if err != nil || stored.ActiveMembers != 99 {
		t.Fatalf("get updated community stats failed: stats=%+v err=%v", stored, err)
	}

	usersRange, err := cr.GetUsersInLevelRange(ctx, 1, 2, 10)
	if err != nil {
		t.Fatalf("users in level range failed: len=%d err=%v", len(usersRange), err)
	}

	levelTop, err := cr.GetTopUsersInLevel(ctx, 1, 10)
	if err != nil || len(levelTop) < 1 {
		t.Fatalf("top users in level failed: len=%d err=%v", len(levelTop), err)
	}

	hof := &model.MonthlyHallOfFame{ID: uuid.New(), UserID: u1.ID, Level: 1, Month: int(time.Now().Month()), Year: time.Now().Year(), Rank: 1, MonthlyXP: 100, Message: "Great"}
	if err := cr.CreateHallOfFameEntry(ctx, hof); err != nil {
		t.Fatalf("create hall of fame entry: %v", err)
	}

	hofEntries, err := cr.GetHallOfFame(ctx, hof.Month, hof.Year, "")
	if err != nil || len(hofEntries) != 1 {
		t.Fatalf("get hall of fame failed: len=%d err=%v", len(hofEntries), err)
	}

	categories := cr.GetHallOfFameCategories(ctx)
	if len(categories) == 0 {
		t.Fatalf("expected hall of fame categories")
	}

	rank, total, err := cr.GetUserRankInLevel(ctx, u1.ID, 1)
	if err != nil || rank < 1 {
		t.Fatalf("get user rank in level failed: rank=%d total=%d err=%v", rank, total, err)
	}

	weeklyXP, weeklyActivities, err := cr.GetWeeklyProgress(ctx, u1.ID)
	if err != nil || weeklyActivities < 2 || weeklyXP < 0 {
		t.Fatalf("weekly progress failed: xp=%d activities=%d err=%v", weeklyXP, weeklyActivities, err)
	}

	monthlyXP, monthlyActivities, err := cr.GetMonthlyProgress(ctx, u1.ID)
	if err != nil || monthlyActivities < 2 || monthlyXP < 0 {
		t.Fatalf("monthly progress failed: xp=%d activities=%d err=%v", monthlyXP, monthlyActivities, err)
	}

	activityMap, err := cr.GetUserActivityTypes(ctx, u1.ID, time.Now().AddDate(0, 0, -7))
	if err != nil || len(activityMap) == 0 {
		t.Fatalf("activity types failed: map=%+v err=%v", activityMap, err)
	}
}

func TestBadgeRepository_CalculateProgress_AllRequirementTypes(t *testing.T) {
	ctx := context.Background()
	db := newExtRepoDB(t)
	br := NewBadgeRepository(db)
	ur := NewUserRepository(db)

	u := &model.User{Name: "Progress User", Username: "progressuser", Email: "progress@x.id", Password: "x", Exp: 250, CurrentStreak: 5}
	if err := ur.Create(ctx, u); err != nil {
		t.Fatalf("create user: %v", err)
	}

	if err := db.Create(&model.LevelConfig{Level: 1, MinExp: 0, BadgeName: "L1", BadgeIcon: "1"}).Error; err != nil {
		t.Fatalf("create level1: %v", err)
	}
	if err := db.Create(&model.LevelConfig{Level: 2, MinExp: 100, BadgeName: "L2", BadgeIcon: "2"}).Error; err != nil {
		t.Fatalf("create level2: %v", err)
	}

	if err := db.Create(&model.ExpHistory{UserID: u.ID, ActivityType: "chat_ai", Points: 10}).Error; err != nil {
		t.Fatalf("create exp history one: %v", err)
	}
	if err := db.Create(&model.ExpHistory{UserID: u.ID, ActivityType: "forum_comment", Points: 10}).Error; err != nil {
		t.Fatalf("create exp history two: %v", err)
	}
	if err := db.Create(&model.InspiringStory{ID: uuid.New(), AuthorID: u.ID, Title: "Story", Content: "ok", Status: model.StoryStatusApproved, CreatedAt: time.Now()}).Error; err != nil {
		t.Fatalf("create approved story: %v", err)
	}

	cases := []struct {
		name        string
		requirement model.BadgeRequirementType
		want        int
	}{
		{name: "streak", requirement: model.BadgeRequirementStreak, want: 5},
		{name: "activity", requirement: model.BadgeRequirementActivityCount, want: 2},
		{name: "level", requirement: model.BadgeRequirementLevel, want: 2},
		{name: "story", requirement: model.BadgeRequirementStory, want: 1},
		{name: "xp", requirement: model.BadgeRequirementXP, want: 250},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			current, target := br.calculateProgress(ctx, u.ID, model.BadgeDefinition{RequirementType: tc.requirement, RequirementValue: 9})
			if current != tc.want || target != 9 {
				t.Fatalf("unexpected progress for %s: current=%d target=%d", tc.name, current, target)
			}
		})
	}

	current, target := br.calculateProgress(ctx, u.ID, model.BadgeDefinition{RequirementType: model.BadgeRequirementType("unknown"), RequirementValue: 3})
	if current != 0 || target != 3 {
		t.Fatalf("unexpected default branch progress: current=%d target=%d", current, target)
	}
}
