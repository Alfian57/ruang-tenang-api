package model

import "testing"

func TestUserMood_GetMoodEmoji_AllBranches(t *testing.T) {
	tests := []struct {
		mood MoodType
		want string
	}{
		{MoodHappy, "😊"},
		{MoodNeutral, "😐"},
		{MoodAngry, "😠"},
		{MoodDisappointed, "😞"},
		{MoodSad, "😢"},
		{MoodCrying, "😭"},
		{MoodType("unknown"), "🙂"},
	}

	for _, tc := range tests {
		m := &UserMood{Mood: tc.mood}
		if got := m.GetMoodEmoji(); got != tc.want {
			t.Fatalf("mood %q expected %q, got %q", tc.mood, tc.want, got)
		}
	}
}
