package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/Alfian57/ruang-tenang-api/pkg/timeutil"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type mockDailyTaskRepo struct {
	createDailyCalls int
	incrementCalls   int
	getSummaryCalls  int
	historyCalls     int

	createDailyUserID uint
	incrementUserID   uint
	incrementTaskType model.DailyTaskType

	summary *model.DailyTaskSummary
	history []model.DailyTaskSummary
	total   int64

	markCompletedErr error
	claimRewardErr   error
	taskByID         *model.DailyTask
	taskByIDErr      error
	tasksForDate     []model.DailyTask
	tasksForDateErr  error

	streakVal  int
	isNewDay   bool
	streakErr  error
	claimCalls int
}

func (m *mockDailyTaskRepo) CreateTask(_ context.Context, _ *model.DailyTask) error { return nil }
func (m *mockDailyTaskRepo) GetTaskByID(_ context.Context, _ uint) (*model.DailyTask, error) {
	if m.taskByIDErr != nil {
		return nil, m.taskByIDErr
	}
	return m.taskByID, nil
}
func (m *mockDailyTaskRepo) UpdateTask(_ context.Context, _ *model.DailyTask) error { return nil }
func (m *mockDailyTaskRepo) GetUserTasksForDate(_ context.Context, _ uint, _ time.Time) ([]model.DailyTask, error) {
	if m.tasksForDateErr != nil {
		return nil, m.tasksForDateErr
	}
	return m.tasksForDate, nil
}
func (m *mockDailyTaskRepo) GetUserTask(_ context.Context, _ uint, _ model.DailyTaskType, _ time.Time) (*model.DailyTask, error) {
	return nil, nil
}
func (m *mockDailyTaskRepo) CreateDailyTasksForUser(_ context.Context, userID uint, _ time.Time) error {
	m.createDailyCalls++
	m.createDailyUserID = userID
	return nil
}
func (m *mockDailyTaskRepo) IncrementTaskProgress(_ context.Context, userID uint, taskType model.DailyTaskType, _ time.Time) error {
	m.incrementCalls++
	m.incrementUserID = userID
	m.incrementTaskType = taskType
	return nil
}
func (m *mockDailyTaskRepo) MarkTaskCompleted(_ context.Context, _ uint, _ model.DailyTaskType, _ time.Time) error {
	return m.markCompletedErr
}
func (m *mockDailyTaskRepo) ClaimTaskReward(_ context.Context, _ uint, _ uint) error {
	m.claimCalls++
	return m.claimRewardErr
}
func (m *mockDailyTaskRepo) UpdateUserLoginStreak(_ context.Context, _ uint, _ time.Time) (int, bool, error) {
	if m.streakErr != nil {
		return 0, false, m.streakErr
	}
	return m.streakVal, m.isNewDay, nil
}
func (m *mockDailyTaskRepo) GetDailyTaskSummary(_ context.Context, _ uint, _ time.Time) (*model.DailyTaskSummary, error) {
	m.getSummaryCalls++
	return m.summary, nil
}
func (m *mockDailyTaskRepo) GetTaskHistory(_ context.Context, _ uint, _, _ int) ([]model.DailyTaskSummary, int64, error) {
	m.historyCalls++
	return m.history, m.total, nil
}

func TestDailyTaskService_InitializeAndProgress(t *testing.T) {
	repo := &mockDailyTaskRepo{}
	svc := NewDailyTaskService(repo, nil)
	ctx := context.Background()

	if err := svc.InitializeDailyTasks(ctx, 42); err != nil {
		t.Fatalf("unexpected init error: %v", err)
	}
	if err := svc.UpdateTaskProgress(ctx, 42, model.TaskTypeBreathing); err != nil {
		t.Fatalf("unexpected progress error: %v", err)
	}

	if repo.createDailyCalls != 2 {
		t.Fatalf("expected create daily calls twice (init + progress), got %d", repo.createDailyCalls)
	}
	if repo.createDailyUserID != 42 {
		t.Fatalf("unexpected create daily user id: %d", repo.createDailyUserID)
	}
	if repo.incrementCalls != 1 || repo.incrementUserID != 42 || repo.incrementTaskType != model.TaskTypeBreathing {
		t.Fatalf("unexpected increment call state: calls=%d userID=%d taskType=%s", repo.incrementCalls, repo.incrementUserID, repo.incrementTaskType)
	}
}

