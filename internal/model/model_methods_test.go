package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/lib/pq"
)

func TestTableNamesBasic(t *testing.T) {
	checks := map[string]string{
		(User{}).TableName():                "users",
		(Forum{}).TableName():               "forums",
		(ForumPost{}).TableName():           "forum_posts",
		(ForumPostVote{}).TableName():       "forum_post_votes",
		(ForumPostReport{}).TableName():     "forum_post_reports",
		(Article{}).TableName():             "articles",
		(ArticleCategory{}).TableName():     "article_categories",
		(ChatSession{}).TableName():         "chat_sessions",
		(ChatMessage{}).TableName():         "chat_messages",
		(BreathingSession{}).TableName():    "breathing_sessions",
		(BreathingTechnique{}).TableName():  "breathing_techniques",
		(BreathingPreference{}).TableName(): "breathing_preferences",
		(BreathingFavorite{}).TableName():   "breathing_favorites",
		(DailyTask{}).TableName():           "daily_tasks",
		(Journal{}).TableName():             "journals",
		(JournalSettings{}).TableName():     "journal_settings",
		(JournalAIAccessLog{}).TableName():  "journal_ai_access_logs",
	}

	for got, want := range checks {
		if got != want {
			t.Fatalf("unexpected table name: got %s want %s", got, want)
		}
	}
}

func TestAdditionalTableNamesAndBeforeCreate(t *testing.T) {
	tableNames := []string{
		(UserActivity{}).TableName(),
		(Appeal{}).TableName(),
		(ContentFlag{}).TableName(),
		(Notification{}).TableName(),
		(FeatureDefinition{}).TableName(),
		(UserFeatureUnlock{}).TableName(),
		(ChatFolder{}).TableName(),
		(UserReport{}).TableName(),
		(UserBlock{}).TableName(),
		(UserStrike{}).TableName(),
		(StoryCategory{}).TableName(),
		(InspiringStory{}).TableName(),
		(StoryTag{}).TableName(),
		(StoryHeart{}).TableName(),
		(StoryComment{}).TableName(),
		(StoryCommentHeart{}).TableName(),
		(StoryCategoryRelation{}).TableName(),
		(MonthlyHallOfFame{}).TableName(),
		(CommunityStats{}).TableName(),
		(SongCategory{}).TableName(),
		(Song{}).TableName(),
		(Playlist{}).TableName(),
		(PlaylistItem{}).TableName(),
		(ExpHistory{}).TableName(),
		(LevelConfig{}).TableName(),
		(CrisisKeyword{}).TableName(),
		(ModeratorAction{}).TableName(),
		(ForumCategory{}).TableName(),
		(ForumLike{}).TableName(),
	}

	for _, tn := range tableNames {
		if tn == "" {
			t.Fatal("table name should not be empty")
		}
	}

	if (BadgeDefinition{}).TableName() != "badge_definitions" {
		t.Fatalf("unexpected badge definitions table name")
	}
	if (UserBadge{}).TableName() != "user_badges" {
		t.Fatalf("unexpected user badges table name")
	}
	if (UserMood{}).TableName() != "user_moods" {
		t.Fatalf("unexpected user moods table name")
	}

	user := &User{Name: "Test User"}
	if err := user.BeforeCreate(nil); err != nil {
		t.Fatalf("user before create failed: %v", err)
	}
	if user.Username == "" {
		t.Fatal("expected generated username on before create")
	}

	articleCategory := &ArticleCategory{Name: "Mind"}
	if err := articleCategory.BeforeCreate(nil); err != nil || articleCategory.Slug == "" {
		t.Fatalf("article category before create failed: err=%v slug=%s", err, articleCategory.Slug)
	}

	article := &Article{Title: "A title"}
	if err := article.BeforeCreate(nil); err != nil || article.Slug == "" {
		t.Fatalf("article before create failed: err=%v slug=%s", err, article.Slug)
	}

	forumCategory := &ForumCategory{Name: "General"}
	if err := forumCategory.BeforeCreate(nil); err != nil || forumCategory.Slug == "" {
		t.Fatalf("forum category before create failed: err=%v slug=%s", err, forumCategory.Slug)
	}

	forum := &Forum{Title: "Forum title"}
	if err := forum.BeforeCreate(nil); err != nil || forum.Slug == "" {
		t.Fatalf("forum before create failed: err=%v slug=%s", err, forum.Slug)
	}

	songCategory := &SongCategory{Name: "Focus"}
	if err := songCategory.BeforeCreate(nil); err != nil || songCategory.Slug == "" {
		t.Fatalf("song category before create failed: err=%v slug=%s", err, songCategory.Slug)
	}

	song := &Song{Title: "Song title"}
	if err := song.BeforeCreate(nil); err != nil || song.Slug == "" {
		t.Fatalf("song before create failed: err=%v slug=%s", err, song.Slug)
	}
}

