package tmm

import (
	"testing"
	"time"
)

func TestExpired(t *testing.T) {
	s := Session{
		lastreset: time.Now().Add(-10 * time.Minute),
	}

	if s.Expired() != true {
		t.Errorf("session should be expired")
	}

	s.lastreset = time.Now()

	if s.Expired() != false {
		t.Errorf("session should NOT be expired")
	}
}

func TestJoin(t *testing.T) {
	tests := []struct {
		name string
		b    string
		n    []string
		want string
	}{
		{"base with one extra", "https://example.com", []string{"mypath"}, "https://example.com/mypath"},
		{"base with no extra", "https://example.com", []string{}, "https://example.com"},
		{"base with three extra", "https://example.com", []string{"one", "two", "three"}, "https://example.com/one/two/three"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := join(tt.b, tt.n...)
			if s != tt.want {
				t.Errorf("Got %s, want %s", s, tt.want)
			}
		})
	}
}
