package handler

import "testing"

func TestHandlerConstructorsAndSetters(t *testing.T) {
	if NewAdminHandler(nil, nil, nil, nil, nil, nil) == nil {
		t.Fatal("NewAdminHandler returned nil")
	}
	if NewArticleHandler(nil) == nil {
		t.Fatal("NewArticleHandler returned nil")
	}
	if NewAuthHandler(nil, nil) == nil {
		t.Fatal("NewAuthHandler returned nil")
	}
	if NewChatHandler(nil) == nil {
		t.Fatal("NewChatHandler returned nil")
	}
	if NewCommunityProgressHandler(nil) == nil {
		t.Fatal("NewCommunityProgressHandler returned nil")
	}
	if NewExpHistoryHandler(nil, nil) == nil {
		t.Fatal("NewExpHistoryHandler returned nil")
	}
	if NewMoodHandler(nil) == nil {
		t.Fatal("NewMoodHandler returned nil")
	}
	if NewNotificationHandler(nil) == nil {
		t.Fatal("NewNotificationHandler returned nil")
	}
	if NewPlaylistHandler(nil) == nil {
		t.Fatal("NewPlaylistHandler returned nil")
	}
	if NewSearchHandler(nil, nil) == nil {
		t.Fatal("NewSearchHandler returned nil")
	}
	if NewSongHandler(nil) == nil {
		t.Fatal("NewSongHandler returned nil")
	}
	if NewUserHandler(nil, nil) == nil {
		t.Fatal("NewUserHandler returned nil")
	}

	article := NewArticleHandler(nil)
	article.SetDailyTaskService(nil)

	chat := NewChatHandler(nil)
	chat.SetDailyTaskService(nil)
}