func TestUserRoleAndAccessHelpers(t *testing.T) {
	admin := &User{Role: RoleAdmin}
	moderator := &User{Role: RoleModerator}
	member := &User{Role: RoleMember}

	if !admin.IsAdmin() || !admin.CanModerate() || admin.IsMember() {
		t.Fatal("admin role helpers invalid")
	}
	if !moderator.IsModerator() || !moderator.CanModerate() || moderator.IsAdmin() {
		t.Fatal("moderator role helpers invalid")
	}
	if !member.IsMember() || member.CanModerate() {
		t.Fatal("member role helpers invalid")
	}

	if !member.CanAccess() {
		t.Fatal("member should access by default")
	}
	member.IsBlocked = true
	if member.CanAccess() {
		t.Fatal("blocked member should not access")
	}

	future := time.Now().Add(2 * time.Hour)
	member.IsBlocked = false
	member.SuspensionEnd = &future
	if !member.IsSuspended() || member.CanAccess() {
		t.Fatal("suspended member should not access")
	}
}

func TestForumPostVoteAndReportHelpers(t *testing.T) {
	v := &ForumPostVote{VoteType: VoteTypeUpvote}
	if !v.IsUpvote() || v.IsDownvote() {
		t.Fatal("upvote helper mismatch")
	}
	v.VoteType = VoteTypeDownvote
	if !v.IsDownvote() || v.IsUpvote() {
		t.Fatal("downvote helper mismatch")
	}

	r := &ForumPostReport{Status: PostReportStatusPending}
	if !r.IsPending() {
		t.Fatal("report should be pending")
	}
	if !IsValidPostReportReason("spam") || IsValidPostReportReason("invalid") {
		t.Fatal("post report reason validator mismatch")
	}
	if len(ValidPostReportReasons()) == 0 {
		t.Fatal("valid post report reasons should not be empty")
	}
}

func TestForumAndArticleHelpers(t *testing.T) {
	f := &Forum{TriggerWarnings: TriggerWarnings{"trauma"}}
	if !f.HasTriggerWarnings() {
		t.Fatal("forum trigger warnings should be detected")
	}

	p := &ForumPost{UpvotesCount: 3, DownvotesCount: 10}
	if p.CalculateNetVotes() != -7 {
		t.Fatalf("expected net votes -7, got %d", p.CalculateNetVotes())
	}
	if !p.ShouldAutoHide() {
		t.Fatal("post with net votes < -5 should auto hide")
	}

	a := &Article{Status: ArticleStatusPublished, IsUserGenerated: true, ModerationStatus: ArticleModerationApproved}
	if !a.IsPublic() {
		t.Fatal("approved published UGC article should be public")
	}
	a.ModerationStatus = ArticleModerationFlagged
	if !a.NeedsModeration() {
		t.Fatal("flagged UGC article should need moderation")
	}
	a.TriggerWarnings = TriggerWarnings{"self_harm"}
	if !a.HasTriggerWarnings() {
		t.Fatal("article trigger warnings should be detected")
	}
}