func TestDailyTaskService_GetTodayTasks(t *testing.T) {
	summary := &model.DailyTaskSummary{TotalTasks: 3, CompletedTasks: 1}
	repo := &mockDailyTaskRepo{summary: summary}
	svc := NewDailyTaskService(repo, nil)

	got, err := svc.GetTodayTasks(context.Background(), 7)
	if err != nil {
		t.Fatalf("unexpected get today tasks error: %v", err)
	}
	if got == nil || got.TotalTasks != 3 || got.CompletedTasks != 1 {
		t.Fatalf("unexpected summary: %+v", got)
	}
	if repo.createDailyCalls != 1 {
		t.Fatalf("expected initialize call before summary, got %d", repo.createDailyCalls)
	}
	if repo.getSummaryCalls != 1 {
		t.Fatalf("expected one summary call, got %d", repo.getSummaryCalls)
	}
}

func TestDailyTaskService_GetTaskHistoryPagination(t *testing.T) {
	repo := &mockDailyTaskRepo{total: 65}
	svc := NewDailyTaskService(repo, nil)

	result, err := svc.GetTaskHistory(context.Background(), 9, 0, 100)
	if err != nil {
		t.Fatalf("unexpected history error: %v", err)
	}

	if result.Page != 1 {
		t.Fatalf("expected page fallback to 1, got %d", result.Page)
	}
	if result.PageSize != 30 {
		t.Fatalf("expected page size capped to 30, got %d", result.PageSize)
	}
	if result.TotalPages != 3 {
		t.Fatalf("expected total pages 3, got %d", result.TotalPages)
	}
	if repo.historyCalls != 1 {
		t.Fatalf("expected one history call, got %d", repo.historyCalls)
	}
}

func TestDailyTaskService_HelperFunctions(t *testing.T) {
	if got := calculateStreakBonus(1); got != 0 {
		t.Fatalf("expected bonus 0, got %d", got)
	}
	if got := calculateStreakBonus(4); got != 15 {
		t.Fatalf("expected bonus 15, got %d", got)
	}
	if got := calculateStreakBonus(50); got != 50 {
		t.Fatalf("expected capped bonus 50, got %d", got)
	}
	if formatStreak(7) != "7 hari" {
		t.Fatalf("unexpected streak format: %s", formatStreak(7))
	}

	if !timeutil.IsSameDay(timeutil.Now(), timeutil.Now()) {
		t.Fatal("sanity check failed for same day")
	}
}

func newDailyTaskUserRepo(t *testing.T, exp int64) *repository.UserRepository {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("migrate user: %v", err)
	}
	if err := db.Create(&model.User{Name: "Task User", Username: "taskuser", Email: "task@x.id", Password: "x", Exp: exp}).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return repository.NewUserRepository(db)
}

func TestDailyTaskService_ProcessDailyLoginBranches(t *testing.T) {
	ctx := context.Background()
	userRepo := newDailyTaskUserRepo(t, 100)

	repoNewDay := &mockDailyTaskRepo{streakVal: 1, isNewDay: true}
	svc := NewDailyTaskService(repoNewDay, userRepo)

	res, err := svc.ProcessDailyLogin(ctx, 1)
	if err != nil {
		t.Fatalf("process daily login new day failed: %v", err)
	}
	if !res.IsNewDay || !res.TaskCompleted || res.TotalXPFromLogin == 0 {
		t.Fatalf("unexpected new day login result: %+v", res)
	}

	repoSameDay := &mockDailyTaskRepo{streakVal: 3, isNewDay: false}
	svc2 := NewDailyTaskService(repoSameDay, userRepo)
	res2, err := svc2.ProcessDailyLogin(ctx, 1)
	if err != nil {
		t.Fatalf("process daily login same day failed: %v", err)
	}
	if res2.IsNewDay || res2.TaskCompleted {
		t.Fatalf("expected same-day no completion, got %+v", res2)
	}
}

