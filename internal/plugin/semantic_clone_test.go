package plugin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoOpPlugin(t *testing.T) {
	p := NewNoOpPlugin()
	require.NotNil(t, p)

	resp, err := p.DetectClones(context.Background(), &SemanticCloneRequest{
		SourceDir: "/test",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Clones)
	assert.Equal(t, "semantic clone detection plugin not available", resp.Error)

	health, err := p.Health(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "unhealthy", health.Status)

	assert.False(t, p.IsAvailable())
}

func TestHTTPPluginClient_IsAvailable(t *testing.T) {
	c := &HTTPPluginClient{available: false}
	assert.False(t, c.IsAvailable())

	c.available = true
	assert.True(t, c.IsAvailable())
}

func TestHTTPPluginClient_DetectClones_NotAvailable(t *testing.T) {
	c := &HTTPPluginClient{available: false}
	_, err := c.DetectClones(context.Background(), &SemanticCloneRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugin not available")
}

func TestHTTPPluginClient_DetectClones_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/detect", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"clones":[{"source_file":"a.go","target_file":"b.go","similarity":0.95}],"stats":{"files_analyzed":2}}`))
	}))
	defer server.Close()

	c := &HTTPPluginClient{
		baseURL:    server.URL,
		httpClient: server.Client(),
		available:  true,
	}

	resp, err := c.DetectClones(context.Background(), &SemanticCloneRequest{
		SourceDir: "/test",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 1, len(resp.Clones))
	assert.Equal(t, "a.go", resp.Clones[0].SourceFile)
	assert.Equal(t, 0.95, resp.Clones[0].Similarity)
}

func TestHTTPPluginClient_DetectClones_NonOK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer server.Close()

	c := &HTTPPluginClient{
		baseURL:    server.URL,
		httpClient: server.Client(),
		available:  true,
	}

	_, err := c.DetectClones(context.Background(), &SemanticCloneRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "returned status 500")
}

func TestHTTPPluginClient_DetectClones_ErrorInResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"clones":[],"error":"analysis failed"}`))
	}))
	defer server.Close()

	c := &HTTPPluginClient{
		baseURL:    server.URL,
		httpClient: server.Client(),
		available:  true,
	}

	_, err := c.DetectClones(context.Background(), &SemanticCloneRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "analysis failed")
}

func TestHTTPPluginClient_DetectClones_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("server crash")
	}))
	defer server.Close()

	c := &HTTPPluginClient{
		baseURL:    server.URL,
		httpClient: server.Client(),
		available:  true,
	}

	_, err := c.DetectClones(context.Background(), &SemanticCloneRequest{})
	assert.Error(t, err)
	assert.False(t, c.available)
}

func TestHTTPPluginClient_Health_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/health", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy","version":"1.0.0","capabilities":["clone-detection"]}`))
	}))
	defer server.Close()

	c := &HTTPPluginClient{
		baseURL:    server.URL,
		httpClient: server.Client(),
		available:  true,
	}

	health, err := c.Health(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "healthy", health.Status)
	assert.Equal(t, "1.0.0", health.Version)
	assert.Contains(t, health.Capabilities, "clone-detection")
}

func TestHTTPPluginClient_Health_NonOK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	c := &HTTPPluginClient{
		baseURL:    server.URL,
		httpClient: server.Client(),
		available:  true,
	}

	health, err := c.Health(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "unhealthy", health.Status)
}

func TestHTTPPluginClient_Health_Error(t *testing.T) {
	c := &HTTPPluginClient{
		baseURL:    "http://nonexistent.local:9999",
		httpClient: &http.Client{},
		available:  true,
	}

	_, err := c.Health(context.Background())
	assert.Error(t, err)
}
