package infrastructure

import (
	"context"
	"strings"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestUserArticlesQueryKeepsOwnershipWhenSearching(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=localhost user=test dbname=test sslmode=disable"}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	repo := NewArticleRepository(db)
	statement := repo.userArticlesQuery(context.Background(), 42, "tenang").Session(&gorm.Session{DryRun: true}).Find(&[]model.Article{}).Statement
	sql := statement.SQL.String()
	if !strings.Contains(sql, "user_id") || !strings.Contains(sql, "ILIKE") {
		t.Fatalf("expected ownership and search filters in SQL, got %s", sql)
	}
	if len(statement.Vars) != 2 || statement.Vars[0] != uint(42) || statement.Vars[1] != "%tenang%" {
		t.Fatalf("unexpected query values: %#v", statement.Vars)
	}
}
