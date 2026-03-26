package xattr

import (
	"errors"
	"strings"
	"syscall"
	"unsafe"
)

// Store defines an interface for reading and writing xattrs.
type Store interface {
	Reader
	Set(path, name string, data []byte) error
	Remove(path, name string) error
}

// Reader defines an interface for reading xattrs.
type Reader interface {
	List(path string) ([]string, error)
	Get(path, name string) ([]byte, error)
}

// SyscallReader implements Reader using native Linux syscalls.
type SyscallReader struct{}

// NewSyscallReader creates a new Reader based on syscalls.
func NewSyscallReader() Reader {
	return &SyscallReader{}
}

func llistxattr(path string, dest []byte) (sz int, err error) {
	var _p0 *byte
	_p0, err = syscall.BytePtrFromString(path)
	if err != nil {
		return 0, err
	}
	var _p1 unsafe.Pointer
	if len(dest) > 0 {
		_p1 = unsafe.Pointer(&dest[0])
	} else {
		_p1 = unsafe.Pointer(nil)
	}
	r0, _, e1 := syscall.Syscall(syscall.SYS_LLISTXATTR, uintptr(unsafe.Pointer(_p0)), uintptr(_p1), uintptr(len(dest)))
	sz = int(r0)
	if e1 != 0 {
		err = e1
	}
	return
}

func lsetxattr(path string, name string, data []byte, flags int) (err error) {
	var _p0 *byte
	_p0, err = syscall.BytePtrFromString(path)
	if err != nil {
		return err
	}
	var _p1 *byte
	_p1, err = syscall.BytePtrFromString(name)
	if err != nil {
		return err
	}
	var _p2 unsafe.Pointer
	if len(data) > 0 {
		_p2 = unsafe.Pointer(&data[0])
	} else {
		_p2 = unsafe.Pointer(nil)
	}
	_, _, e1 := syscall.Syscall6(syscall.SYS_LSETXATTR, uintptr(unsafe.Pointer(_p0)), uintptr(unsafe.Pointer(_p1)), uintptr(_p2), uintptr(len(data)), uintptr(flags), 0)
	if e1 != 0 {
		err = e1
	}
	return
}

func lremovexattr(path string, name string) (err error) {
	var _p0 *byte
	_p0, err = syscall.BytePtrFromString(path)
	if err != nil {
		return err
	}
	var _p1 *byte
	_p1, err = syscall.BytePtrFromString(name)
	if err != nil {
		return err
	}
	_, _, e1 := syscall.Syscall(syscall.SYS_LREMOVEXATTR, uintptr(unsafe.Pointer(_p0)), uintptr(unsafe.Pointer(_p1)), 0)
	if e1 != 0 {
		err = e1
	}
	return
}

func lgetxattr(path string, name string, dest []byte) (sz int, err error) {
	var _p0 *byte
	_p0, err = syscall.BytePtrFromString(path)
	if err != nil {
		return 0, err
	}
	var _p1 *byte
	_p1, err = syscall.BytePtrFromString(name)
	if err != nil {
		return 0, err
	}
	var _p2 unsafe.Pointer
	if len(dest) > 0 {
		_p2 = unsafe.Pointer(&dest[0])
	} else {
		_p2 = unsafe.Pointer(nil)
	}
	r0, _, e1 := syscall.Syscall6(syscall.SYS_LGETXATTR, uintptr(unsafe.Pointer(_p0)), uintptr(unsafe.Pointer(_p1)), uintptr(_p2), uintptr(len(dest)), 0, 0)
	sz = int(r0)
	if e1 != 0 {
		err = e1
	}
	return
}

// List returns a list of xattr names for a given file path.
func (r *SyscallReader) List(path string) ([]string, error) {
	// 1. Get size
	sz, err := llistxattr(path, nil)
	if err != nil {
		if errors.Is(err, syscall.ENODATA) || errors.Is(err, syscall.ENOTSUP) || errors.Is(err, syscall.EOPNOTSUPP) {
			return nil, nil // No attributes
		}
		return nil, err
	}
	if sz <= 0 {
		return nil, nil
	}

	// 2. Read names
	buf := make([]byte, sz)
	sz, err = llistxattr(path, buf)
	if err != nil {
		return nil, err
	}
	if sz <= 0 {
		return nil, nil
	}

	// 3. Parse null-terminated strings
	var names []string
	start := 0
	for i := 0; i < sz; i++ {
		if buf[i] == 0 {
			if i > start {
				names = append(names, string(buf[start:i]))
			}
			start = i + 1
		}
	}
	return names, nil
}

