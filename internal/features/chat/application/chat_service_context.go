package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
)

const defaultChatSessionIntent = "general"

var validChatSessionIntents = map[string]struct{}{
	"general":    {},
	"grounding":  {},
	"planning":   {},
	"reflection": {},
	"coping":     {},
}

func normalizeSessionIntent(value string) string {
	normalized := strings.TrimSpace(strings.ToLower(value))
	if normalized == "" {
		return defaultChatSessionIntent
	}
	if _, exists := validChatSessionIntents[normalized]; !exists {
		return defaultChatSessionIntent
	}
	return normalized
}

func describeSessionIntent(intent string) string {
	switch normalizeSessionIntent(intent) {
	case "grounding":
		return "fokus pada menenangkan diri dan grounding langkah demi langkah"
	case "planning":
		return "fokus pada perencanaan aksi kecil yang realistis"
	case "reflection":
		return "fokus pada refleksi pola emosi, kebutuhan, dan pelajaran"
	case "coping":
		return "fokus pada strategi coping yang aman dan praktis"
	default:
		return "fokus pada dukungan umum dengan empati"
	}
}

func (s *ChatService) getSessionContextPreferences(session *model.ChatSession) dto.ChatContextPreferencesDTO {
	if session == nil {
		return dto.ChatContextPreferencesDTO{
			EnableMoodContext:        true,
			EnableJournalContext:     false,
			EnableDailyTaskContext:   true,
			EnableXPLevelContext:     true,
			EnableBreathingContext:   true,
			EnablePlaylistContext:    false,
			EnableRewardsContext:     false,
			EnableProgressMapContext: false,
			EnableSocialContext:      false,
			SessionIntent:            defaultChatSessionIntent,
		}
	}

	return dto.ChatContextPreferencesDTO{
		EnableMoodContext:        session.EnableMoodContext,
		EnableJournalContext:     session.EnableJournalContext,
		EnableDailyTaskContext:   session.EnableDailyTaskContext,
		EnableXPLevelContext:     session.EnableXPLevelContext,
		EnableBreathingContext:   session.EnableBreathingContext,
		EnablePlaylistContext:    session.EnablePlaylistContext,
		EnableRewardsContext:     session.EnableRewardsContext,
		EnableProgressMapContext: session.EnableProgressMapContext,
		EnableSocialContext:      session.EnableSocialContext,
		SessionIntent:            normalizeSessionIntent(string(session.SessionIntent)),
	}
}

func (s *ChatService) applyMessageContextHints(base dto.ChatContextPreferencesDTO, hints *dto.MessageContextHints) dto.ChatContextPreferencesDTO {
	if hints == nil {
		return base
	}

	if hints.EnableMoodContext != nil {
		base.EnableMoodContext = *hints.EnableMoodContext
	}
	if hints.EnableJournalContext != nil {
		base.EnableJournalContext = *hints.EnableJournalContext
	}
	if hints.EnableDailyTaskContext != nil {
		base.EnableDailyTaskContext = *hints.EnableDailyTaskContext
	}
	if hints.EnableXPLevelContext != nil {
		base.EnableXPLevelContext = *hints.EnableXPLevelContext
	}
	if hints.EnableBreathingContext != nil {
		base.EnableBreathingContext = *hints.EnableBreathingContext
	}
	if hints.EnablePlaylistContext != nil {
		base.EnablePlaylistContext = *hints.EnablePlaylistContext
	}
	if hints.EnableRewardsContext != nil {
		base.EnableRewardsContext = *hints.EnableRewardsContext
	}
	if hints.EnableProgressMapContext != nil {
		base.EnableProgressMapContext = *hints.EnableProgressMapContext
	}
	if hints.EnableSocialContext != nil {
		base.EnableSocialContext = *hints.EnableSocialContext
	}
	if strings.TrimSpace(hints.SessionIntent) != "" {
		base.SessionIntent = normalizeSessionIntent(hints.SessionIntent)
	}

	return base
}

func (s *ChatService) resolveContextPreferences(session *model.ChatSession, hints *dto.MessageContextHints) dto.ChatContextPreferencesDTO {
	base := s.getSessionContextPreferences(session)
	return s.applyMessageContextHints(base, hints)
}

