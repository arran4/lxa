package xattr

import (
	"reflect"
	"testing"
)

// mockReader implements Reader for tests.
type mockReader struct {
	xattrs map[string][]byte
}

func (m *mockReader) List(path string) ([]string, error) {
	var names []string
	for k := range m.xattrs {
		names = append(names, k)
	}
	return names, nil
}

func (m *mockReader) Get(path, name string) ([]byte, error) {
	return m.xattrs[name], nil
}

func TestReadMetadata(t *testing.T) {
	reader := &mockReader{
		xattrs: map[string][]byte{
			"user.xdg.tags":    []byte("projectX, urgent\x00"),
			"user.xdg.comment": []byte("Need to fix this\x00"),
			"user.other":       []byte("test"),
		},
	}

	md, err := ReadMetadata(reader, "testfile")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !md.HasXDG {
		t.Errorf("expected HasXDG to be true")
	}
	if !md.HasTags {
		t.Errorf("expected HasTags to be true")
	}
	if !md.HasCmnt {
		t.Errorf("expected HasCmnt to be true")
	}

	expectedTags := []string{"projectX", "urgent"}
	if !reflect.DeepEqual(md.Tags, expectedTags) {
		t.Errorf("expected tags %v, got %v", expectedTags, md.Tags)
	}

	if md.Comment != "Need to fix this" {
		t.Errorf("expected comment 'Need to fix this', got %q", md.Comment)
	}
}
