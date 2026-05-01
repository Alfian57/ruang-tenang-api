package presentation

import (
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

// SeedFeatureDefinitions seeds the feature definitions for progressive unlocking
func SeedFeatureDefinitions(db *gorm.DB) error {
	features := []model.FeatureDefinition{
		// Level 1-2 (Free for all)
		{FeatureKey: "basic_access", FeatureName: "Basic Access", Description: "Access to all core features", Icon: "✅", RequiredLevel: 1, Category: "core", DisplayOrder: 1},
		{FeatureKey: "basic_breathing", FeatureName: "4-7-8 Breathing", Description: "Basic breathing technique", Icon: "🌬️", RequiredLevel: 1, Category: "breathing", DisplayOrder: 2},

		// Level 3
		{FeatureKey: "custom_playlists", FeatureName: "Custom Playlists", Description: "Create your own music playlists", Icon: "🎵", RequiredLevel: 3, Category: "music", DisplayOrder: 10},
		{FeatureKey: "advanced_mood_tracking", FeatureName: "Advanced Mood Charts", Description: "View detailed mood analytics and charts", Icon: "📊", RequiredLevel: 3, Category: "mood", DisplayOrder: 11},
		{FeatureKey: "breathing_timer_custom", FeatureName: "Custom Breathing Timer", Description: "Customize breathing exercise duration", Icon: "⏱️", RequiredLevel: 3, Category: "breathing", DisplayOrder: 12},

		// Level 4
		{FeatureKey: "all_breathing_techniques", FeatureName: "All Breathing Techniques", Description: "Access to 5+ breathing techniques", Icon: "🧘", RequiredLevel: 4, Category: "breathing", DisplayOrder: 20},
		{FeatureKey: "chat_folders", FeatureName: "AI Chat Folders", Description: "Organize your conversations in folders", Icon: "💬", RequiredLevel: 4, Category: "chat", DisplayOrder: 21},
		{FeatureKey: "profile_themes", FeatureName: "Profile Themes", Description: "Customize your profile appearance", Icon: "🎨", RequiredLevel: 4, Category: "profile", DisplayOrder: 22},

		// Level 5
		{FeatureKey: "private_journal", FeatureName: "Private Journal", Description: "Write personal journal entries", Icon: "📝", RequiredLevel: 5, Category: "journal", DisplayOrder: 30},
		{FeatureKey: "pin_messages", FeatureName: "Pin Messages", Description: "Pin important chat messages", Icon: "📌", RequiredLevel: 5, Category: "chat", DisplayOrder: 31},
		{FeatureKey: "publish_articles", FeatureName: "Publish Articles", Description: "Write and publish user-generated articles", Icon: "📖", RequiredLevel: 5, Category: "content", DisplayOrder: 32},
		{FeatureKey: "submit_stories", FeatureName: "Submit Inspiring Stories", Description: "Share your recovery journey", Icon: "✨", RequiredLevel: 5, Category: "content", DisplayOrder: 33},

		// Level 6
		{FeatureKey: "chat_export", FeatureName: "Export Conversations", Description: "Export chat history to PDF/TXT", Icon: "📥", RequiredLevel: 6, Category: "chat", DisplayOrder: 40},
		{FeatureKey: "voice_input", FeatureName: "Voice Input", Description: "Use voice input for chat messages", Icon: "🎤", RequiredLevel: 6, Category: "chat", DisplayOrder: 41},
		{FeatureKey: "journal_tags", FeatureName: "Custom Journal Tags", Description: "Create custom tags for journal entries", Icon: "🏷️", RequiredLevel: 6, Category: "journal", DisplayOrder: 42},
		{FeatureKey: "create_polls", FeatureName: "Create Forum Polls", Description: "Create polls in forum discussions", Icon: "🗳️", RequiredLevel: 6, Category: "forum", DisplayOrder: 43},

		// Level 7
		{FeatureKey: "chat_summaries", FeatureName: "AI Chat Summaries", Description: "Generate AI summaries of conversations", Icon: "📋", RequiredLevel: 7, Category: "chat", DisplayOrder: 50},
		{FeatureKey: "report_content", FeatureName: "Report Content", Description: "Report inappropriate content", Icon: "🚨", RequiredLevel: 7, Category: "moderation", DisplayOrder: 51},
		{FeatureKey: "vote_best_answer", FeatureName: "Vote Best Answer", Description: "Vote for best answers in forum", Icon: "🎯", RequiredLevel: 7, Category: "forum", DisplayOrder: 52},
		{FeatureKey: "custom_notifications", FeatureName: "Custom Notifications", Description: "Advanced notification preferences", Icon: "🔔", RequiredLevel: 7, Category: "profile", DisplayOrder: 53},

		// Level 8
		{FeatureKey: "premium_themes", FeatureName: "Premium Themes", Description: "Access to premium themes and customizations", Icon: "🌈", RequiredLevel: 8, Category: "profile", DisplayOrder: 60},
		{FeatureKey: "exclusive_content", FeatureName: "Exclusive Content", Description: "Access exclusive mental health resources", Icon: "📚", RequiredLevel: 8, Category: "content", DisplayOrder: 61},
		{FeatureKey: "nominate_stories", FeatureName: "Nominate Stories", Description: "Nominate inspiring stories for featured section", Icon: "🏅", RequiredLevel: 8, Category: "content", DisplayOrder: 62},
		{FeatureKey: "advanced_search", FeatureName: "Advanced Search", Description: "Advanced search filters", Icon: "⚙️", RequiredLevel: 8, Category: "general", DisplayOrder: 63},

		// Level 9
		{FeatureKey: "moderator_application", FeatureName: "Moderator Application", Description: "Apply to become a community moderator", Icon: "👤", RequiredLevel: 9, Category: "moderation", DisplayOrder: 70},
		{FeatureKey: "mentor_badge", FeatureName: "Community Mentor Badge", Description: "Special Community Mentor badge", Icon: "✨", RequiredLevel: 9, Category: "badge", DisplayOrder: 71},
		{FeatureKey: "beta_features", FeatureName: "Beta Features", Description: "Early access to new features", Icon: "🎁", RequiredLevel: 9, Category: "general", DisplayOrder: 72},
		{FeatureKey: "featured_shoutout", FeatureName: "Featured Shoutout", Description: "Opportunity for featured member shoutout", Icon: "💌", RequiredLevel: 9, Category: "profile", DisplayOrder: 73},

		// Level 10
		{FeatureKey: "all_features", FeatureName: "All Features", Description: "Lifetime access to all features", Icon: "👑", RequiredLevel: 10, Category: "general", DisplayOrder: 80},
		{FeatureKey: "guardian_badge", FeatureName: "Guardian Badge", Description: "Exclusive Guardian badge", Icon: "🎖️", RequiredLevel: 10, Category: "badge", DisplayOrder: 81},
		{FeatureKey: "community_voting", FeatureName: "Community Voting", Description: "Vote on community feature decisions", Icon: "📢", RequiredLevel: 10, Category: "moderation", DisplayOrder: 82},
		{FeatureKey: "guardian_profile", FeatureName: "Guardian Profile", Description: "Profile featured in Community Guardians", Icon: "🌟", RequiredLevel: 10, Category: "profile", DisplayOrder: 83},
	}

	for _, feature := range features {
		var existing model.FeatureDefinition
		if db.Where("feature_key = ?", feature.FeatureKey).First(&existing).RowsAffected == 0 {
			if err := db.Create(&feature).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
