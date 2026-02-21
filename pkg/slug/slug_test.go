package slug

import (
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	t.Run("normalizes input", func(t *testing.T) {
		got := Generate("Hello, World! Go @ Ruang Tenang")
		if got != "hello-world-go-ruang-tenang" {
			t.Fatalf("unexpected slug: %s", got)
		}
	})

	t.Run("empty fallback", func(t *testing.T) {
		got := Generate("   --- ###   ")
		if got != "untitled" {
			t.Fatalf("expected untitled, got %s", got)
		}
	})

	t.Run("limits length", func(t *testing.T) {
		input := strings.Repeat("a", 260)
		got := Generate(input)
		if len(got) > 200 {
			t.Fatalf("slug should be <= 200 chars, got %d", len(got))
		}
	})
}

func TestGenerateUnique(t *testing.T) {
	base := Generate("My Title")
	got := GenerateUnique("My Title")

	if !strings.HasPrefix(got, base+"-") {
		t.Fatalf("expected prefix %q in %q", base+"-", got)
	}

	suffix := strings.TrimPrefix(got, base+"-")
	if len(suffix) != 6 {
		t.Fatalf("expected suffix length 6, got %d", len(suffix))
	}
}
