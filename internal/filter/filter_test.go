package filter

import (
	"testing"

	"github.com/lxa-project/lxa/internal/xattr"
)

func TestFilter(t *testing.T) {
	cases := []struct {
		name     string
		expr     string
		metadata xattr.Metadata
		expected bool
	}{
		{
			name:     "empty filter allows all",
			expr:     "",
			metadata: xattr.Metadata{},
			expected: true,
		},
		{
			name:     "has:tags matches",
			expr:     "has:tags",
			metadata: xattr.Metadata{HasTags: true},
			expected: true,
		},
		{
			name:     "has:tags fails",
			expr:     "has:tags",
			metadata: xattr.Metadata{HasTags: false},
			expected: false,
		},
		{
			name:     "not has:comment",
			expr:     "not has:comment",
			metadata: xattr.Metadata{HasCmnt: false},
			expected: true,
		},
		{
			name:     "and condition",
			expr:     "has:tags and has:comment",
			metadata: xattr.Metadata{HasTags: true, HasCmnt: true},
			expected: true,
		},
		{
			name:     "or condition",
			expr:     "tag:urgent or tag:projectX",
			metadata: xattr.Metadata{Tags: []string{"projectX"}},
			expected: true,
		},
		{
			name:     "grouping",
			expr:     "( tag:urgent or tag:projectX ) and has:comment",
			metadata: xattr.Metadata{Tags: []string{"projectX"}, HasCmnt: true},
			expected: true,
		},
		{
			name:     "grouping fail",
			expr:     "( tag:urgent or tag:projectX ) and has:comment",
			metadata: xattr.Metadata{Tags: []string{"projectX"}, HasCmnt: false},
			expected: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			eval, err := NewEvaluator(tc.expr)
			if err != nil {
				t.Fatalf("Failed to parse %q: %v", tc.expr, err)
			}
			result := eval.Eval(tc.metadata)
			if result != tc.expected {
				t.Errorf("Eval(%q) = %v; want %v", tc.expr, result, tc.expected)
			}
		})
	}
}
