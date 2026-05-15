package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

// TestHelperProcess isn't a real test; it's a helper process invoked by tests.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	// Read input
	var input PluginInput
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		fmt.Fprintf(os.Stderr, "failed to decode input: %v", err)
		os.Exit(1)
	}

	// Check if we should simulate failure
	if input.Config["fail"] == true {
		fmt.Fprint(os.Stderr, "simulated failure")
		os.Exit(1)
	}

	// Output violations
	violations := []rules.Violation{
		{Rule: "test-rule", Path: "test.go", Message: "test violation"},
	}
	if err := json.NewEncoder(os.Stdout).Encode(violations); err != nil {
		t.Fatalf("Failed to encode violations: %v", err)
	}
}

func TestProcessPlugin_Check(t *testing.T) {
	// Use the test binary itself as the plugin executable
	exe, err := os.Executable()
	if err != nil {
		t.Skip("Skipping test: cannot get executable path")
	}

	// Create plugin that calls TestHelperProcess
	p := NewProcessPlugin("test-rule", exe, "-test.run=TestHelperProcess")
	
	// We need to inject the environment variable.
	// Since ProcessPlugin uses exec.CommandContext which inherits os.Environ(),
	// we can set the env var for the current process, but that's racy in parallel tests.
	// Ideally ProcessPlugin should support custom Env.
	// For now, we'll assume sequential execution or just set it.
	_ = os.Setenv("GO_WANT_HELPER_PROCESS", "1")
	defer func() { _ = os.Unsetenv("GO_WANT_HELPER_PROCESS") }()

	// Test success
	files := []walker.FileInfo{{AbsPath: "test.go"}}
	config := map[string]interface{}{"fail": false}
	
	violations, err := p.Check(context.Background(), files, config)
	if err != nil {
		t.Fatalf("Check() failed: %v", err)
	}
	if len(violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(violations))
	}

	// Test failure
	configFail := map[string]interface{}{"fail": true}
	_, err = p.Check(context.Background(), files, configFail)
	if err == nil {
		t.Error("Expected Check() to fail")
	}
}
