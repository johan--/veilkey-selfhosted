//go:build !windows

package main

import (
	"os"
	"testing"
)

func TestTransformPastedInputIssuesTempWhenConfigured(t *testing.T) {
	t.Setenv("VEILKEY_PLAINTEXT_ACTION", "issue-temp-and-block")

	logger := NewSessionLogger(t.TempDir() + "/session.log")
	d := &SecretDetector{
		config:   &CompiledConfig{},
		client:   &VeilKeyClient{},
		logger:   logger,
		cache:    map[string]string{"demo-secret": "VK:TEMP:testref"},
		scanOnly: false,
	}

	got := transformPastedInput(d, []byte("demo-secret\n"))
	if got != "VK:TEMP:testref\n" {
		t.Fatalf("expected issued token, got %q", got)
	}
	if d.Stats.Detections != 1 {
		t.Fatalf("expected 1 detection, got %d", d.Stats.Detections)
	}
}

func TestTransformPastedInputLeavesExistingRefs(t *testing.T) {
	t.Setenv("VEILKEY_PLAINTEXT_ACTION", "issue-temp-and-block")

	d := &SecretDetector{config: &CompiledConfig{}, cache: map[string]string{}}
	got := transformPastedInput(d, []byte("VK:TEMP:abcd1234\n"))
	if got != "VK:TEMP:abcd1234\n" {
		t.Fatalf("expected existing ref to pass through, got %q", got)
	}
}

func TestTransformPastedInputFallsBackWithoutAction(t *testing.T) {
	os.Unsetenv("VEILKEY_PLAINTEXT_ACTION")

	d := &SecretDetector{config: &CompiledConfig{}, cache: map[string]string{}, scanOnly: true}
	got := transformPastedInput(d, []byte("plain text\n"))
	if got != "plain text\n" {
		t.Fatalf("expected fallback passthrough, got %q", got)
	}
}
