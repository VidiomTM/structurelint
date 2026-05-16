package predicate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/config"
	"github.com/Jonathangadeaharder/structurelint/internal/graph"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
	"github.com/stretchr/testify/assert"
)

func testFile(path string, isDir bool) walker.FileInfo {
	return walker.FileInfo{
		Path:    path,
		AbsPath: path,
		IsDir:   isDir,
		Depth:   2,
	}
}

func testCtx() *Context {
	return &Context{}
}

func TestNew_AlwaysTrue(t *testing.T) {
	p := New().Build()
	assert.True(t, p(testFile("any.go", false), testCtx()))
}

func TestWithPredicate(t *testing.T) {
	p := WithPredicate(func(f walker.FileInfo, ctx *Context) bool {
		return f.Path == "specific.go"
	}).Build()
	assert.True(t, p(testFile("specific.go", false), testCtx()))
	assert.False(t, p(testFile("other.go", false), testCtx()))
}

func TestAnd(t *testing.T) {
	isGo := HasExtension(".go")
	isFile := IsFile()
	p := New().And(isGo).And(isFile).Build()
	assert.True(t, p(testFile("main.go", false), testCtx()))
	assert.False(t, p(testFile("main.py", false), testCtx()))
	assert.False(t, p(testFile("dir", true), testCtx()))
}

func TestOr(t *testing.T) {
	isGo := HasExtension(".go")
	isPy := HasExtension(".py")
	p := WithPredicate(isGo).Or(isPy).Build()
	assert.True(t, p(testFile("main.go", false), testCtx()))
	assert.True(t, p(testFile("main.py", false), testCtx()))
	assert.False(t, p(testFile("main.rs", false), testCtx()))
}

func TestNot(t *testing.T) {
	isGo := HasExtension(".go")
	p := WithPredicate(isGo).Not().Build()
	assert.False(t, p(testFile("main.go", false), testCtx()))
	assert.True(t, p(testFile("main.py", false), testCtx()))
}

func TestPathMatches(t *testing.T) {
	p := PathMatches("src/*.go")
	assert.True(t, p(testFile("src/main.go", false), testCtx()))
	assert.False(t, p(testFile("src/sub/main.go", false), testCtx()))
}

func TestPathContains(t *testing.T) {
	p := PathContains("test")
	assert.True(t, p(testFile("src/test_main.go", false), testCtx()))
	assert.False(t, p(testFile("src/main.go", false), testCtx()))
}

func TestPathStartsWith(t *testing.T) {
	p := PathStartsWith("src/")
	assert.True(t, p(testFile("src/main.go", false), testCtx()))
	assert.False(t, p(testFile("lib/main.go", false), testCtx()))
}

func TestPathEndsWith(t *testing.T) {
	p := PathEndsWith("_test.go")
	assert.True(t, p(testFile("main_test.go", false), testCtx()))
	assert.False(t, p(testFile("main.go", false), testCtx()))
}

func TestPathRegex(t *testing.T) {
	p := PathRegex(`\.go$`)
	assert.True(t, p(testFile("main.go", false), testCtx()))
	assert.False(t, p(testFile("main.py", false), testCtx()))
}

func TestIsFile(t *testing.T) {
	p := IsFile()
	assert.True(t, p(testFile("main.go", false), testCtx()))
	assert.False(t, p(testFile("dir", true), testCtx()))
}

func TestIsDirectory(t *testing.T) {
	p := IsDirectory()
	assert.False(t, p(testFile("main.go", false), testCtx()))
	assert.True(t, p(testFile("dir", true), testCtx()))
}

func TestHasExtension(t *testing.T) {
	assert.True(t, HasExtension(".go")(testFile("main.go", false), testCtx()))
	assert.True(t, HasExtension("go")(testFile("main.go", false), testCtx()))
	assert.False(t, HasExtension(".py")(testFile("main.go", false), testCtx()))
}

