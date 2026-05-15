package graph

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/config"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBuilder(t *testing.T) {
	layers := []config.Layer{
		{Name: "domain", Path: "src/domain/**", DependsOn: []string{}},
	}

	builder := NewBuilder("/test/path", layers)

	require.NotNil(t, builder)
	assert.Equal(t, "/test/path", builder.rootPath)
	assert.Equal(t, 1, len(builder.layers))
}

func TestBuild_EmptyFileList(t *testing.T) {
	builder := NewBuilder("/test", []config.Layer{})
	graph, err := builder.Build([]walker.FileInfo{})

	require.NoError(t, err)
	require.NotNil(t, graph)
	assert.Empty(t, graph.AllFiles)
}

func TestBuild_WithLayers(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "graph-test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	domainDir := filepath.Join(tmpDir, "src", "domain")
	require.NoError(t, os.MkdirAll(domainDir, 0755))

	userFile := filepath.Join(domainDir, "user.ts")
	require.NoError(t, os.WriteFile(userFile, []byte("export class User {}"), 0644))

	layers := []config.Layer{
		{Name: "domain", Path: "src/domain/**", DependsOn: []string{}},
		{Name: "application", Path: "src/application/**", DependsOn: []string{"domain"}},
	}

	builder := NewBuilder(tmpDir, layers)

	files := []walker.FileInfo{
		{Path: "src/domain/user.ts", AbsPath: userFile, IsDir: false},
	}

	graph, err := builder.Build(files)
	require.NoError(t, err)

	assert.Equal(t, 1, len(graph.AllFiles))

	layer := graph.GetLayerForFile("src/domain/user.ts")
	require.NotNil(t, layer)
	assert.Equal(t, "domain", layer.Name)
}

func TestMatchesLayerPath_SimplePrefix(t *testing.T) {
	builder := NewBuilder("/test", []config.Layer{})

	assert.True(t, builder.matchesLayerPath("src/domain/user.ts", "src/domain"))
	assert.False(t, builder.matchesLayerPath("src/application/service.ts", "src/domain"))
}

func TestMatchesLayerPath_Glob(t *testing.T) {
	builder := NewBuilder("/test", []config.Layer{})

	assert.True(t, builder.matchesLayerPath("src/domain/user.ts", "src/domain/**"))
	assert.True(t, builder.matchesLayerPath("src/domain/models/user.ts", "src/domain/**"))
	assert.False(t, builder.matchesLayerPath("src/application/service.ts", "src/domain/**"))
}

func TestMatchesLayerPath_GlobWithSuffix(t *testing.T) {
	builder := NewBuilder("/test", []config.Layer{})

	assert.True(t, builder.matchesLayerPath("src/domain/user.model.ts", "src/**/*.model.ts"))
}

func TestFindLayerForFile(t *testing.T) {
	layers := []config.Layer{
		{Name: "domain", Path: "src/domain/**", DependsOn: []string{}},
		{Name: "application", Path: "src/application/**", DependsOn: []string{"domain"}},
		{Name: "infrastructure", Path: "src/infrastructure/**", DependsOn: []string{"domain"}},
	}

	builder := NewBuilder("/test", layers)

	layer := builder.findLayerForFile("src/domain/user.ts")
	require.NotNil(t, layer)
	assert.Equal(t, "domain", layer.Name)

	layer = builder.findLayerForFile("src/application/userService.ts")
	require.NotNil(t, layer)
	assert.Equal(t, "application", layer.Name)

	layer = builder.findLayerForFile("src/other/file.ts")
	assert.Nil(t, layer)
}

func TestCanLayerDependOn_AllowedDependency(t *testing.T) {
	graph := &ImportGraph{
		Layers: []config.Layer{
			{Name: "presentation", Path: "src/presentation/**", DependsOn: []string{"application", "domain"}},
			{Name: "application", Path: "src/application/**", DependsOn: []string{"domain"}},
			{Name: "domain", Path: "src/domain/**", DependsOn: []string{}},
		},
	}

	presentation := graph.FindLayerByName("presentation")
	application := graph.FindLayerByName("application")
	domain := graph.FindLayerByName("domain")

	assert.True(t, graph.CanLayerDependOn(presentation, application))
	assert.True(t, graph.CanLayerDependOn(presentation, domain))
	assert.True(t, graph.CanLayerDependOn(application, domain))
}