func (s *ChatService) buildContextState(ctx context.Context, session *model.ChatSession, userID uint, preferences dto.ChatContextPreferencesDTO, hints *dto.MessageContextHints) *dto.ChatContextStateDTO {
	runtime := dto.ChatContextRuntimeDTO{
		EffectiveSources: make([]string, 0, 9),
	}

	var user *model.User
	if (preferences.EnableXPLevelContext || preferences.EnableRewardsContext) && s.userRepo != nil {
		if existingUser, err := s.userRepo.FindByID(ctx, userID); err == nil && existingUser != nil {
			user = existingUser
		}
	}

	if preferences.EnableMoodContext {
		moodValue := ""
		moodEmoji := ""
		if hints != nil {
			moodValue = strings.TrimSpace(hints.CurrentMood)
		}
		if moodValue != "" {
			runtime.Mood = &dto.ChatContextMoodDTO{Mood: moodValue, Emoji: moodEmoji}
			runtime.EffectiveSources = append(runtime.EffectiveSources, "mood")
		} else if s.userContextCache != nil {
			if moodCtx := s.userContextCache.GetMoodContext(ctx, userID); moodCtx != nil {
				runtime.Mood = &dto.ChatContextMoodDTO{Mood: moodCtx.Mood, Emoji: moodCtx.Emoji}
				runtime.EffectiveSources = append(runtime.EffectiveSources, "mood")
			}
		}
	}

	journalEnabledGlobally := false
	journalSharedCount := 0
	if s.journalRepo != nil && s.journalSettingsRepo != nil {
		if settings, err := s.journalSettingsRepo.FindByUserID(ctx, userID); err == nil && settings != nil && settings.AllowAIAccess {
			journalEnabledGlobally = true
			if count, err := s.journalRepo.CountSharedWithAI(ctx, userID); err == nil {
				journalSharedCount = int(count)
			}
		}
	}
	runtime.JournalSharedCount = journalSharedCount
	if preferences.EnableJournalContext && journalEnabledGlobally && journalSharedCount > 0 {
		runtime.EffectiveSources = append(runtime.EffectiveSources, "journal")
	}

	if preferences.EnableDailyTaskContext && s.dailyTaskService != nil {
		if summary, err := s.dailyTaskService.GetTodayTasks(ctx, userID); err == nil && summary != nil {
			pending := summary.TotalTasks - summary.CompletedTasks
			if pending < 0 {
				pending = 0
			}
			runtime.DailyTask = &dto.ChatContextDailyTaskDTO{
				Completed: summary.CompletedTasks,
				Pending:   pending,
			}
			runtime.EffectiveSources = append(runtime.EffectiveSources, "daily_task")
		}
	}

	if preferences.EnableXPLevelContext {
		if user != nil {
			xpLevel := &dto.ChatContextXPLevelDTO{
				Exp:           user.Exp,
				CurrentStreak: user.CurrentStreak,
			}
			if s.levelConfigService != nil {
				currentLevel, nextLevel, err := s.levelConfigService.GetUserLevelInfo(ctx, user.Exp)
				if err == nil {
					if currentLevel != nil {
						xpLevel.CurrentLevel = currentLevel.Level
					}
					if nextLevel != nil {
						xpLevel.NextLevel = nextLevel.Level
					}
				}
			}
			runtime.XPLevel = xpLevel
			runtime.EffectiveSources = append(runtime.EffectiveSources, "xp_level")
		}
	}

	if preferences.EnableBreathingContext && s.breathingRepo != nil {
		breathing := &dto.ChatContextBreathingDTO{}
		hasBreathingData := false

		if sessionsToday, err := s.breathingRepo.GetUserSessionsToday(ctx, userID); err == nil {
			breathing.SessionsToday = len(sessionsToday)
			hasBreathingData = true
		}

		if sessionsLast7Days, err := s.breathingRepo.CountSessionsSince(ctx, userID, time.Now().AddDate(0, 0, -7)); err == nil {
			breathing.SessionsLast7Days = int(sessionsLast7Days)
			hasBreathingData = true
		}

		if technique, _, err := s.breathingRepo.GetMostUsedTechnique(ctx, userID); err == nil && technique != nil {
			breathing.MostUsedTechnique = technique.Name
			hasBreathingData = true
		}

		if hasBreathingData {
			runtime.Breathing = breathing
			runtime.EffectiveSources = append(runtime.EffectiveSources, "breathing")
		}
	}

	if preferences.EnablePlaylistContext && s.playlistRepo != nil {
		if playlists, itemCounts, err := s.playlistRepo.FindByUserIDWithItemCount(ctx, userID); err == nil {
			playlist := &dto.ChatContextPlaylistDTO{
				TotalPlaylists: len(playlists),
			}

			totalSavedSongs := 0
			for _, currentPlaylist := range playlists {
				totalSavedSongs += itemCounts[currentPlaylist.ID]
			}
			playlist.TotalSavedSongs = totalSavedSongs

			if len(playlists) > 0 {
				playlist.LatestPlaylistTitle = strings.TrimSpace(playlists[0].Name)
			}

			runtime.Playlist = playlist
			runtime.EffectiveSources = append(runtime.EffectiveSources, "playlist")
		}
	}

	if preferences.EnableRewardsContext {
		rewards := &dto.ChatContextRewardsDTO{}
		hasRewardsData := false

		if user != nil {
			rewards.GoldCoins = user.GoldCoins
			hasRewardsData = true
		}

		if s.rewardRepo != nil {
			if claims, total, err := s.rewardRepo.GetUserClaims(ctx, userID, 1, 0); err == nil {
				rewards.ClaimCount = int(total)
				if len(claims) > 0 {
					rewards.LatestRewardName = strings.TrimSpace(claims[0].Reward.Name)
				}
				hasRewardsData = true
			}
		}

		if hasRewardsData {
			runtime.Rewards = rewards
			runtime.EffectiveSources = append(runtime.EffectiveSources, "rewards")
		}
	}

	if preferences.EnableProgressMapContext && s.progressMapRepo != nil {
		progressMap := &dto.ChatContextProgressMapDTO{}
		hasProgressData := false

		if unlockedRegions, err := s.progressMapRepo.CountUnlockedRegions(ctx, userID); err == nil {
			progressMap.UnlockedRegions = unlockedRegions
			hasProgressData = true
		}

		if unlockedLandmarks, err := s.progressMapRepo.CountUnlockedLandmarks(ctx, userID); err == nil {
			progressMap.UnlockedLandmarks = unlockedLandmarks
			hasProgressData = true
		}

		if latestUnlockName, err := s.progressMapRepo.GetLatestUnlock(ctx, userID); err == nil {
			progressMap.LatestUnlockName = strings.TrimSpace(latestUnlockName)
			hasProgressData = true
		}

		if hasProgressData {
			runtime.ProgressMap = progressMap
			runtime.EffectiveSources = append(runtime.EffectiveSources, "progress_map")
		}
	}

	if preferences.EnableSocialContext {
		social := &dto.ChatContextSocialDTO{}
		hasSocialData := false

		if s.badgeRepo != nil {
			if badgeCount, err := s.badgeRepo.GetUserBadgeCount(ctx, userID); err == nil {
				social.BadgeCount = badgeCount
				hasSocialData = true
			}
		}

		if s.guildRepo != nil {
			if membership, err := s.guildRepo.GetUserGuild(ctx, userID); err == nil && membership != nil {
				social.GuildRole = strings.TrimSpace(string(membership.Role))
				if membership.Guild != nil {
					social.GuildName = strings.TrimSpace(membership.Guild.Name)
				}
				if memberCount, err := s.guildRepo.GetMemberCount(ctx, membership.GuildID); err == nil {
					social.GuildMemberCount = memberCount
				}
				hasSocialData = true
			}
		}

		if hasSocialData {
			runtime.Social = social
			runtime.EffectiveSources = append(runtime.EffectiveSources, "social")
		}
	}

	return &dto.ChatContextStateDTO{
		SessionUUID: session.UUID.String(),
		Preferences: preferences,
		Runtime:     runtime,
	}
}

