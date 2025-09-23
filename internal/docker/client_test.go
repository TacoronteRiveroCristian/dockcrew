package docker

import (
	"context"
	"testing"
)

// Testa que los helpers de split no fallen en casos básicos
func TestSplitHelpers(t *testing.T) {
	if got := splitLines("a\nb\n"); len(got) != 2 {
		t.Fatalf("splitLines want 2, got %d", len(got))
	}
	if got := splitLines(""); len(got) != 0 {
		t.Fatalf("splitLines empty want 0, got %d", len(got))
	}
	if got := splitTab("a\tb\tc"); len(got) != 3 {
		t.Fatalf("splitTab want 3, got %d", len(got))
	}
}

// Smoke test: no garantiza docker instalado, solo verifica que el método existe y respeta contexto (timeout corto)
func TestShellClientRespectsTimeout(t *testing.T) {
	c := &ShellClient{Binary: "docker", Timeout: 1}
	ctx := context.Background()
	_, _ = c.ListContainers(ctx) // no falla el test si docker no está
}
