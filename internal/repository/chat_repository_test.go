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

func setupChatRepoDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	createFolders := `CREATE TABLE chat_folders (
		id INTEGER PRIMARY KEY,
		uuid TEXT,
		user_id INTEGER,
		name TEXT,
		color TEXT,
		icon TEXT,
		position INTEGER,
		created_at DATETIME,
		updated_at DATETIME
	)`
	createSessions := `CREATE TABLE chat_sessions (
		id INTEGER PRIMARY KEY,
		uuid TEXT,
		user_id INTEGER,
		folder_id INTEGER,
		title TEXT,
		summary TEXT,
		summary_generated_at DATETIME,
		is_favorite BOOLEAN,
		is_trash BOOLEAN,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`
	createMessages := `CREATE TABLE chat_messages (
		id INTEGER PRIMARY KEY,
		uuid TEXT,
		chat_session_id INTEGER,
		role TEXT,
		content TEXT,
		type TEXT,
		is_liked BOOLEAN,
		is_disliked BOOLEAN,
		is_pinned BOOLEAN,
		created_at DATETIME,
		updated_at DATETIME
	)`

	for _, query := range []string{createFolders, createSessions, createMessages} {
		if err := db.Exec(query).Error; err != nil {
			t.Fatalf("create table failed: %v", err)
		}
	}

	now := time.Now()
	if err := db.Exec(`INSERT INTO chat_folders (id, uuid, user_id, name, position, created_at, updated_at) VALUES (1, '11111111-1111-1111-1111-111111111111', 7, 'Folder A', 0, ?, ?)`, now, now).Error; err != nil {
		t.Fatalf("seed folder: %v", err)
	}
	if err := db.Exec(`INSERT INTO chat_sessions (id, uuid, user_id, folder_id, title, is_favorite, is_trash, created_at, updated_at) VALUES (1, '22222222-2222-2222-2222-222222222222', 7, 1, 'Session 1', 0, 0, ?, ?)`, now, now).Error; err != nil {
		t.Fatalf("seed session 1: %v", err)
	}
	if err := db.Exec(`INSERT INTO chat_sessions (id, uuid, user_id, folder_id, title, is_favorite, is_trash, created_at, updated_at) VALUES (2, '33333333-3333-3333-3333-333333333333', 7, 1, 'Session 2', 1, 0, ?, ?)`, now, now).Error; err != nil {
		t.Fatalf("seed session 2: %v", err)
	}
	if err := db.Exec(`INSERT INTO chat_messages (id, uuid, chat_session_id, role, content, is_pinned, created_at, updated_at) VALUES (1, '44444444-4444-4444-4444-444444444444', 1, 'user', 'hello', 0, ?, ?)`, now, now).Error; err != nil {
		t.Fatalf("seed message 1: %v", err)
	}
	if err := db.Exec(`INSERT INTO chat_messages (id, uuid, chat_session_id, role, content, is_pinned, created_at, updated_at) VALUES (2, '55555555-5555-5555-5555-555555555555', 1, 'ai', 'hi', 1, ?, ?)`, now, now).Error; err != nil {
		t.Fatalf("seed message 2: %v", err)
	}

	return db
}

func TestChatFolderRepository_BasicOps(t *testing.T) {
	db := setupChatRepoDB(t)
	repo := NewChatFolderRepository(db)
	ctx := context.Background()

	folders, err := repo.FindByUserID(ctx, 7)
	if err != nil || len(folders) == 0 {
		t.Fatalf("find by user failed: %v", err)
	}

	f, err := repo.FindByID(ctx, 1)
	if err != nil || f.ID != 1 {
		t.Fatalf("find by id failed: %v", err)
	}
	if _, err := repo.FindByID(ctx, 999999); err == nil {
		t.Fatal("expected folder FindByID missing error")
	}

	withSessions, err := repo.FindByIDWithSessions(ctx, 1)
	if err != nil {
		t.Fatalf("find by id with sessions failed: %v", err)
	}
	if len(withSessions.Sessions) == 0 {
		t.Fatal("expected preloaded sessions")
	}
	if _, err := repo.FindByIDWithSessions(ctx, 999999); err == nil {
		t.Fatal("expected folder FindByIDWithSessions missing error")
	}

	count, err := repo.CountSessionsInFolder(ctx, 1)
	if err != nil || count < 1 {
		t.Fatalf("count sessions failed: count=%d err=%v", count, err)
	}

	maxPos, err := repo.GetMaxPosition(ctx, 7)
	if err != nil || maxPos < 0 {
		t.Fatalf("get max position failed: %d %v", maxPos, err)
	}

	newFolder := &model.ChatFolder{UUID: uuid.New(), UserID: 7, Name: "Folder B", Position: 2}
	if err := repo.Create(ctx, newFolder); err != nil {
		t.Fatalf("create folder failed: %v", err)
	}
	newFolder.Name = "Folder B Updated"
	if err := repo.Update(ctx, newFolder); err != nil {
		t.Fatalf("update folder failed: %v", err)
	}

	if err := repo.ReorderFolders(ctx, 7, []uint{1, newFolder.ID}); err != nil {
		t.Fatalf("reorder folders failed: %v", err)
	}

	if err := repo.Delete(ctx, newFolder.ID); err != nil {
		t.Fatalf("delete folder failed: %v", err)
	}
}

