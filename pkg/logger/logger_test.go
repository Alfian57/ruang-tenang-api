package logger

import (
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestInitDevelopment(t *testing.T) {
	tmp := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd error: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir error: %v", err)
	}

	if err := Init("development"); err != nil {
		t.Fatalf("init development failed: %v", err)
	}
	if Log == nil || Sugar == nil {
		t.Fatalf("expected loggers to be initialized")
	}
	if _, err := os.Stat(filepath.Join(tmp, "logs")); err != nil {
		t.Fatalf("expected logs dir to be created: %v", err)
	}

	Info("dev info")
	Debug("dev debug")
	Warn("dev warn")
	Error("dev error")
	Sync()
}

func TestInitProductionAndSyncNil(t *testing.T) {
	tmp := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd error: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir error: %v", err)
	}

	if err := Init("production"); err != nil {
		t.Fatalf("init production failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmp, "logs", "app.log")); err != nil {
		t.Fatalf("expected app log file to be created: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmp, "logs", "error.log")); err != nil {
		t.Fatalf("expected error log file to be created: %v", err)
	}

	Sync()

	Log = nil
	Sync()

	Log = zap.NewNop()
	Info("nop")
	Debug("nop")
	Warn("nop")
	Error("nop")
}

func TestInitFailsWhenLogsPathIsFile(t *testing.T) {
	tmp := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd error: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir error: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmp, "logs"), []byte("x"), 0644); err != nil {
		t.Fatalf("write logs file error: %v", err)
	}

	if err := Init("development"); err == nil {
		t.Fatal("expected init to fail when logs path is a file")
	}
}

func TestFatalWithGoexitHook(t *testing.T) {
	base := zap.NewNop()
	Log = base.WithOptions(zap.WithFatalHook(zapcore.WriteThenGoexit))

	done := make(chan struct{})
	returned := make(chan bool, 1)
	go func() {
		defer close(done)
		Fatal("fatal test")
		returned <- true
	}()

	<-done
	select {
	case <-returned:
		t.Fatal("fatal should have called runtime.Goexit in goroutine")
	default:
	}
}