func (s *ChatService) buildDynamicContextPrompt(ctx context.Context, session *model.ChatSession, userID uint, req *dto.SendMessageRequest) string {
	if session == nil {
		return ""
	}

	var hints *dto.MessageContextHints
	if req != nil {
		hints = req.Context
	}

	preferences := s.resolveContextPreferences(session, hints)
	state := s.buildContextState(ctx, session, userID, preferences, hints)

	contextLines := []string{
		fmt.Sprintf("Fokus sesi saat ini: %s.", describeSessionIntent(state.Preferences.SessionIntent)),
	}

	if mood := state.Runtime.Mood; mood != nil {
		if mood.Emoji != "" {
			contextLines = append(contextLines, fmt.Sprintf("Mood user: %s %s.", mood.Emoji, mood.Mood))
		} else {
			contextLines = append(contextLines, fmt.Sprintf("Mood user: %s.", mood.Mood))
		}
	}

	if state.Preferences.EnableJournalContext {
		if state.Runtime.JournalSharedCount > 0 {
			contextLines = append(contextLines, fmt.Sprintf("User membagikan %d jurnal untuk konteks AI.", state.Runtime.JournalSharedCount))
		} else {
			contextLines = append(contextLines, "Konteks jurnal aktif, tetapi belum ada jurnal yang dibagikan ke AI.")
		}
	}

	if dailyTask := state.Runtime.DailyTask; dailyTask != nil {
		contextLines = append(contextLines, fmt.Sprintf("Progress tugas hari ini: %d selesai, %d tersisa.", dailyTask.Completed, dailyTask.Pending))
	}

	if xpLevel := state.Runtime.XPLevel; xpLevel != nil {
		if xpLevel.CurrentLevel > 0 {
			contextLines = append(contextLines, fmt.Sprintf("Progress pengguna: EXP %d, level %d, streak %d.", xpLevel.Exp, xpLevel.CurrentLevel, xpLevel.CurrentStreak))
		} else {
			contextLines = append(contextLines, fmt.Sprintf("Progress pengguna: EXP %d, streak %d.", xpLevel.Exp, xpLevel.CurrentStreak))
		}
	}

	if breathing := state.Runtime.Breathing; breathing != nil {
		if breathing.MostUsedTechnique != "" {
			contextLines = append(contextLines, fmt.Sprintf("Aktivitas napas: %d sesi hari ini, %d sesi dalam 7 hari, teknik favorit %s.", breathing.SessionsToday, breathing.SessionsLast7Days, breathing.MostUsedTechnique))
		} else {
			contextLines = append(contextLines, fmt.Sprintf("Aktivitas napas: %d sesi hari ini, %d sesi dalam 7 hari.", breathing.SessionsToday, breathing.SessionsLast7Days))
		}
	}

	if playlist := state.Runtime.Playlist; playlist != nil {
		if playlist.TotalPlaylists > 0 {
			if playlist.LatestPlaylistTitle != "" {
				contextLines = append(contextLines, fmt.Sprintf("Playlist pengguna: %d playlist dengan total %d lagu tersimpan, terbaru \"%s\".", playlist.TotalPlaylists, playlist.TotalSavedSongs, playlist.LatestPlaylistTitle))
			} else {
				contextLines = append(contextLines, fmt.Sprintf("Playlist pengguna: %d playlist dengan total %d lagu tersimpan.", playlist.TotalPlaylists, playlist.TotalSavedSongs))
			}
		} else {
			contextLines = append(contextLines, "Belum ada playlist personal yang dibuat pengguna.")
		}
	}

	if rewards := state.Runtime.Rewards; rewards != nil {
		if rewards.LatestRewardName != "" {
			contextLines = append(contextLines, fmt.Sprintf("Status reward: %d koin, %d klaim reward, klaim terbaru %s.", rewards.GoldCoins, rewards.ClaimCount, rewards.LatestRewardName))
		} else {
			contextLines = append(contextLines, fmt.Sprintf("Status reward: %d koin, %d klaim reward.", rewards.GoldCoins, rewards.ClaimCount))
		}
	}

	if progressMap := state.Runtime.ProgressMap; progressMap != nil {
		if progressMap.LatestUnlockName != "" {
			contextLines = append(contextLines, fmt.Sprintf("Progress map: %d region dan %d landmark terbuka, unlock terbaru %s.", progressMap.UnlockedRegions, progressMap.UnlockedLandmarks, progressMap.LatestUnlockName))
		} else {
			contextLines = append(contextLines, fmt.Sprintf("Progress map: %d region dan %d landmark terbuka.", progressMap.UnlockedRegions, progressMap.UnlockedLandmarks))
		}
	}

	if social := state.Runtime.Social; social != nil {
		socialLine := fmt.Sprintf("Status sosial: %d badge diperoleh", social.BadgeCount)
		if social.GuildName != "" {
			if social.GuildMemberCount > 0 {
				socialLine = fmt.Sprintf("%s, tergabung di guild %s (%s, %d anggota)", socialLine, social.GuildName, social.GuildRole, social.GuildMemberCount)
			} else if social.GuildRole != "" {
				socialLine = fmt.Sprintf("%s, tergabung di guild %s (%s)", socialLine, social.GuildName, social.GuildRole)
			} else {
				socialLine = fmt.Sprintf("%s, tergabung di guild %s", socialLine, social.GuildName)
			}
		}
		contextLines = append(contextLines, socialLine+".")
	}

	if len(state.Runtime.EffectiveSources) > 0 {
		contextLines = append(contextLines, fmt.Sprintf("Gunakan konteks ini seperlunya: %s.", strings.Join(state.Runtime.EffectiveSources, ", ")))
	}

	if len(contextLines) == 0 {
		return ""
	}

	return "\n\n=== KONTEKS DINAMIS PENGGUNA ===\n" + strings.Join(contextLines, "\n") + "\n=== AKHIR KONTEKS DINAMIS ===\n"
}

