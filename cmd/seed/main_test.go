package main

import (
	"errors"
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/pkg/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	origSeedLoadConfigFn             = seedLoadConfigFn
	origSeedConnectDBFn              = seedConnectDBFn
	origSeedRunProductionSeederFn    = seedRunProductionSeederFn
	origSeedRunDevelopmentSeederFn   = seedRunDevelopmentSeederFn
	origSeedRunTestingSeederFn       = seedRunTestingSeederFn
	origResetAllTablesFn             = resetAllTablesFn
	origRunSeedCLIFn                 = runSeedCLIFn
	origSeedMainFatalFn              = seedMainFatalFn
	origDevelopmentProductionSeeders = developmentProductionSeeders
	origDevelopmentTestSeeders       = developmentTestSeeders
	origProductionSeeders            = productionSeeders
	origSeedSetenvFn                 = seedSetenvFn
	origSeedGetenvFn                 = seedGetenvFn
)

func resetSeedDeps() {
	seedLoadConfigFn = origSeedLoadConfigFn
	seedConnectDBFn = origSeedConnectDBFn
	seedRunProductionSeederFn = origSeedRunProductionSeederFn
	seedRunDevelopmentSeederFn = origSeedRunDevelopmentSeederFn
	seedRunTestingSeederFn = origSeedRunTestingSeederFn
	resetAllTablesFn = origResetAllTablesFn
	runSeedCLIFn = origRunSeedCLIFn
	seedMainFatalFn = origSeedMainFatalFn
	developmentProductionSeeders = origDevelopmentProductionSeeders
	developmentTestSeeders = origDevelopmentTestSeeders
	productionSeeders = origProductionSeeders
	seedSetenvFn = origSeedSetenvFn
	seedGetenvFn = origSeedGetenvFn
}

func TestNormalizeSeederNameAndShouldRunSeeder(t *testing.T) {
	if got := normalizeSeederName("  Test Users "); got != "testusers" {
		t.Fatalf("expected testusers, got %s", got)
	}

	if !shouldRunSeeder("", "Anything") {
		t.Fatalf("expected empty only to run all")
	}
	if !shouldRunSeeder("all", "Anything") {
		t.Fatalf("expected all to run all")
	}
	if !shouldRunSeeder("testusers", "Test Users") {
		t.Fatalf("expected normalized match")
	}
	if shouldRunSeeder("articles", "Songs") {
		t.Fatalf("expected different seeder name not to run")
	}
}

func TestRunTestingSeeder_InvalidOnly(t *testing.T) {
	err := runTestingSeeder(nil, SeedOptions{Only: "invalid"})
	if err == nil {
		t.Fatalf("expected invalid only error")
	}
}

func TestRunTestingSeeder_SuccessAndResetError(t *testing.T) {
	if err := logger.Init("test"); err != nil {
		t.Fatalf("init logger: %v", err)
	}
	defer logger.Sync()

	for _, only := range []string{"", "all", "seed"} {
		err := runTestingSeeder(nil, SeedOptions{Only: only})
		if err != nil {
			t.Fatalf("expected success for only=%q, got %v", only, err)
		}
	}

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := runTestingSeeder(db, SeedOptions{Reset: true, Only: "all"}); err != nil {
		t.Fatalf("expected reset success on sqlite testing seeder, got %v", err)
	}
}

func TestRunProductionSeeder_AdminEnvValidation(t *testing.T) {
	oldCreate := os.Getenv("SEED_PROD_CREATE_ADMIN")
	oldEmail := os.Getenv("SEED_ADMIN_EMAIL")
	oldPass := os.Getenv("SEED_ADMIN_PASSWORD")
	defer func() {
		_ = os.Setenv("SEED_PROD_CREATE_ADMIN", oldCreate)
		_ = os.Setenv("SEED_ADMIN_EMAIL", oldEmail)
		_ = os.Setenv("SEED_ADMIN_PASSWORD", oldPass)
	}()

	_ = os.Setenv("SEED_PROD_CREATE_ADMIN", "true")
	_ = os.Unsetenv("SEED_ADMIN_EMAIL")
	_ = os.Unsetenv("SEED_ADMIN_PASSWORD")

	err := runProductionSeeder(nil, SeedOptions{Only: "nonexistent"})
	if err == nil {
		t.Fatalf("expected env validation error")
	}
}