func TestDailyTaskService_ClaimTaskRewardBranches(t *testing.T) {
	ctx := context.Background()
	userRepo := newDailyTaskUserRepo(t, 120)

	repoNotFound := &mockDailyTaskRepo{taskByIDErr: gorm.ErrRecordNotFound}
	svc := NewDailyTaskService(repoNotFound, userRepo)
	if _, err := svc.ClaimTaskReward(ctx, 1, 1); !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}

	repoWrongOwner := &mockDailyTaskRepo{taskByID: &model.DailyTask{ID: 1, UserID: 2, TaskType: model.TaskTypeDailyLogin}}
	svc = NewDailyTaskService(repoWrongOwner, userRepo)
	if _, err := svc.ClaimTaskReward(ctx, 1, 1); !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound for wrong owner, got %v", err)
	}

	repoIncomplete := &mockDailyTaskRepo{taskByID: &model.DailyTask{ID: 1, UserID: 1, TaskType: model.TaskTypeDailyLogin, IsCompleted: false}}
	svc = NewDailyTaskService(repoIncomplete, userRepo)
	if _, err := svc.ClaimTaskReward(ctx, 1, 1); !errors.Is(err, ErrTaskNotCompleted) {
		t.Fatalf("expected ErrTaskNotCompleted, got %v", err)
	}

	repoClaimed := &mockDailyTaskRepo{taskByID: &model.DailyTask{ID: 1, UserID: 1, TaskType: model.TaskTypeDailyLogin, IsCompleted: true, IsClaimed: true}}
	svc = NewDailyTaskService(repoClaimed, userRepo)
	if _, err := svc.ClaimTaskReward(ctx, 1, 1); !errors.Is(err, ErrTaskAlreadyClaimed) {
		t.Fatalf("expected ErrTaskAlreadyClaimed, got %v", err)
	}

	repoSuccess := &mockDailyTaskRepo{taskByID: &model.DailyTask{ID: 1, UserID: 1, TaskType: model.TaskTypeDailyLogin, IsCompleted: true, IsClaimed: false}}
	svc = NewDailyTaskService(repoSuccess, userRepo)
	claim, err := svc.ClaimTaskReward(ctx, 1, 1)
	if err != nil {
		t.Fatalf("expected claim success, got %v", err)
	}
	if claim == nil || claim.TaskID != 1 || claim.TotalXP != 120 {
		t.Fatalf("unexpected claim result: %+v", claim)
	}
	if repoSuccess.claimCalls != 1 {
		t.Fatalf("expected one claim call, got %d", repoSuccess.claimCalls)
	}
}

func TestDailyTaskService_ClaimAllRewards(t *testing.T) {
	ctx := context.Background()
	userRepo := newDailyTaskUserRepo(t, 200)

	repo := &mockDailyTaskRepo{
		tasksForDate: []model.DailyTask{
			{ID: 1, UserID: 1, TaskType: model.TaskTypeDailyLogin, IsCompleted: true, IsClaimed: false},
			{ID: 2, UserID: 1, TaskType: model.TaskTypeBreathing, IsCompleted: true, IsClaimed: false},
			{ID: 3, UserID: 1, TaskType: model.TaskTypeReadArticle, IsCompleted: false, IsClaimed: false},
		},
		taskByID: &model.DailyTask{ID: 1, UserID: 1, TaskType: model.TaskTypeDailyLogin, IsCompleted: true, IsClaimed: false},
	}

	svc := NewDailyTaskService(repo, userRepo)
	result, err := svc.ClaimAllRewards(ctx, 1)
	if err != nil {
		t.Fatalf("claim all rewards failed: %v", err)
	}
	if result.TotalClaimed == 0 || len(result.ClaimedTasks) == 0 {
		t.Fatalf("expected claimed tasks, got %+v", result)
	}
}
