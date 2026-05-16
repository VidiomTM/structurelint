package plugin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessPlugin_Name(t *testing.T) {
	p := NewProcessPlugin("my-plugin", "/usr/bin/true")
	assert.Equal(t, "my-plugin", p.Name())
}

func TestNewHTTPPluginClient_Unavailable(t *testing.T) {
	c := NewHTTPPluginClient("http://nonexistent.local:19999")
	require.NotNil(t, c)
	assert.False(t, c.available)
}

func TestHTTPPluginClient_Health_DecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	c := &HTTPPluginClient{
		baseURL:    server.URL,
		httpClient: server.Client(),
		available:  true,
	}
	_, err := c.Health(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode health response")
}
