package config

import "testing"

func TestValidate(t *testing.T) {
	if err := Default().Validate(); err != nil {
		t.Fatalf("default config should validate: %v", err)
	}

	cfg := Default()
	cfg.Format = "yaml"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected invalid format to fail")
	}
}