func TestChatAndBreathingHelpers(t *testing.T) {
	session := &ChatSession{Messages: []ChatMessage{{IsPinned: true}, {IsPinned: false}, {IsPinned: true}}}
	if len(session.GetPinnedMessages()) != 2 {
		t.Fatalf("expected 2 pinned messages, got %d", len(session.GetPinnedMessages()))
	}

	msg := &ChatMessage{Role: ChatRoleAI}
	if !msg.IsAI() || msg.IsUser() {
		t.Fatal("chat role AI helper mismatch")
	}
	msg.Role = ChatRoleUser
	if !msg.IsUser() || msg.IsAI() {
		t.Fatal("chat role user helper mismatch")
	}

	tech := &BreathingTechnique{InhaleDuration: 4, InhaleHoldDuration: 2, ExhaleDuration: 6, ExhaleHoldDuration: 0}
	if tech.GetTotalCycleDuration() != 12 {
		t.Fatalf("expected total cycle duration 12, got %d", tech.GetTotalCycleDuration())
	}
	if tech.GetCyclesForDuration(120) != 10 {
		t.Fatalf("expected 10 cycles, got %d", tech.GetCyclesForDuration(120))
	}

	zero := &BreathingTechnique{}
	if zero.GetCyclesForDuration(120) != 0 {
		t.Fatalf("expected 0 cycles for zero-duration technique, got %d", zero.GetCyclesForDuration(120))
	}
}

func TestDailyTaskAndJournalHelpers(t *testing.T) {
	if GetTaskConfig(TaskTypeDailyLogin) == nil {
		t.Fatal("expected daily login task config")
	}
	if GetTotalPossibleXP() <= 0 {
		t.Fatal("total possible XP should be > 0")
	}

	task := &DailyTask{TaskType: TaskTypeChatAI, CurrentCount: 2, TargetCount: 3}
	task.PopulateTaskInfo()
	if task.TaskName == "" {
		t.Fatal("task name should be populated")
	}
	if task.Progress() != 66 {
		t.Fatalf("expected progress 66, got %d", task.Progress())
	}

	task.TargetCount = 0
	if task.Progress() != 0 {
		t.Fatalf("expected progress 0 for zero target, got %d", task.Progress())
	}

	task.TargetCount = 1
	task.CurrentCount = 5
	if task.Progress() != 100 {
		t.Fatalf("expected progress capped at 100, got %d", task.Progress())
	}

	mood := &UserMood{Mood: MoodHappy}
	j := &Journal{Mood: mood, Tags: pq.StringArray{"a", "b"}}
	if j.GetMoodLabel() != "happy" || j.GetMoodEmoji() != "😊" {
		t.Fatalf("unexpected journal mood mapping: %s %s", j.GetMoodLabel(), j.GetMoodEmoji())
	}

	j.Mood = nil
	if j.GetMoodLabel() != "" || j.GetMoodEmoji() != "" {
		t.Fatalf("expected empty mood label/emoji for nil mood, got %q %q", j.GetMoodLabel(), j.GetMoodEmoji())
	}
}

func TestTriggerWarningsJSONScanValue(t *testing.T) {
	var nilTW TriggerWarnings
	vNil, err := nilTW.Value()
	if err != nil {
		t.Fatalf("unexpected nil value error: %v", err)
	}
	if vNil != nil {
		t.Fatalf("expected nil driver value for nil warnings, got %#v", vNil)
	}

	tw := TriggerWarnings{"a", "b"}
	v, err := tw.Value()
	if err != nil {
		t.Fatalf("unexpected value error: %v", err)
	}
	if v == nil {
		t.Fatal("value should not be nil")
	}

	var parsed TriggerWarnings
	raw, _ := json.Marshal([]string{"x", "y"})
	if err := parsed.Scan(raw); err != nil {
		t.Fatalf("unexpected scan error: %v", err)
	}
	if len(parsed) != 2 || parsed[0] != "x" {
		t.Fatalf("unexpected parsed warnings: %+v", parsed)
	}

	if err := parsed.Scan(nil); err != nil {
		t.Fatalf("scan nil should not error: %v", err)
	}
	if parsed != nil {
		t.Fatal("expected parsed warnings to become nil on nil scan")
	}

	parsed = TriggerWarnings{"existing"}
	if err := parsed.Scan("not-bytes"); err != nil {
		t.Fatalf("scan non-byte should not error: %v", err)
	}
	if len(parsed) != 1 || parsed[0] != "existing" {
		t.Fatalf("expected scan non-byte to keep previous value, got %+v", parsed)
	}
}
