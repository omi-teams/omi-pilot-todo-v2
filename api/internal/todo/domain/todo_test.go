package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestNormalizeTitle(t *testing.T) {
	for _, tc := range []struct {
		name, input, want string
		invalid           bool
	}{
		{"trim", " \tbuy milk\n", "buy milk", false},
		{"empty", "", "", true},
		{"whitespace", " \t\n\u2003", "", true},
		{"one character", "界", "界", false},
		{"200 characters", strings.Repeat("界", 200), strings.Repeat("界", 200), false},
		{"201 characters", strings.Repeat("界", 201), "", true},
		{"trim before count", "  " + strings.Repeat("a", 200) + "  ", strings.Repeat("a", 200), false},
		{"invalid utf8", string([]byte{0xff}), "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NormalizeTitle(tc.input)
			if got != tc.want || errors.Is(err, ErrInvalidTitle) != tc.invalid {
				t.Fatalf("got (%q, %v), want (%q, invalid=%v)", got, err, tc.want, tc.invalid)
			}
		})
	}
}
