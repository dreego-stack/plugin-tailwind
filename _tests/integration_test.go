package tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	tailwind "github.com/dreego-stack/plugin-tailwind"
)

func TestTailwindVersion(t *testing.T) {
	if tailwind.Version == "" {
		t.Fatal("Version must not be empty")
	}
}

func TestTailwindManifestValid(t *testing.T) {
	manifestPath := filepath.Join("..", "dreego-plugin.json")
	body, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read dreego-plugin.json: %v", err)
	}
	var manifest struct {
		Build struct {
			Steps []struct {
				Cmd  string `json:"cmd"`
				When string `json:"when"`
			} `json:"steps"`
		} `json:"build"`
	}
	if err := json.Unmarshal(body, &manifest); err != nil {
		t.Fatalf("parse dreego-plugin.json: %v", err)
	}
	if len(manifest.Build.Steps) == 0 {
		t.Fatal("no build steps in dreego-plugin.json")
	}
	for i, step := range manifest.Build.Steps {
		if step.Cmd == "" {
			t.Errorf("step %d: cmd is empty", i)
		}
		if step.When != "pre-build" {
			t.Errorf("step %d: when = %q, want 'pre-build'", i, step.When)
		}
	}
}