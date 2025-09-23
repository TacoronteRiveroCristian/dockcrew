package config

import "testing"

func TestDefaultAndValidate(t *testing.T) {
	cfg := Default()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate default failed: %v", err)
	}
	cfg.Logs.MaxLines = 10
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected error for low max_lines")
	}
}
