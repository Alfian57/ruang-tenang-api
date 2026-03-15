package contentctx

import (
	"fmt"
	"strings"
)

// SearchArticles performs keyword-based search on in-memory article cache.
// No database query is made — only the pre-cached map is searched.
func (s *ContentContextService) SearchArticles(query string, category string, limit int) []ArticleSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 5
	}

	queryLower := strings.ToLower(query)
	categoryLower := strings.ToLower(category)

	var results []ArticleSummary
	for _, article := range s.articles {
		if category != "" && !strings.EqualFold(article.Category, category) {
			continue
		}

		titleLower := strings.ToLower(article.Title)
		if query == "" || strings.Contains(titleLower, queryLower) || strings.Contains(strings.ToLower(article.Category), categoryLower) {
			results = append(results, *article)
			if len(results) >= limit {
				break
			}
		}
	}

	return results
}

// SearchMusic searches music categories from in-memory cache by mood/keyword.
func (s *ContentContextService) SearchMusic(mood string) []SongCategorySummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if mood == "" {
		// Return all categories
		var results []SongCategorySummary
		for _, cat := range s.songCategories {
			results = append(results, *cat)
		}
		return results
	}

	moodLower := strings.ToLower(mood)
	var results []SongCategorySummary
	for _, cat := range s.songCategories {
		nameLower := strings.ToLower(cat.Name)
		if strings.Contains(nameLower, moodLower) || matchesMoodToCategory(moodLower, nameLower) {
			results = append(results, *cat)
		}
	}

	// If no specific match found, return all categories as suggestions
	if len(results) == 0 {
		for _, cat := range s.songCategories {
			results = append(results, *cat)
		}
	}

	return results
}

// SearchForums performs keyword-based search on in-memory forum cache.
func (s *ContentContextService) SearchForums(query string, limit int) []ForumSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 5
	}

	queryLower := strings.ToLower(query)

	var results []ForumSummary
	for _, forum := range s.forums {
		titleLower := strings.ToLower(forum.Title)
		if query == "" || strings.Contains(titleLower, queryLower) {
			results = append(results, *forum)
			if len(results) >= limit {
				break
			}
		}
	}

	return results
}

// FormatArticleResults formats article search results for Gemini function response.
func FormatArticleResults(articles []ArticleSummary) string {
	if len(articles) == 0 {
		return "Tidak ada artikel yang ditemukan untuk pencarian ini."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Ditemukan %d artikel relevan:\n", len(articles)))
	for _, a := range articles {
		sb.WriteString(fmt.Sprintf("- \"%s\" (Kategori: %s) URL: https://ruang-tenang.site/articles/%d\n", a.Title, a.Category, a.ID))
	}
	sb.WriteString("\nGunakan format markdown clickable [Judul](URL) saat merekomendasikan.")
	return sb.String()
}

// FormatMusicResults formats music search results for Gemini function response.
func FormatMusicResults(categories []SongCategorySummary) string {
	if len(categories) == 0 {
		return "Tidak ada kategori musik yang ditemukan."
	}

	var sb strings.Builder
	sb.WriteString("Kategori musik yang tersedia:\n")
	for _, c := range categories {
		sb.WriteString(fmt.Sprintf("- \"%s\" (%d lagu) URL: https://ruang-tenang.site/dashboard/music\n", c.Name, c.SongCount))
	}
	sb.WriteString("\nGunakan format markdown clickable [Nama Kategori](https://ruang-tenang.site/dashboard/music) saat merekomendasikan.")
	return sb.String()
}

// FormatForumResults formats forum search results for Gemini function response.
func FormatForumResults(forums []ForumSummary) string {
	if len(forums) == 0 {
		return "Tidak ada topik forum yang ditemukan untuk pencarian ini."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Ditemukan %d topik forum relevan:\n", len(forums)))
	for _, f := range forums {
		sb.WriteString(fmt.Sprintf("- \"%s\" (%d balasan) URL: https://ruang-tenang.site/dashboard/forum/%d\n", f.Title, f.RepliesCount, f.ID))
	}
	sb.WriteString("\nGunakan format markdown clickable [Judul](URL) saat merekomendasikan.")
	return sb.String()
}

// matchesMoodToCategory maps common mood keywords to category names.
func matchesMoodToCategory(mood string, category string) bool {
	moodMappings := map[string][]string{
		"sedih":    {"tenang", "relaksasi", "calm", "santai", "healing"},
		"sad":      {"tenang", "relaksasi", "calm", "santai", "healing"},
		"crying":   {"tenang", "relaksasi", "calm", "santai", "healing"},
		"cemas":    {"tenang", "relaksasi", "calm", "meditasi", "nature"},
		"anxious":  {"tenang", "relaksasi", "calm", "meditasi", "nature"},
		"marah":    {"tenang", "relaksasi", "calm", "nature"},
		"angry":    {"tenang", "relaksasi", "calm", "nature"},
		"happy":    {"semangat", "ceria", "upbeat", "energi"},
		"senang":   {"semangat", "ceria", "upbeat", "energi"},
		"neutral":  {"tenang", "santai", "lo-fi"},
		"stres":    {"tenang", "relaksasi", "meditasi", "nature", "healing"},
		"stress":   {"tenang", "relaksasi", "meditasi", "nature", "healing"},
		"insomnia": {"tidur", "sleep", "tenang", "relaksasi"},
		"tidur":    {"tidur", "sleep", "tenang", "relaksasi"},
	}

	keywords, exists := moodMappings[mood]
	if !exists {
		return false
	}

	for _, kw := range keywords {
		if strings.Contains(category, kw) {
			return true
		}
	}
	return false
}
