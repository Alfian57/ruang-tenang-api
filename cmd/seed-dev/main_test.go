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
	origDevRunCLIFn          = runCLIFn
	origDevMainFatalFn       = mainFatalFn
	origDevResetTablesFn     = resetTablesFn
	origDevSetenvFn          = setenvFn
	origDevProductionSeeders = devProductionSeeders
	origDevTestSeeders       = devTestSeeders
	origDevLoadConfigFn      = seed.LoadConfigFn
	origDevConnectDBFn       = seed.ConnectDBFn
)

func resetDeps() {
	runCLIFn = origDevRunCLIFn
	mainFatalFn = origDevMainFatalFn
	resetTablesFn = origDevResetTablesFn
	setenvFn = origDevSetenvFn
	devProductionSeeders = origDevProductionSeeders
	devTestSeeders = origDevTestSeeders
	seed.LoadConfigFn = origDevLoadConfigFn
	seed.ConnectDBFn = origDevConnectDBFn
}

func TestRunDevelopmentSeeder_SkipAll(t *testing.T) {
	if err := runDevelopmentSeeder(nil, seed.SeedOptions{Only: "nonexistent"}); err != nil {
		t.Fatalf("expected skip-all success, got %v", err)
	}
}

func TestRunDevelopmentSeeder_ResetError(t *testing.T) {
	t.Cleanup(resetDeps)

	resetTablesFn = func(*gorm.DB) error { return errors.New("reset fail") }
	if err := runDevelopmentSeeder(nil, seed.SeedOptions{Reset: true}); err == nil || !strings.Contains(err.Error(), "reset fail") {
		t.Fatalf("expected reset error, got %v", err)
	}
}

func TestRunDevelopmentSeeder_ProductionPhaseError(t *testing.T) {
	t.Cleanup(resetDeps)

	devProductionSeeders = []seed.SeederRunner{{Name: "Prod A", Fn: func(*gorm.DB) error { return errors.New("prod fail") }}}
	devTestSeeders = []seed.SeederRunner{}
	if err := runDevelopmentSeeder(nil, seed.SeedOptions{Only: "proda"}); err == nil || !strings.Contains(err.Error(), "prod fail") {
		t.Fatalf("expected prod error, got %v", err)
	}
}

func TestRunDevelopmentSeeder_DevPhaseError(t *testing.T) {
	t.Cleanup(resetDeps)

	devProductionSeeders = []seed.SeederRunner{}
	devTestSeeders = []seed.SeederRunner{{Name: "Dev A", Fn: func(*gorm.DB) error { return errors.New("dev fail") }}}
	if err := runDevelopmentSeeder(nil, seed.SeedOptions{Only: "deva"}); err == nil || !strings.Contains(err.Error(), "dev fail") {
		t.Fatalf("expected dev error, got %v", err)
	}
}

func TestRunDevelopmentSeeder_ResetSuccess(t *testing.T) {
	t.Cleanup(resetDeps)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	devProductionSeeders = []seed.SeederRunner{}
	devTestSeeders = []seed.SeederRunner{}

	if err := runDevelopmentSeeder(db, seed.SeedOptions{Reset: true, Only: "nonexistent"}); err != nil {
		t.Fatalf("expected reset + skip-all success, got %v", err)
	}
}

func TestRunDevelopmentSeeder_CountBranch(t *testing.T) {
	t.Cleanup(resetDeps)

	devProductionSeeders = []seed.SeederRunner{}
	devTestSeeders = []seed.SeederRunner{}

	if err := runDevelopmentSeeder(nil, seed.SeedOptions{Count: 5, Only: "nonexistent"}); err != nil {
		t.Fatalf("expected count + skip-all success, got %v", err)
	}
}

func TestRunCLI_Success(t *testing.T) {
	t.Cleanup(resetDeps)

	seed.LoadConfigFn = func() (*config.Config, error) { return &config.Config{AppEnv: "test"}, nil }
	seed.ConnectDBFn = func(*config.Config) (*gorm.DB, error) { return &gorm.DB{}, nil }
	devProductionSeeders = []seed.SeederRunner{}
	devTestSeeders = []seed.SeederRunner{}

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

	setEnvCalled := ""
	setenvFn = func(key, val string) error {
		if key == "SEED_COUNT" {
			setEnvCalled = val
		}
		return nil
	}

	flag.CommandLine = flag.NewFlagSet("seed-dev-test", flag.ContinueOnError)
	os.Args = []string{"seeder-dev", "--reset", "--count=7", "--only=Articles"}
	main()

	if !received.Reset || received.Count != 7 || received.Only != "articles" {
		t.Fatalf("unexpected forwarded opts: %+v", received)
	}
	if setEnvCalled != "7" {
		t.Fatalf("expected SEED_COUNT=7, got %q", setEnvCalled)
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

	flag.CommandLine = flag.NewFlagSet("seed-dev-err", flag.ContinueOnError)
	os.Args = []string{"seeder-dev"}
	main()

	if !fatalCalled {
		t.Fatal("expected fatal path")
	}
}
