package xattr

import (
	"reflect"
	"testing"
)

// mockReader implements Store for tests.
type mockReader struct {
	xattrs map[string][]byte
}

func (m *mockReader) Set(path, name string, data []byte) error {
	if m.xattrs == nil {
		m.xattrs = make(map[string][]byte)
	}
	m.xattrs[name] = data
	return nil
}

func (m *mockReader) Remove(path, name string) error {
	if m.xattrs != nil {
		delete(m.xattrs, name)
	}
	return nil
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

func TestReadMetadata_NewAttrs(t *testing.T) {
	reader := &mockReader{
		xattrs: map[string][]byte{
			"security.selinux":        []byte("unconfined_u:object_r:user_home_t:s0\x00"),
			"user.DOSATTRIB":          []byte("0x20\x00"),
			"security.capability":     []byte{1, 0, 0, 0},
			"system.posix_acl_access": []byte{2, 0, 0, 0},
			"user.xdg.rating":         []byte("5\x00"),
		},
	}

	md, err := ReadMetadata(reader, "testfile")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if md.SELinux != "unconfined_u:object_r:user_home_t:s0" {
		t.Errorf("expected SELinux 'unconfined_u:object_r:user_home_t:s0', got %q", md.SELinux)
	}
	if md.DOSAttrib != "0x20" {
		t.Errorf("expected DOSAttrib '0x20', got %q", md.DOSAttrib)
	}
	if !reflect.DeepEqual(md.Capabilities, []byte{1, 0, 0, 0}) {
		t.Errorf("expected Capabilities [1 0 0 0], got %v", md.Capabilities)
	}
	if md.ACL != "access" {
		t.Errorf("expected ACL 'access', got %q", md.ACL)
	}
	if md.Rating != "5" {
		t.Errorf("expected Rating '5', got %q", md.Rating)
	}
	if !md.HasRating {
		t.Errorf("expected HasRating to be true")
	}
}

func TestMockReaderMutations(t *testing.T) {
	reader := &mockReader{
		xattrs: make(map[string][]byte),
	}

	_ = reader.Set("test", "user.xdg.tags", []byte("a, b"))
	if string(reader.xattrs["user.xdg.tags"]) != "a, b" {
		t.Errorf("set failed")
	}

	_ = reader.Remove("test", "user.xdg.tags")
	if _, ok := reader.xattrs["user.xdg.tags"]; ok {
		t.Errorf("remove failed")
	}
}
