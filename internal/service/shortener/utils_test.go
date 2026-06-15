package shortener

import "testing"

func TestEncodeBase62(t *testing.T) {
	tests := []struct {
		name string
		in   uint64
		want string
	}{
		{"zero", 0, "a"},
		{"one", 1, "b"},
		{"max-single-digit", 61, "9"},
		{"two-digits", 62, "ba"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := encodeBase62(tt.in); got != tt.want {
				t.Errorf("encodeBase62(%d) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestShortenURLDeterministicAndLength(t *testing.T) {
	const url = "https://example.com/foo?bar=baz"

	first := shortenURL(url)
	if first != shortenURL(url) {
		t.Fatalf("shortenURL is not deterministic for the same input")
	}
	if l := len(first); l == 0 || l > 8 {
		t.Fatalf("len(%q) = %d, want between 1 and 8", first, l)
	}
}