func TestCanLayerDependOn_ForbiddenDependency(t *testing.T) {
	graph := &ImportGraph{
		Layers: []config.Layer{
			{Name: "presentation", Path: "src/presentation/**", DependsOn: []string{"application"}},
			{Name: "application", Path: "src/application/**", DependsOn: []string{"domain"}},
			{Name: "domain", Path: "src/domain/**", DependsOn: []string{}},
		},
	}

	presentation := graph.FindLayerByName("presentation")
	application := graph.FindLayerByName("application")
	domain := graph.FindLayerByName("domain")

	assert.False(t, graph.CanLayerDependOn(domain, application))
	assert.False(t, graph.CanLayerDependOn(domain, presentation))
	assert.False(t, graph.CanLayerDependOn(application, presentation))
}

func TestCanLayerDependOn_SameLayer(t *testing.T) {
	graph := &ImportGraph{
		Layers: []config.Layer{
			{Name: "domain", Path: "src/domain/**", DependsOn: []string{}},
		},
	}

	domain := graph.FindLayerByName("domain")

	assert.True(t, graph.CanLayerDependOn(domain, domain))
}

func TestCanLayerDependOn_Wildcard(t *testing.T) {
	graph := &ImportGraph{
		Layers: []config.Layer{
			{Name: "presentation", Path: "src/presentation/**", DependsOn: []string{"*"}},
			{Name: "application", Path: "src/application/**", DependsOn: []string{"domain"}},
			{Name: "domain", Path: "src/domain/**", DependsOn: []string{}},
		},
	}

	presentation := graph.FindLayerByName("presentation")
	application := graph.FindLayerByName("application")
	domain := graph.FindLayerByName("domain")

	assert.True(t, graph.CanLayerDependOn(presentation, application))
	assert.True(t, graph.CanLayerDependOn(presentation, domain))
}

func TestCanLayerDependOn_NilLayers(t *testing.T) {
	graph := &ImportGraph{}

	assert.True(t, graph.CanLayerDependOn(nil, nil))

	layer := &config.Layer{Name: "test", Path: "test/**", DependsOn: []string{}}

	assert.True(t, graph.CanLayerDependOn(layer, nil))
	assert.True(t, graph.CanLayerDependOn(nil, layer))
}

func TestGetLayerForFile(t *testing.T) {
	domain := &config.Layer{Name: "domain", Path: "src/domain/**", DependsOn: []string{}}

	graph := &ImportGraph{
		FileLayers: map[string]*config.Layer{
			"src/domain/user.ts": domain,
		},
	}

	layer := graph.GetLayerForFile("src/domain/user.ts")
	require.NotNil(t, layer)
	assert.Equal(t, "domain", layer.Name)

	layer = graph.GetLayerForFile("src/other/file.ts")
	assert.Nil(t, layer)
}

func TestGetDependencies(t *testing.T) {
	graph := &ImportGraph{
		Dependencies: map[string][]string{
			"src/presentation/component.ts": {"src/application/service", "src/domain/user"},
		},
	}

	deps := graph.GetDependencies("src/presentation/component.ts")
	assert.Equal(t, 2, len(deps))

	deps = graph.GetDependencies("src/other/file.ts")
	assert.Empty(t, deps)
}

func TestFindLayerByName(t *testing.T) {
	graph := &ImportGraph{
		Layers: []config.Layer{
			{Name: "domain", Path: "src/domain/**", DependsOn: []string{}},
			{Name: "application", Path: "src/application/**", DependsOn: []string{"domain"}},
		},
	}

	layer := graph.FindLayerByName("domain")
	require.NotNil(t, layer)
	assert.Equal(t, "domain", layer.Name)

	layer = graph.FindLayerByName("nonexistent")
	assert.Nil(t, layer)
}

func TestBuild_IncomingReferences(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "graph-ref-test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	srcDir := filepath.Join(tmpDir, "src")
	require.NoError(t, os.MkdirAll(srcDir, 0755))

	userFile := filepath.Join(srcDir, "user.ts")
	serviceFile := filepath.Join(srcDir, "service.ts")

	require.NoError(t, os.WriteFile(userFile, []byte("export class User {}"), 0644))
	require.NoError(t, os.WriteFile(serviceFile, []byte("import { User } from './user'"), 0644))

	builder := NewBuilder(tmpDir, []config.Layer{})

	files := []walker.FileInfo{
		{Path: filepath.Join("src", "user.ts"), AbsPath: userFile, IsDir: false},
		{Path: filepath.Join("src", "service.ts"), AbsPath: serviceFile, IsDir: false},
	}

	graph, err := builder.Build(files)
	require.NoError(t, err)

	deps := graph.GetDependencies(filepath.Join("src", "service.ts"))
	assert.NotEmpty(t, deps)

	assert.GreaterOrEqual(t, graph.IncomingRefs[filepath.Join("src", "user.ts")], 1)
}
