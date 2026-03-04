package main

import (
	"errors"
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/internal/seed"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	origRunCLIFn          = runCLIFn
	origMainFatalFn       = mainFatalFn
	origGetenvFn          = getenvFn
	origProductionSeeders = productionSeeders
	origLoadConfigFn      = seed.LoadConfigFn
	origConnectDBFn       = seed.ConnectDBFn
)

func resetDeps() {
	runCLIFn = origRunCLIFn
	mainFatalFn = origMainFatalFn
	getenvFn = origGetenvFn
	productionSeeders = origProductionSeeders
	seed.LoadConfigFn = origLoadConfigFn
	seed.ConnectDBFn = origConnectDBFn
}

func TestRunProductionSeeder_SkipAll(t *testing.T) {
	if err := runProductionSeeder(nil, seed.SeedOptions{Only: "nonexistent"}); err != nil {
		t.Fatalf("expected skip-all success, got %v", err)
	}
}

func TestRunProductionSeeder_SeederError(t *testing.T) {
	t.Cleanup(func() { productionSeeders = origProductionSeeders })

	productionSeeders = []seed.SeederRunner{{Name: "Fail", Fn: func(*gorm.DB) error { return errors.New("boom") }}}
	if err := runProductionSeeder(nil, seed.SeedOptions{Only: "fail"}); err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected seeder error, got %v", err)
	}
}

func TestRunProductionSeeder_AdminEnvValidation(t *testing.T) {
	t.Cleanup(resetDeps)

	getenvFn = func(key string) string {
		if key == "SEED_PROD_CREATE_ADMIN" {
			return "true"
		}
		return ""
	}

	err := runProductionSeeder(nil, seed.SeedOptions{Only: "nonexistent"})
	if err == nil || !strings.Contains(err.Error(), "SEED_ADMIN_EMAIL") {
		t.Fatalf("expected env validation error, got %v", err)
	}
}

func TestRunCLI_Success(t *testing.T) {
	t.Cleanup(resetDeps)

	seed.LoadConfigFn = func() (*config.Config, error) { return &config.Config{AppEnv: "test"}, nil }
	seed.ConnectDBFn = func(*config.Config) (*gorm.DB, error) { return &gorm.DB{}, nil }
	productionSeeders = []seed.SeederRunner{}

	if err := runCLI(seed.SeedOptions{Only: "nonexistent"}); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestRunCLI_ConfigError(t *testing.T) {
	t.Cleanup(resetDeps)

	seed.LoadConfigFn = func() (*config.Config, error) { return nil, errors.New("cfg") }
	if err := runCLI(seed.SeedOptions{}); err == nil || !strings.Contains(err.Error(), "Failed to load configuration") {
		t.Fatalf("expected config error, got %v", err)
	}
}

func TestRunCLI_DBError(t *testing.T) {
	t.Cleanup(resetDeps)

	seed.LoadConfigFn = func() (*config.Config, error) { return &config.Config{}, nil }
	seed.ConnectDBFn = func(*config.Config) (*gorm.DB, error) { return nil, errors.New("db") }
	if err := runCLI(seed.SeedOptions{}); err == nil || !strings.Contains(err.Error(), "Failed to connect to database") {
		t.Fatalf("expected db error, got %v", err)
	}
}

func TestMain_FlagForwarding(t *testing.T) {
	t.Cleanup(resetDeps)

	origArgs := os.Args
	origCmd := flag.CommandLine
	defer func() {
		os.Args = origArgs
		flag.CommandLine = origCmd
	}()

	var received seed.SeedOptions
	runCLIFn = func(opts seed.SeedOptions) error {
		received = opts
		return nil
	}
	mainFatalFn = func(v ...any) {
		t.Fatalf("did not expect fatal, got %v", v)
	}

	flag.CommandLine = flag.NewFlagSet("seed-prod-test", flag.ContinueOnError)
	os.Args = []string{"seeder-prod", "--only=Level Configs"}
	main()

	if received.Only != "levelconfigs" {
		t.Fatalf("expected only=levelconfigs, got %q", received.Only)
	}
}

func TestMain_ErrorCallsFatal(t *testing.T) {
	t.Cleanup(resetDeps)

	origArgs := os.Args
	origCmd := flag.CommandLine
	defer func() {
		os.Args = origArgs
		flag.CommandLine = origCmd
	}()

	fatalCalled := false
	runCLIFn = func(seed.SeedOptions) error { return errors.New("boom") }
	mainFatalFn = func(v ...any) { fatalCalled = true }

	flag.CommandLine = flag.NewFlagSet("seed-prod-err", flag.ContinueOnError)
	os.Args = []string{"seeder-prod"}
	main()

	if !fatalCalled {
		t.Fatal("expected fatal path")
	}
}

func TestResetAllTables_SharedLib(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	_ = db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)`)
	_ = db.Exec(`INSERT INTO users (name) VALUES ('alice')`)

	if err := seed.ResetAllTables(db); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	var count int64
	db.Raw(`SELECT COUNT(*) FROM users`).Scan(&count)
	if count != 0 {
		t.Fatalf("expected 0 users, got %d", count)
	}
}
