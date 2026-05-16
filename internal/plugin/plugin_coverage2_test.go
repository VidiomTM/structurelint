package plugin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/walker"
	"github.com/stretchr/testify/assert"
)

func TestProcessPlugin_Check_ExecFailure(t *testing.T) {
	p := NewProcessPlugin("test", "/nonexistent/plugin")
	files := []walker.FileInfo{{AbsPath: "test.go"}}
	_, err := p.Check(context.Background(), files, map[string]interface{}{})
	assert.Error(t, err)
}

func TestProcessPlugin_Check_BadOutput(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "badplugin.sh")
	os.WriteFile(script, []byte("#!/bin/sh\necho 'not valid json'\n"), 0755)

	p := NewProcessPlugin("test", "/bin/sh", "-c", "echo 'not valid json'")
	files := []walker.FileInfo{{AbsPath: "test.go"}}
	_, err := p.Check(context.Background(), files, map[string]interface{}{})
	assert.Error(t, err)
}

func TestNewHTTPPluginClient_Available(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy"}`))
	}))
	defer server.Close()

	c := NewHTTPPluginClient(server.URL)
	assert.True(t, c.available)
}