func (s *ChatService) GetContextState(ctx context.Context, sessionID, userID uint) (*dto.ChatContextStateDTO, error) {
	session, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, errors.New("session not found")
	}
	if session.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	preferences := s.getSessionContextPreferences(session)
	return s.buildContextState(ctx, session, userID, preferences, nil), nil
}

func (s *ChatService) UpdateContextPreferences(ctx context.Context, sessionID, userID uint, req *dto.UpdateChatContextPreferencesRequest) (*dto.ChatContextStateDTO, error) {
	session, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, errors.New("session not found")
	}
	if session.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	if req == nil {
		return s.GetContextState(ctx, sessionID, userID)
	}

	updates := make(map[string]interface{})
	if req.EnableMoodContext != nil {
		updates["enable_mood_context"] = *req.EnableMoodContext
	}
	if req.EnableJournalContext != nil {
		updates["enable_journal_context"] = *req.EnableJournalContext
	}
	if req.EnableDailyTaskContext != nil {
		updates["enable_daily_task_context"] = *req.EnableDailyTaskContext
	}
	if req.EnableXPLevelContext != nil {
		updates["enable_xp_level_context"] = *req.EnableXPLevelContext
	}
	if req.EnableBreathingContext != nil {
		updates["enable_breathing_context"] = *req.EnableBreathingContext
	}
	if req.EnablePlaylistContext != nil {
		updates["enable_playlist_context"] = *req.EnablePlaylistContext
	}
	if req.EnableRewardsContext != nil {
		updates["enable_rewards_context"] = *req.EnableRewardsContext
	}
	if req.EnableProgressMapContext != nil {
		updates["enable_progress_map_context"] = *req.EnableProgressMapContext
	}
	if req.EnableSocialContext != nil {
		updates["enable_social_context"] = *req.EnableSocialContext
	}
	if req.SessionIntent != nil {
		updates["session_intent"] = normalizeSessionIntent(*req.SessionIntent)
	}

	if err := s.sessionRepo.UpdateContextPreferences(ctx, session.ID, updates); err != nil {
		return nil, err
	}

	return s.GetContextState(ctx, session.ID, userID)
}