func TestInLayer(t *testing.T) {
	layer := config.Layer{Name: "domain"}
	g := &graph.ImportGraph{
		FileLayers: map[string]*config.Layer{
			"src/domain/user.go": &layer,
		},
	}
	ctx := &Context{Graph: g}
	assert.True(t, InLayer("domain")(testFile("src/domain/user.go", false), ctx))
	assert.False(t, InLayer("app")(testFile("src/domain/user.go", false), ctx))
}

func TestInLayer_NilGraph(t *testing.T) {
	assert.False(t, InLayer("domain")(testFile("file.go", false), testCtx()))
}

func TestHasLayer(t *testing.T) {
	layer := config.Layer{Name: "domain"}
	g := &graph.ImportGraph{
		FileLayers: map[string]*config.Layer{
			"src/domain/user.go": &layer,
		},
	}
	ctx := &Context{Graph: g}
	assert.True(t, HasLayer()(testFile("src/domain/user.go", false), ctx))
	assert.False(t, HasLayer()(testFile("other.go", false), ctx))
}

func TestHasLayer_NilGraph(t *testing.T) {
	assert.False(t, HasLayer()(testFile("file.go", false), testCtx()))
}

func TestDependsOn(t *testing.T) {
	g := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"src/app/service.go": {"src/domain/user", "src/domain/product"},
		},
	}
	ctx := &Context{Graph: g}
	p := DependsOn("src/domain/*")
	assert.True(t, p(testFile("src/app/service.go", false), ctx))
	assert.False(t, p(testFile("other.go", false), ctx))
}

func TestDependsOn_NilGraph(t *testing.T) {
	assert.False(t, DependsOn("*")(testFile("file.go", false), testCtx()))
}

func TestHasDependencies(t *testing.T) {
	g := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"has_deps.go":  {"dep1", "dep2"},
			"no_deps.go":   {},
		},
	}
	ctx := &Context{Graph: g}
	assert.True(t, HasDependencies()(testFile("has_deps.go", false), ctx))
	assert.False(t, HasDependencies()(testFile("no_deps.go", false), ctx))
}

func TestHasDependencies_NilGraph(t *testing.T) {
	assert.False(t, HasDependencies()(testFile("file.go", false), testCtx()))
}

func TestHasIncomingRefs(t *testing.T) {
	g := &graph.ImportGraph{
		IncomingRefs: map[string]int{
			"popular.go":  5,
			"lonely.go":   0,
		},
	}
	ctx := &Context{Graph: g}
	assert.True(t, HasIncomingRefs()(testFile("popular.go", false), ctx))
	assert.False(t, HasIncomingRefs()(testFile("lonely.go", false), ctx))
}

func TestHasIncomingRefs_NilGraph(t *testing.T) {
	assert.False(t, HasIncomingRefs()(testFile("file.go", false), testCtx()))
}

func TestIsOrphaned(t *testing.T) {
	g := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"orphan.go":  {},
			"popular.go": {"dep"},
		},
		IncomingRefs: map[string]int{
			"orphan.go":  0,
			"popular.go": 3,
		},
	}
	ctx := &Context{Graph: g}
	assert.True(t, IsOrphaned()(testFile("orphan.go", false), ctx))
	assert.False(t, IsOrphaned()(testFile("popular.go", false), ctx))
}

func TestIsOrphaned_NilGraph(t *testing.T) {
	assert.False(t, IsOrphaned()(testFile("file.go", false), testCtx()))
}

func TestSizeGreaterThan(t *testing.T) {
	dir := t.TempDir()
	smallFile := filepath.Join(dir, "small.go")
	largeFile := filepath.Join(dir, "large.go")
	os.WriteFile(smallFile, []byte("small"), 0644)
	os.WriteFile(largeFile, []byte("this is a larger file with more content"), 0644)

	p := SizeGreaterThan(10)
	assert.True(t, p(testFile(largeFile, false), &Context{}))
	assert.False(t, p(testFile(smallFile, false), &Context{}))
}

func TestSizeGreaterThan_Error(t *testing.T) {
	p := SizeGreaterThan(10)
	assert.False(t, p(testFile("/nonexistent/file.go", false), testCtx()))
}

