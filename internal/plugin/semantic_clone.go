// Package plugin provides interfaces and clients for optional plugins
package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SemanticCloneRequest represents a request to detect semantic clones
type SemanticCloneRequest struct {
	// SourceDir is the root directory to analyze
	SourceDir string `json:"source_dir"`

	// Languages to analyze (e.g., "go", "python", "javascript")
	Languages []string `json:"languages,omitempty"`

	// ExcludePatterns are glob patterns to exclude
	ExcludePatterns []string `json:"exclude_patterns,omitempty"`

	// SimilarityThreshold (0.0-1.0) for clone detection
	SimilarityThreshold float64 `json:"similarity_threshold,omitempty"`

	// MaxResults limits the number of clone pairs returned
	MaxResults int `json:"max_results,omitempty"`
}

// SemanticClone represents a detected semantic clone pair
type SemanticClone struct {
	// SourceFile path of the first code snippet
	SourceFile string `json:"source_file"`

	// SourceStartLine in the source file
	SourceStartLine int `json:"source_start_line"`

	// SourceEndLine in the source file
	SourceEndLine int `json:"source_end_line"`

	// TargetFile path of the second code snippet
	TargetFile string `json:"target_file"`

	// TargetStartLine in the target file
	TargetStartLine int `json:"target_start_line"`

	// TargetEndLine in the target file
	TargetEndLine int `json:"target_end_line"`

	// Similarity score (0.0-1.0)
	Similarity float64 `json:"similarity"`

	// Explanation of why these are considered clones (optional)
	Explanation string `json:"explanation,omitempty"`
}

// SemanticCloneResponse represents the plugin's response
type SemanticCloneResponse struct {
	// Clones detected
	Clones []SemanticClone `json:"clones"`

	// Statistics about the analysis
	Stats SemanticCloneStats `json:"stats"`

	// Error message if analysis failed
	Error string `json:"error,omitempty"`
}

// SemanticCloneStats provides statistics about the analysis
type SemanticCloneStats struct {
	// FilesAnalyzed count
	FilesAnalyzed int `json:"files_analyzed"`

	// FunctionsAnalyzed count
	FunctionsAnalyzed int `json:"functions_analyzed"`

	// DurationMs of the analysis
	DurationMs int64 `json:"duration_ms"`

	// ModelUsed name
	ModelUsed string `json:"model_used,omitempty"`
}

// HealthResponse represents the plugin health check response
type HealthResponse struct {
	// Status of the plugin ("healthy", "degraded", "unhealthy")
	Status string `json:"status"`

	// Version of the plugin
	Version string `json:"version,omitempty"`

	// Capabilities supported by this plugin
	Capabilities []string `json:"capabilities,omitempty"`

	// Message with additional information
	Message string `json:"message,omitempty"`
}

// SemanticClonePlugin defines the interface for semantic clone detection plugins
type SemanticClonePlugin interface {
	// DetectClones analyzes code for semantic clones
	DetectClones(ctx context.Context, req *SemanticCloneRequest) (*SemanticCloneResponse, error)

	// Health checks if the plugin is available and healthy
	Health(ctx context.Context) (*HealthResponse, error)

	// IsAvailable returns true if the plugin is accessible
	IsAvailable() bool
}

// HTTPPluginClient is an HTTP-based plugin client
type HTTPPluginClient struct {
	baseURL    string
	httpClient *http.Client
	available  bool
}

// NewHTTPPluginClient creates a new HTTP plugin client
func NewHTTPPluginClient(baseURL string) *HTTPPluginClient {
	client := &HTTPPluginClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // Semantic analysis can be slow
		},
		available: false,
	}

	// Check availability on creation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	health, err := client.Health(ctx)
	if err == nil && health.Status == "healthy" {
		client.available = true
	}

	return client
}

// DetectClones sends a clone detection request to the plugin
func (c *HTTPPluginClient) DetectClones(ctx context.Context, req *SemanticCloneRequest) (*SemanticCloneResponse, error) {
	if !c.available {
		return nil, fmt.Errorf("plugin not available")
	}

	// Marshal request
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/detect", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		c.available = false // Mark as unavailable on error
		return nil, fmt.Errorf("plugin request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("plugin returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result SemanticCloneResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for error in response
	if result.Error != "" {
		return nil, fmt.Errorf("plugin error: %s", result.Error)
	}

	return &result, nil
}

// Health checks the plugin's health status
func (c *HTTPPluginClient) Health(ctx context.Context) (*HealthResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/health", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("health check failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return &HealthResponse{Status: "unhealthy"}, nil
	}

	var health HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return nil, fmt.Errorf("failed to decode health response: %w", err)
	}

	return &health, nil
}

// IsAvailable returns true if the plugin is accessible
func (c *HTTPPluginClient) IsAvailable() bool {
	return c.available
}

// NoOpPlugin is a plugin that does nothing (for graceful degradation)
type NoOpPlugin struct{}

// NewNoOpPlugin creates a no-op plugin
func NewNoOpPlugin() *NoOpPlugin {
	return &NoOpPlugin{}
}

// DetectClones returns empty results
func (p *NoOpPlugin) DetectClones(ctx context.Context, req *SemanticCloneRequest) (*SemanticCloneResponse, error) {
	return &SemanticCloneResponse{
		Clones: []SemanticClone{},
		Stats: SemanticCloneStats{
			FilesAnalyzed:     0,
			FunctionsAnalyzed: 0,
			DurationMs:        0,
		},
		Error: "semantic clone detection plugin not available",
	}, nil
}

// Health returns unhealthy status
func (p *NoOpPlugin) Health(ctx context.Context) (*HealthResponse, error) {
	return &HealthResponse{
		Status:  "unhealthy",
		Message: "semantic clone detection plugin not configured",
	}, nil
}

// IsAvailable always returns false
func (p *NoOpPlugin) IsAvailable() bool {
	return false
}