func TestChatSessionRepository_Ops(t *testing.T) {
	db := setupChatRepoDB(t)
	repo := NewChatSessionRepository(db)
	ctx := context.Background()

	sessions, total, err := repo.FindByUserID(ctx, 7, "", "", nil, 1, 10)
	if err != nil || total == 0 || len(sessions) == 0 {
		t.Fatalf("find by user failed: total=%d len=%d err=%v", total, len(sessions), err)
	}

	fav, _, err := repo.FindByUserID(ctx, 7, "favorites", "", nil, 1, 10)
	if err != nil || len(fav) == 0 {
		t.Fatalf("favorites filter failed: %v", err)
	}

	_, _, err = repo.FindByUserID(ctx, 7, "", "Session", nil, 1, 10)
	if err == nil {
		t.Fatal("expected sqlite ILIKE error for search branch")
	}

	grouped, err := repo.FindByUserIDGroupedByFolder(ctx, 7, "")
	if err != nil {
		t.Fatalf("find grouped by folder failed: %v", err)
	}
	if len(grouped) == 0 {
		t.Fatal("expected grouped sessions")
	}

	groupedFav, err := repo.FindByUserIDGroupedByFolder(ctx, 7, "favorites")
	if err != nil {
		t.Fatalf("find grouped favorites failed: %v", err)
	}
	if len(groupedFav) == 0 {
		t.Fatal("expected grouped favorite sessions")
	}

	groupedTrash, err := repo.FindByUserIDGroupedByFolder(ctx, 7, "trash")
	if err != nil {
		t.Fatalf("find grouped trash failed: %v", err)
	}

	byID, err := repo.FindByID(ctx, 1)
	if err != nil || byID.ID != 1 {
		t.Fatalf("find by id failed: %v", err)
	}
	if _, err := repo.FindByID(ctx, 999999); err == nil {
		t.Fatal("expected session FindByID missing error")
	}

	sessionUUID, _ := uuid.Parse("22222222-2222-2222-2222-222222222222")
	byUUID, err := repo.FindByUUID(ctx, sessionUUID)
	if err != nil || byUUID.ID != 1 {
		t.Fatalf("find by uuid failed: %v", err)
	}
	if _, err := repo.FindByUUID(ctx, uuid.New()); err == nil {
		t.Fatal("expected FindByUUID missing error")
	}

	withMessages, err := repo.FindByIDWithMessages(ctx, 1)
	if err != nil || len(withMessages.Messages) == 0 {
		t.Fatalf("find with messages failed: %v", err)
	}
	if _, err := repo.FindByIDWithMessages(ctx, 999999); err == nil {
		t.Fatal("expected FindByIDWithMessages missing error")
	}

	pinnedOnly, err := repo.FindByIDWithPinnedMessages(ctx, 1)
	if err != nil {
		t.Fatalf("find with pinned failed: %v", err)
	}
	if len(pinnedOnly.Messages) != 1 {
		t.Fatalf("expected one pinned message, got %d", len(pinnedOnly.Messages))
	}
	if _, err := repo.FindByIDWithPinnedMessages(ctx, 999999); err == nil {
		t.Fatal("expected FindByIDWithPinnedMessages missing error")
	}

	newSession := &model.ChatSession{UUID: uuid.New(), UserID: 7, Title: "Session New"}
	if err := repo.Create(ctx, newSession); err != nil {
		t.Fatalf("create session failed: %v", err)
	}
	newSession.Title = "Session Updated"
	if err := repo.Update(ctx, newSession); err != nil {
		t.Fatalf("update session failed: %v", err)
	}

	if err := repo.ToggleTrash(ctx, newSession.ID); err != nil {
		t.Fatalf("toggle trash failed: %v", err)
	}
	if err := repo.ToggleFavorite(ctx, newSession.ID); err != nil {
		t.Fatalf("toggle favorite failed: %v", err)
	}
	if err := repo.MoveToFolder(ctx, newSession.ID, nil); err != nil {
		t.Fatalf("move to folder nil failed: %v", err)
	}
	if err := repo.UpdateSummary(ctx, newSession.ID, "ringkas"); err != nil {
		t.Fatalf("update summary failed: %v", err)
	}
	if err := repo.Delete(ctx, newSession.ID); err != nil {
		t.Fatalf("delete session failed: %v", err)
	}

	_ = groupedTrash
	if err := db.Migrator().DropTable(&model.ChatSession{}); err != nil {
		t.Fatalf("drop chat_sessions table failed: %v", err)
	}
	if _, err := repo.FindByUserIDGroupedByFolder(ctx, 7, "trash"); err == nil {
		t.Fatal("expected grouped-by-folder error when sessions table is missing")
	}
}

