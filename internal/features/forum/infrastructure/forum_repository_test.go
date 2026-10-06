package infrastructure

import (
	"context"
	"strings"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestSupportCircleFilterIsAppliedBeforePagination(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{DSN: "test:test@tcp(localhost:3306)/test", SkipInitializeWithVersion: true}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	repo := &forumRepository{db: db}
	statement := repo.forumsQuery(context.Background(), "", nil, "pemulihan_burnout").Session(&gorm.Session{DryRun: true}).Limit(10).Offset(20).Find(&[]model.Forum{}).Statement
	sql := statement.SQL.String()
	if !strings.Contains(sql, "JOIN forum_categories") || !strings.Contains(sql, "LIKE") || !strings.Contains(sql, "LIMIT") || !strings.Contains(sql, "OFFSET") {
		t.Fatalf("expected server-side circle filter and pagination, got %s", sql)
	}
	if len(statement.Vars) < 3 || statement.Vars[0] != "%burnout%" || statement.Vars[len(statement.Vars)-2] != 10 || statement.Vars[len(statement.Vars)-1] != 20 {
		t.Fatalf("unexpected circle keywords: %#v", statement.Vars)
	}
}
