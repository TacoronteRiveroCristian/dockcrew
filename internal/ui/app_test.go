package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/TacoronteRiveroCristian/dockcrew/internal/domain"
)

func TestFormatPorts(t *testing.T) {
	ports := []domain.PortBinding{{IP: "0.0.0.0", Public: 8080, Private: 80, Type: "tcp"}, {Public: 0, Private: 5432, Type: "tcp"}}
	out := formatPorts(ports)
	if !strings.Contains(out, "0.0.0.0:8080->80/tcp") {
		t.Fatalf("expected host mapping, got %q", out)
	}
	if !strings.Contains(out, "5432/tcp") {
		t.Fatalf("expected internal port, got %q", out)
	}
}

func TestBuildRowsTruncatesAndFormats(t *testing.T) {
	created := time.Now().Add(-2 * time.Hour)
	containers := []domain.Container{{
		ID:        "123",
		Name:      "very-long-container-name-that-should-truncate",
		Image:     "ghcr.io/example/app:latest",
		State:     "running",
		Status:    "Up 2 hours",
		Stack:     "demo-stack",
		CreatedAt: created,
		Ports:     []domain.PortBinding{{IP: "", Public: 9000, Private: 9000, Type: "tcp"}},
	}}
	rows := buildRows(containers)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if !strings.Contains(rows[0][0], "Up 2 hours") {
		t.Fatalf("status column missing: %q", rows[0][0])
	}
	if len([]rune(rows[0][1])) > 24 {
		t.Fatalf("name not truncated: %q", rows[0][1])
	}
	if rows[0][5] == "" {
		t.Fatalf("created column empty")
	}
}

func TestRelativeTime(t *testing.T) {
	ts := time.Now().Add(-90 * time.Minute)
	out := relativeTime(ts)
	if !strings.Contains(out, "h") {
		t.Fatalf("expected hours display, got %q", out)
	}
}
