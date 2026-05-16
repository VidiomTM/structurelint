package ci

import (
	"fmt"
	"os"
)

type FileReader interface {
	ReadFile(path string) ([]byte, error)
}

type OSFileReader struct{}

func (OSFileReader) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

type MockFileReader struct {
	Files map[string]string
}

func (m MockFileReader) ReadFile(path string) ([]byte, error) {
	content, ok := m.Files[path]
	if !ok {
		return nil, fmt.Errorf("file not found: %s", path)
	}
	return []byte(content), nil
}
