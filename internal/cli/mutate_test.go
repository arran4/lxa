package cli

import (
	"bytes"
	"strings"
	"testing"
)

type mockStore struct {
	xattrs map[string]map[string][]byte
}

func (m *mockStore) List(path string) ([]string, error) {
	if m.xattrs == nil || m.xattrs[path] == nil {
		return nil, nil
	}
	var names []string
	for k := range m.xattrs[path] {
		names = append(names, k)
	}
	return names, nil
}

func (m *mockStore) Get(path, name string) ([]byte, error) {
	if m.xattrs == nil || m.xattrs[path] == nil {
		return nil, nil
	}
	return m.xattrs[path][name], nil
}

func (m *mockStore) Set(path, name string, data []byte) error {
	if m.xattrs == nil {
		m.xattrs = make(map[string]map[string][]byte)
	}
	if m.xattrs[path] == nil {
		m.xattrs[path] = make(map[string][]byte)
	}
	m.xattrs[path][name] = data
	return nil
}

func (m *mockStore) Remove(path, name string) error {
	if m.xattrs != nil && m.xattrs[path] != nil {
		delete(m.xattrs[path], name)
	}
	return nil
}

func TestMutateFlags(t *testing.T) {
	store := &mockStore{
		xattrs: map[string]map[string][]byte{
			"file1.txt": {
				"user.xdg.tags":    []byte("a, b\x00"),
				"user.xdg.comment": []byte("test\x00"),
			},
		},
	}

	var out, errOut bytes.Buffer
	err := Run([]string{"--add-tags=c, d", "--set-rating=5", "file1.txt"}, &out, &errOut, WithXattrStore(store))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tags := string(store.xattrs["file1.txt"]["user.xdg.tags"])
	if !strings.Contains(tags, "a") || !strings.Contains(tags, "c") || !strings.Contains(tags, "d") {
		t.Errorf("expected tags 'a,b,c,d', got %q", tags)
	}

	rating := string(store.xattrs["file1.txt"]["user.xdg.rating"])
	if rating != "5\x00" {
		t.Errorf("expected rating '5\\x00', got %q", rating)
	}
}

func TestMutateFlags_Clear(t *testing.T) {
	store := &mockStore{
		xattrs: map[string]map[string][]byte{
			"file1.txt": {
				"user.xdg.tags":    []byte("a, b\x00"),
				"user.xdg.comment": []byte("test\x00"),
			},
		},
	}

	var out, errOut bytes.Buffer
	err := Run([]string{"--clear-tags", "--clear-comment", "file1.txt"}, &out, &errOut, WithXattrStore(store))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := store.xattrs["file1.txt"]["user.xdg.tags"]; ok {
		t.Errorf("expected tags to be cleared")
	}

	if _, ok := store.xattrs["file1.txt"]["user.xdg.comment"]; ok {
		t.Errorf("expected comment to be cleared")
	}
}

func TestMutateFlags_Errors(t *testing.T) {
	var out, errOut bytes.Buffer
	err := Run([]string{"--clear-tags", "--set-tags=a", "file1.txt"}, &out, &errOut)
	if err == nil {
		t.Errorf("expected error for mutually exclusive flags")
	}

	err = Run([]string{"--set-rating=abc", "file1.txt"}, &out, &errOut)
	if err == nil {
		t.Errorf("expected error for invalid rating")
	}
}