func TestSizeGreaterThan_Dir(t *testing.T) {
	dir := t.TempDir()
	p := SizeGreaterThan(0)
	assert.False(t, p(testFile(dir, true), &Context{}))
}

func TestSizeLessThan(t *testing.T) {
	dir := t.TempDir()
	smallFile := filepath.Join(dir, "small.go")
	largeFile := filepath.Join(dir, "large.go")
	os.WriteFile(smallFile, []byte("small"), 0644)
	os.WriteFile(largeFile, []byte("this is a larger file with much more content in it"), 0644)

	p := SizeLessThan(20)
	assert.True(t, p(testFile(smallFile, false), &Context{}))
	assert.False(t, p(testFile(largeFile, false), &Context{}))
}

func TestSizeLessThan_Error(t *testing.T) {
	p := SizeLessThan(100)
	assert.False(t, p(testFile("/nonexistent", false), testCtx()))
}

func TestSizeLessThan_Dir(t *testing.T) {
	dir := t.TempDir()
	assert.False(t, SizeLessThan(1000)(testFile(dir, true), &Context{}))
}

func TestDepthEquals(t *testing.T) {
	p := DepthEquals(2)
	assert.True(t, p(testFile("file.go", false), testCtx()))

	p1 := DepthEquals(3)
	assert.False(t, p1(testFile("file.go", false), testCtx()))
}

func TestDepthGreaterThan(t *testing.T) {
	f := testFile("file.go", false)
	f.Depth = 3
	assert.True(t, DepthGreaterThan(2)(f, testCtx()))
	assert.False(t, DepthGreaterThan(5)(f, testCtx()))
}

func TestDepthLessThan(t *testing.T) {
	f := testFile("file.go", false)
	f.Depth = 1
	assert.True(t, DepthLessThan(2)(f, testCtx()))
	assert.False(t, DepthLessThan(1)(f, testCtx()))
}

func TestNameMatches(t *testing.T) {
	p := NameMatches("*.go")
	assert.True(t, p(testFile("main.go", false), testCtx()))
	assert.False(t, p(testFile("main.py", false), testCtx()))
}

func TestNameContains(t *testing.T) {
	p := NameContains("test")
	assert.True(t, p(testFile("src/test_main.go", false), testCtx()))
	assert.False(t, p(testFile("src/main.go", false), testCtx()))
}

func TestNameRegex(t *testing.T) {
	p := NameRegex(`test.*\.go`)
	assert.True(t, p(testFile("src/test_util.go", false), testCtx()))
	assert.False(t, p(testFile("main.go", false), testCtx()))
}

func TestAll(t *testing.T) {
	isGo := HasExtension(".go")
	isNotDir := IsFile()
	p := All(isGo, isNotDir)
	assert.True(t, p(testFile("main.go", false), testCtx()))
	assert.False(t, p(testFile("main.py", false), testCtx()))
}

func TestAny(t *testing.T) {
	isGo := HasExtension(".go")
	isPy := HasExtension(".py")
	p := Any(isGo, isPy)
	assert.True(t, p(testFile("main.go", false), testCtx()))
	assert.True(t, p(testFile("main.py", false), testCtx()))
	assert.False(t, p(testFile("main.rs", false), testCtx()))
}

func TestNone(t *testing.T) {
	isGo := HasExtension(".go")
	isPy := HasExtension(".py")
	p := None(isGo, isPy)
	assert.False(t, p(testFile("main.go", false), testCtx()))
	assert.True(t, p(testFile("main.rs", false), testCtx()))
}

func TestNotPredicate(t *testing.T) {
	isGo := HasExtension(".go")
	p := Not(isGo)
	assert.False(t, p(testFile("main.go", false), testCtx()))
	assert.True(t, p(testFile("main.py", false), testCtx()))
}

func TestCustom(t *testing.T) {
	p := Custom(func(file walker.FileInfo, ctx *Context) bool {
		return file.Depth > 0
	})
	assert.True(t, p(testFile("main.go", false), testCtx()))
}