func TestChatMessageRepository_Ops(t *testing.T) {
	db := setupChatRepoDB(t)
	repo := NewChatMessageRepository(db)
	ctx := context.Background()

	msg, err := repo.FindByID(ctx, 1)
	if err != nil || msg.ID != 1 {
		t.Fatalf("find message failed: %v", err)
	}
	if _, err := repo.FindByID(ctx, 999999); err == nil {
		t.Fatal("expected message FindByID missing error")
	}

	msg.Content = "updated"
	if err := repo.Update(ctx, msg); err != nil {
		t.Fatalf("update message failed: %v", err)
	}

	if err := repo.ToggleLike(ctx, 1); err != nil {
		t.Fatalf("toggle like failed: %v", err)
	}
	if err := repo.ToggleDislike(ctx, 1); err != nil {
		t.Fatalf("toggle dislike failed: %v", err)
	}
	if err := repo.TogglePin(ctx, 1); err != nil {
		t.Fatalf("toggle pin failed: %v", err)
	}
	if err := repo.SetPinned(ctx, 1, true); err != nil {
		t.Fatalf("set pin failed: %v", err)
	}

	pinned, err := repo.FindPinnedBySessionID(ctx, 1)
	if err != nil || len(pinned) == 0 {
		t.Fatalf("find pinned failed: %v", err)
	}
	count, err := repo.CountPinnedBySessionID(ctx, 1)
	if err != nil || count == 0 {
		t.Fatalf("count pinned failed: %d %v", count, err)
	}

	newMsg := &model.ChatMessage{UUID: uuid.New(), ChatSessionID: 1, Role: model.ChatRoleUser, Content: "new"}
	if err := repo.Create(ctx, newMsg); err != nil {
		t.Fatalf("create message failed: %v", err)
	}
}

func TestChatFolderRepository_Delete_ErrorBranch(t *testing.T) {
	ctx := context.Background()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.Exec(`CREATE TABLE chat_folders (
		id INTEGER PRIMARY KEY,
		uuid TEXT,
		user_id INTEGER,
		name TEXT,
		color TEXT,
		icon TEXT,
		position INTEGER,
		created_at DATETIME,
		updated_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create chat_folders failed: %v", err)
	}
	if err := db.Exec(`INSERT INTO chat_folders (id, uuid, user_id, name, position) VALUES (1, '11111111-1111-1111-1111-111111111111', 7, 'Folder A', 0)`).Error; err != nil {
		t.Fatalf("seed chat_folders failed: %v", err)
	}

	repo := NewChatFolderRepository(db)
	if err := repo.Delete(ctx, 1); err == nil {
		t.Fatal("expected delete to fail when chat_sessions table is missing")
	}
}

func TestChatFolderRepository_ReorderFolders_ErrorBranch(t *testing.T) {
	ctx := context.Background()
	db := setupChatRepoDB(t)
	repo := NewChatFolderRepository(db)

	if err := db.Exec(`DROP TABLE chat_folders`).Error; err != nil {
		t.Fatalf("drop chat_folders table failed: %v", err)
	}

	if err := repo.ReorderFolders(ctx, 7, []uint{1}); err == nil {
		t.Fatal("expected ReorderFolders error when chat_folders table is missing")
	}
}
