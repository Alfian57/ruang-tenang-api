package infrastructure

import (
	"context"
	"strings"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestAvailableRewardsTypeFilterRetainsActiveBoundary(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{DSN: "test:test@tcp(localhost:3306)/test", SkipInitializeWithVersion: true}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	repo := NewRewardRepository(db)
	statement := repo.availableRewardsQuery(context.Background(), "theme").Session(&gorm.Session{DryRun: true}).Find(&[]model.Reward{}).Statement
	if !strings.Contains(statement.SQL.String(), "is_active") || !strings.Contains(statement.SQL.String(), "reward_type") {
		t.Fatalf("expected active and type filters in SQL, got %s", statement.SQL.String())
	}
	if len(statement.Vars) != 2 || statement.Vars[0] != true || statement.Vars[1] != "theme" {
		t.Fatalf("unexpected query values: %#v", statement.Vars)
	}
}
