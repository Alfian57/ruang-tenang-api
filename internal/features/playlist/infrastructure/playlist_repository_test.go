package infrastructure

import (
	"context"
	"strings"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestPublicPlaylistKindFilters(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=localhost user=test dbname=test sslmode=disable"}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	repo := NewPlaylistRepository(db)
	for _, tc := range []struct {
		kind  string
		admin bool
	}{{"official", true}, {"community", false}} {
		statement := repo.publicPlaylistsQuery(context.Background(), tc.kind).Session(&gorm.Session{DryRun: true}).Find(&[]model.Playlist{}).Statement
		if !strings.Contains(statement.SQL.String(), "is_public") || !strings.Contains(statement.SQL.String(), "is_admin_playlist") {
			t.Fatalf("%s must filter public and admin flag: %s", tc.kind, statement.SQL.String())
		}
		if len(statement.Vars) != 2 || statement.Vars[0] != true || statement.Vars[1] != tc.admin {
			t.Fatalf("%s has wrong filter values: %#v", tc.kind, statement.Vars)
		}
	}
}
