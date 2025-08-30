package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// helper to capture stdout during fn
func captureStdout(fn func()) string {
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w

	fn()

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = old
	return string(out)
}

func TestStartDryRunWithPlatform(t *testing.T) {
	// create temp config file with one platform
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")
	cfg := map[string]interface{}{
		"platforms": []map[string]string{
			{
				"name":                       "itest",
				"ANTHROPIC_BASE_URL":         "https://api.itest.local",
				"ANTHROPIC_AUTH_TOKEN":       "tokentest123456",
				"ANTHROPIC_MODEL":            "m1",
				"ANTHROPIC_SMALL_FAST_MODEL": "m1-fast",
			},
		},
	}
	b, _ := json.MarshalIndent(cfg, "", "  ")
	if err := os.WriteFile(cfgPath, b, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// set cfgFile global to the temp config path
	cfgFile = cfgPath

	// load platforms directly and verify values and mask behavior
	platforms, err := loadPlatforms(cfgPath)
	if err != nil {
		t.Fatalf("loadPlatforms failed: %v", err)
	}
	if len(platforms) != 1 {
		t.Fatalf("expected 1 platform, got %d", len(platforms))
	}
	p := platforms[0]
	if p.AnthropicBaseURL != "https://api.itest.local" {
		t.Fatalf("expected base url, got: %s", p.AnthropicBaseURL)
	}
	masked := maskToken(p.AnthropicAuthToken)
	if masked == p.AnthropicAuthToken || masked == "" {
		t.Fatalf("maskToken did not mask token properly: %s", masked)
	}
}

func TestCompletionBashOutputsSomething(t *testing.T) {
	// reuse cfgFile default
	// capture stdout
	out := captureStdout(func() {
		rootCmd.SetArgs([]string{"completion", "bash"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("completion execute failed: %v", err)
		}
	})

	if len(out) < 50 {
		t.Fatalf("expected non-empty completion script, got: %q", out)
	}
	if !strings.Contains(out, "complete") && !strings.Contains(out, "__start_cctool") {
		// try to be flexible about the exact contents
		t.Fatalf("completion output doesn't look like a shell completion script: %q", out[:200])
	}
}
