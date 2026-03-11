package annotation

import "testing"

func boolPtr(b bool) *bool { return &b }

func TestIsResolved(t *testing.T) {
	tests := []struct {
		name string
		a    Annotation
		want bool
	}{
		{"nil resolved", Annotation{}, false},
		{"resolved false", Annotation{Resolved: boolPtr(false)}, false},
		{"resolved true", Annotation{Resolved: boolPtr(true)}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsResolved(tt.a); got != tt.want {
				t.Errorf("IsResolved() = %v, want %v", got, tt.want)
			}
		})
	}
}
