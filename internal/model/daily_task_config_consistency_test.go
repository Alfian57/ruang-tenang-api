package model

import "testing"

func TestDailyTaskConfigs_HaveUniqueTypesAndPositiveValues(t *testing.T) {
	configs := GetDailyTaskConfigs()
	if len(configs) == 0 {
		t.Fatal("expected daily task configs to be non-empty")
	}

	seen := map[DailyTaskType]bool{}
	for _, cfg := range configs {
		if seen[cfg.Type] {
			t.Fatalf("duplicate task type found: %s", cfg.Type)
		}
		seen[cfg.Type] = true

		if cfg.Name == "" || cfg.Description == "" || cfg.Icon == "" {
			t.Fatalf("task config should have non-empty presentation fields: %+v", cfg)
		}
		if cfg.TargetCount <= 0 {
			t.Fatalf("task target count should be > 0: %+v", cfg)
		}
		if cfg.XPReward <= 0 {
			t.Fatalf("task XP reward should be > 0: %+v", cfg)
		}
	}
}