func TestRunProductionAndDevelopmentSeeder_SkipAll(t *testing.T) {
	if err := runProductionSeeder(nil, SeedOptions{Only: "nonexistent"}); err != nil {
		t.Fatalf("expected production skip-all success, got %v", err)
	}

	if err := runDevelopmentSeeder(nil, SeedOptions{Only: "nonexistent"}); err != nil {
		t.Fatalf("expected development skip-all success, got %v", err)
	}
}

func TestRunProductionAndDevelopmentSeeder_CountAndResetBranches(t *testing.T) {
	if err := runProductionSeeder(nil, SeedOptions{Only: "nonexistent", Count: 5}); err != nil {
		t.Fatalf("expected production count+skip-all success, got %v", err)
	}
	if err := runDevelopmentSeeder(nil, SeedOptions{Only: "nonexistent", Count: 3}); err != nil {
		t.Fatalf("expected development count+skip-all success, got %v", err)
	}

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := runDevelopmentSeeder(db, SeedOptions{Reset: true, Only: "nonexistent"}); err != nil {
		t.Fatalf("expected development reset success on sqlite, got %v", err)
	}
}

func TestResetAllTables_Branches(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.Exec(`CREATE TABLE schema_migrations (version TEXT)`).Error; err != nil {
		t.Fatalf("create schema_migrations: %v", err)
	}
	if err := db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)`).Error; err != nil {
		t.Fatalf("create users: %v", err)
	}
	if err := db.Exec(`CREATE TABLE songs (id INTEGER PRIMARY KEY, title TEXT)`).Error; err != nil {
		t.Fatalf("create songs: %v", err)
	}

	if err := db.Exec(`INSERT INTO schema_migrations (version) VALUES ('1')`).Error; err != nil {
		t.Fatalf("insert schema_migrations: %v", err)
	}
	if err := db.Exec(`INSERT INTO users (name) VALUES ('alice')`).Error; err != nil {
		t.Fatalf("insert users: %v", err)
	}
	if err := db.Exec(`INSERT INTO songs (title) VALUES ('song-a')`).Error; err != nil {
		t.Fatalf("insert songs: %v", err)
	}

	if err := resetAllTables(db); err != nil {
		t.Fatalf("expected sqlite resetAllTables success, got %v", err)
	}

	var usersCount, songsCount, migrationCount int64
	if err := db.Raw(`SELECT COUNT(*) FROM users`).Scan(&usersCount).Error; err != nil {
		t.Fatalf("count users: %v", err)
	}
	if err := db.Raw(`SELECT COUNT(*) FROM songs`).Scan(&songsCount).Error; err != nil {
		t.Fatalf("count songs: %v", err)
	}
	if err := db.Raw(`SELECT COUNT(*) FROM schema_migrations`).Scan(&migrationCount).Error; err != nil {
		t.Fatalf("count schema_migrations: %v", err)
	}

	if usersCount != 0 || songsCount != 0 {
		t.Fatalf("expected app tables truncated, got users=%d songs=%d", usersCount, songsCount)
	}
	if migrationCount != 1 {
		t.Fatalf("expected schema_migrations unchanged, got %d", migrationCount)
	}
}

func TestResetAllTables_EmptyAndMigrationsOnly(t *testing.T) {
	emptyDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite empty db: %v", err)
	}
	if err := resetAllTables(emptyDB); err != nil {
		t.Fatalf("expected success on empty sqlite db, got %v", err)
	}

	migrationsOnlyDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite migrations-only db: %v", err)
	}
	if err := migrationsOnlyDB.Exec(`CREATE TABLE schema_migrations (version TEXT)`).Error; err != nil {
		t.Fatalf("create schema_migrations: %v", err)
	}
	if err := migrationsOnlyDB.Exec(`INSERT INTO schema_migrations (version) VALUES ('1')`).Error; err != nil {
		t.Fatalf("insert schema_migrations: %v", err)
	}

	if err := resetAllTables(migrationsOnlyDB); err != nil {
		t.Fatalf("expected success for migrations-only db, got %v", err)
	}

	var migrationCount int64
	if err := migrationsOnlyDB.Raw(`SELECT COUNT(*) FROM schema_migrations`).Scan(&migrationCount).Error; err != nil {
		t.Fatalf("count schema_migrations: %v", err)
	}
	if migrationCount != 1 {
		t.Fatalf("expected schema_migrations untouched, got %d", migrationCount)
	}
}

func TestResetAllTables_SqliteSequenceBranch(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)`).Error; err != nil {
		t.Fatalf("create users: %v", err)
	}
	if err := db.Exec(`INSERT INTO users (name) VALUES ('alice')`).Error; err != nil {
		t.Fatalf("insert users: %v", err)
	}

	// Ensure sqlite_sequence exists before reset.
	var seqBefore int64
	if err := db.Raw(`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'sqlite_sequence'`).Scan(&seqBefore).Error; err != nil {
		t.Fatalf("check sqlite_sequence before: %v", err)
	}
	if seqBefore != 1 {
		t.Fatalf("expected sqlite_sequence to exist, got %d", seqBefore)
	}

	if err := resetAllTables(db); err != nil {
		t.Fatalf("resetAllTables failed: %v", err)
	}

	if err := db.Exec(`INSERT INTO users (name) VALUES ('bob')`).Error; err != nil {
		t.Fatalf("insert users after reset: %v", err)
	}

	var newID int64
	if err := db.Raw(`SELECT id FROM users WHERE name = 'bob'`).Scan(&newID).Error; err != nil {
		t.Fatalf("read new id: %v", err)
	}
	if newID != 1 {
		t.Fatalf("expected sqlite sequence reset to 1, got %d", newID)
	}
}