// Set writes the value of an xattr.
func (r *SyscallReader) Set(path, name string, data []byte) error {
	return lsetxattr(path, name, data, 0)
}

// Remove deletes an xattr.
func (r *SyscallReader) Remove(path, name string) error {
	return lremovexattr(path, name)
}

// Get reads the value of an xattr.
func (r *SyscallReader) Get(path, name string) ([]byte, error) {
	// 1. Get size
	sz, err := lgetxattr(path, name, nil)
	if err != nil {
		if errors.Is(err, syscall.ENODATA) || errors.Is(err, syscall.ENOTSUP) || errors.Is(err, syscall.EOPNOTSUPP) {
			return nil, nil
		}
		return nil, err
	}
	if sz <= 0 {
		return []byte{}, nil
	}

	// 2. Read value
	buf := make([]byte, sz)
	sz, err = lgetxattr(path, name, buf)
	if err != nil {
		return nil, err
	}
	return buf[:sz], nil
}

// Metadata holds parsed xattrs.
type Metadata struct {
	All          map[string][]byte
	XDG          map[string][]byte
	Tags         []string // parsed from user.xdg.tags
	Comment      string   // parsed from user.xdg.comment
	Rating       string   // parsed from user.xdg.rating
	HasXDG       bool
	HasTags      bool
	HasCmnt      bool
	HasRating    bool
	SELinux      string   // parsed from security.selinux
	DOSAttrib    string   // parsed from user.DOSATTRIB
	Capabilities []byte   // parsed from security.capability
	ACL          string   // indicates presence of POSIX ACLs
}

// ReadMetadata reads and parses all xattrs.
func ReadMetadata(r Reader, path string) (Metadata, error) {
	md := Metadata{
		All: make(map[string][]byte),
		XDG: make(map[string][]byte),
	}

	names, err := r.List(path)
	if err != nil {
		return md, err
	}

	for _, name := range names {
		val, err := r.Get(path, name)
		if err != nil {
			continue // Skip errors for individual attributes
		}
		md.All[name] = val

		if strings.HasPrefix(name, "user.xdg.") {
			md.HasXDG = true
			md.XDG[name] = val

			if name == "user.xdg.tags" {
				md.HasTags = true
				if len(val) > 0 {
					// Parse tags (comma separated)
					strVal := string(val)
					// Remove trailing null byte if present
					if strings.HasSuffix(strVal, "\x00") {
						strVal = strVal[:len(strVal)-1]
					}
					parts := strings.Split(strVal, ",")
					for _, p := range parts {
						p = strings.TrimSpace(p)
						if p != "" {
							md.Tags = append(md.Tags, p)
						}
					}
				}
			} else if name == "user.xdg.comment" {
				md.HasCmnt = true
				strVal := string(val)
				if strings.HasSuffix(strVal, "\x00") {
					strVal = strVal[:len(strVal)-1]
				}
				md.Comment = strVal
			} else if name == "user.xdg.rating" {
				md.HasRating = true
				strVal := string(val)
				if strings.HasSuffix(strVal, "\x00") {
					strVal = strVal[:len(strVal)-1]
				}
				md.Rating = strVal
			}
		}

		// Other standard attributes
		if name == "security.selinux" {
			strVal := string(val)
			if strings.HasSuffix(strVal, "\x00") {
				strVal = strVal[:len(strVal)-1]
			}
			md.SELinux = strVal
		} else if name == "user.DOSATTRIB" {
			strVal := string(val)
			if strings.HasSuffix(strVal, "\x00") {
				strVal = strVal[:len(strVal)-1]
			}
			md.DOSAttrib = strVal
		} else if name == "security.capability" {
			md.Capabilities = val
		} else if name == "system.posix_acl_access" {
			if md.ACL == "" {
				md.ACL = "access"
			} else if md.ACL == "default" {
				md.ACL = "both"
			}
		} else if name == "system.posix_acl_default" {
			if md.ACL == "" {
				md.ACL = "default"
			} else if md.ACL == "access" {
				md.ACL = "both"
			}
		}
	}

	return md, nil
}
