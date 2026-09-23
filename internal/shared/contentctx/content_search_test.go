package contentctx

import (
	"strings"
	"testing"
)

func TestFormatForumResultsUsesCommunitySlugRoute(t *testing.T) {
	result := FormatForumResults([]ForumSummary{{
		ID:           17,
		Slug:         "ruang-aman-untuk-cerita",
		Title:        "Ruang aman untuk cerita",
		RepliesCount: 4,
	}})

	want := "https://ruang-tenang.site/dashboard/community/forum/ruang-aman-untuk-cerita"
	if !strings.Contains(result, want) {
		t.Fatalf("FormatForumResults() tidak memuat URL kanonis %q: %s", want, result)
	}
	if strings.Contains(result, "/dashboard/forum/") {
		t.Fatalf("FormatForumResults() masih memuat route forum lama: %s", result)
	}
}

func TestForumURLFallsBackToCommunity(t *testing.T) {
	if got := forumURL(ForumSummary{}); got != "https://ruang-tenang.site/dashboard/community" {
		t.Fatalf("forumURL() = %q", got)
	}
}