type postgresNamedDialector struct {
	gorm.Dialector
}

func (d postgresNamedDialector) Name() string {
	return "postgres"
}

func TestResetAllTables_PostgresBranchQueryError(t *testing.T) {
	db, err := gorm.Open(postgresNamedDialector{Dialector: sqlite.Open(":memory:")}, &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite with postgres-named dialector: %v", err)
	}

	if err := resetAllTables(db); err == nil {
		t.Fatal("expected error on postgres branch query against sqlite backend")
	}
}

func TestResolveSeedMode_PriorityAndAliases(t *testing.T) {
	tests := []struct {
		name   string
		mode   string
		env    string
		legacy string
		want   string
	}{
		{name: "mode wins", mode: "testing", env: "production", legacy: "development", want: "testing"},
		{name: "env used when mode empty", mode: "", env: "production", legacy: "development", want: "production"},
		{name: "legacy used when mode and env empty", mode: "", env: "", legacy: "testing", want: "testing"},
		{name: "default development", mode: "", env: "", legacy: "", want: "development"},
		{name: "prod alias", mode: "prod", env: "", legacy: "", want: "production"},
		{name: "dev alias", mode: "dev", env: "", legacy: "", want: "development"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveSeedMode(tt.mode, tt.env, tt.legacy)
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestRunSeedCLI_ValidationBranches(t *testing.T) {
	t.Cleanup(resetSeedDeps)

	seedGetenvFn = func(key string) string {
		switch key {
		case "SEED_MODE":
			return ""
		case "SEED_CONFIRM":
			return ""
		default:
			return ""
		}
	}

	if err := runSeedCLI(seedCLIArgs{modeFlag: "unknown"}); err == nil || !strings.Contains(err.Error(), "Invalid SEED_MODE") {
		t.Fatalf("expected invalid mode error, got %v", err)
	}

	if err := runSeedCLI(seedCLIArgs{modeFlag: "production", reset: true}); err == nil || !strings.Contains(err.Error(), "--reset is forbidden") {
		t.Fatalf("expected reset forbidden error, got %v", err)
	}

	if err := runSeedCLI(seedCLIArgs{modeFlag: "production"}); err == nil || !strings.Contains(err.Error(), "SEED_CONFIRM=YES") {
		t.Fatalf("expected confirmation error, got %v", err)
	}
}

func TestRunSeedCLI_DependencyAndDispatchBranches(t *testing.T) {
	t.Cleanup(resetSeedDeps)

	seedLoadConfigFn = func() (*config.Config, error) { return &config.Config{AppEnv: "test"}, nil }
	seedConnectDBFn = func(*config.Config) (*gorm.DB, error) { return &gorm.DB{}, nil }
	seedGetenvFn = func(key string) string {
		if key == "SEED_CONFIRM" {
			return "YES"
		}
		return ""
	}

	called := ""
	seedRunProductionSeederFn = func(_ *gorm.DB, opts SeedOptions) error {
		called = "production:" + opts.Only
		return nil
	}
	seedRunDevelopmentSeederFn = func(_ *gorm.DB, opts SeedOptions) error {
		called = "development:" + opts.Only
		return nil
	}
	seedRunTestingSeederFn = func(_ *gorm.DB, opts SeedOptions) error {
		called = "testing:" + opts.Only
		return nil
	}

	setEnvValue := ""
	seedSetenvFn = func(key string, value string) error {
		if key == "SEED_COUNT" {
			setEnvValue = value
		}
		return nil
	}

	if err := runSeedCLI(seedCLIArgs{modeFlag: "production", only: " Test Users "}); err != nil {
		t.Fatalf("expected production success, got %v", err)
	}
	if called != "production:testusers" {
		t.Fatalf("expected production dispatcher, got %q", called)
	}

	if err := runSeedCLI(seedCLIArgs{modeFlag: "development", count: 7, only: " Songs "}); err != nil {
		t.Fatalf("expected development success, got %v", err)
	}
	if called != "development:songs" {
		t.Fatalf("expected development dispatcher, got %q", called)
	}
	if setEnvValue != "7" {
		t.Fatalf("expected SEED_COUNT=7, got %q", setEnvValue)
	}

	seedRunTestingSeederFn = func(_ *gorm.DB, _ SeedOptions) error {
		return errors.New("testing failed")
	}
	if err := runSeedCLI(seedCLIArgs{modeFlag: "testing"}); err == nil || !strings.Contains(err.Error(), "Testing seeding failed") {
		t.Fatalf("expected testing seeding error, got %v", err)
	}
}

func TestRunSeedCLI_ConfigAndDBErrors(t *testing.T) {
	t.Cleanup(resetSeedDeps)

	seedGetenvFn = func(string) string { return "" }
	seedLoadConfigFn = func() (*config.Config, error) { return nil, errors.New("cfg") }
	if err := runSeedCLI(seedCLIArgs{}); err == nil || !strings.Contains(err.Error(), "Failed to load configuration") {
		t.Fatalf("expected config error, got %v", err)
	}

	seedLoadConfigFn = func() (*config.Config, error) { return &config.Config{}, nil }
	seedConnectDBFn = func(*config.Config) (*gorm.DB, error) { return nil, errors.New("db") }
	if err := runSeedCLI(seedCLIArgs{}); err == nil || !strings.Contains(err.Error(), "Failed to connect to database") {
		t.Fatalf("expected db error, got %v", err)
	}
}

func TestRunDevelopmentSeeder_BranchesWithInjectedSeeders(t *testing.T) {
	t.Cleanup(resetSeedDeps)

	resetAllTablesFn = func(*gorm.DB) error { return errors.New("reset fail") }
	if err := runDevelopmentSeeder(nil, SeedOptions{Reset: true}); err == nil || !strings.Contains(err.Error(), "reset fail") {
		t.Fatalf("expected reset error, got %v", err)
	}

	resetAllTablesFn = func(*gorm.DB) error { return nil }
	developmentProductionSeeders = []seederRunner{{name: "Prod A", fn: func(*gorm.DB) error { return errors.New("prod seeder fail") }}}
	developmentTestSeeders = []seederRunner{{name: "Dev A", fn: func(*gorm.DB) error { return nil }}}
	if err := runDevelopmentSeeder(nil, SeedOptions{Only: "proda"}); err == nil || !strings.Contains(err.Error(), "prod seeder fail") {
		t.Fatalf("expected production phase seeder error, got %v", err)
	}

	developmentProductionSeeders = []seederRunner{{name: "Prod A", fn: func(*gorm.DB) error { return nil }}}
	developmentTestSeeders = []seederRunner{{name: "Dev A", fn: func(*gorm.DB) error { return errors.New("dev seeder fail") }}}
	if err := runDevelopmentSeeder(nil, SeedOptions{Only: "deva", Count: 3}); err == nil || !strings.Contains(err.Error(), "dev seeder fail") {
		t.Fatalf("expected development phase seeder error, got %v", err)
	}

	developmentTestSeeders = []seederRunner{{name: "Dev A", fn: func(*gorm.DB) error { return nil }}}
	if err := runDevelopmentSeeder(nil, SeedOptions{Only: "deva", Count: 5}); err != nil {
		t.Fatalf("expected development seeder success, got %v", err)
	}
}

func TestRunProductionSeeder_InjectedSeedersAndErrors(t *testing.T) {
	t.Cleanup(resetSeedDeps)

	productionSeeders = []seederRunner{{name: "Seed A", fn: func(*gorm.DB) error { return errors.New("seed fail") }}}
	if err := runProductionSeeder(nil, SeedOptions{Only: "seeda"}); err == nil || !strings.Contains(err.Error(), "seed fail") {
		t.Fatalf("expected production seeder fail, got %v", err)
	}

	productionSeeders = []seederRunner{{name: "Seed A", fn: func(*gorm.DB) error { return nil }}}
	if err := runProductionSeeder(nil, SeedOptions{Only: "seeda", Count: 1}); err != nil {
		t.Fatalf("expected production seeder success, got %v", err)
	}
}

func TestMain_BranchesAndFlagForwarding(t *testing.T) {
	t.Cleanup(resetSeedDeps)

	origArgs := os.Args
	origCmd := flag.CommandLine
	defer func() {
		os.Args = origArgs
		flag.CommandLine = origCmd
	}()

	var received seedCLIArgs
	runSeedCLIFn = func(args seedCLIArgs) error {
		received = args
		return nil
	}
	seedMainFatalFn = func(v ...any) {
		t.Fatalf("did not expect fatal on success, got %v", v)
	}

	flag.CommandLine = flag.NewFlagSet("seed-test", flag.ContinueOnError)
	os.Args = []string{"seed", "--mode=testing", "--env=dev", "--reset", "--count=7", "--only=Test Users"}
	main()

	if received.modeFlag != "testing" || received.legacyEnvFlag != "dev" || !received.reset || received.count != 7 || received.only != "Test Users" {
		t.Fatalf("unexpected forwarded args: %+v", received)
	}

	fatalCalled := false
	runSeedCLIFn = func(seedCLIArgs) error { return errors.New("boom") }
	seedMainFatalFn = func(v ...any) { fatalCalled = true }

	flag.CommandLine = flag.NewFlagSet("seed-test-err", flag.ContinueOnError)
	os.Args = []string{"seed", "--mode=testing"}
	main()
	if !fatalCalled {
		t.Fatal("expected fatal path when runSeedCLI returns error")
	}
}
