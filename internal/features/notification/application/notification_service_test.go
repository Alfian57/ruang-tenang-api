package application

import (
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
)

func TestNotificationTargetURL(t *testing.T) {
	tests := []struct {
		name string
		n    model.Notification
		want string
	}{
		{
			name: "story detail",
			n: model.Notification{
				Type: model.NotificationTypeStoryApproved,
				Data: `{"story_id":"story-123"}`,
			},
			want: "/dashboard/community/stories/story-123",
		},
		{
			name: "story list fallback",
			n:    model.Notification{Type: model.NotificationTypeStoryRejected},
			want: "/dashboard/community?tab=stories",
		},
		{
			name: "level journey",
			n:    model.Notification{Type: model.NotificationTypeLevelUp},
			want: "/dashboard/journey",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := notificationTargetURL(&tt.n); got != tt.want {
				t.Fatalf("notificationTargetURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
