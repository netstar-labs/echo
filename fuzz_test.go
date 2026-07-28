package echo

import (
	"strings"
	"testing"
)

var fuzzSeeds = []string{
	"", "phone", "fone", "Schmidt", "Tymczak", "Ashcraft",
	"g00gle", "PayPal!", "münchen", "日本語", "\x00\xff\xfe",
	strings.Repeat("a", 500), strings.Repeat("SCH", 100),
}

// FuzzMetaphone asserts Double Metaphone never panics on arbitrary/Unicode/empty
// input, is deterministic, caps both codes at four characters, and returns "" for
// the secondary whenever it would equal the primary.
func FuzzMetaphone(f *testing.F) {
	for _, s := range fuzzSeeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		p, sec := Metaphone(s)
		if p2, sec2 := Metaphone(s); p2 != p || sec2 != sec {
			t.Fatalf("Metaphone(%q) not deterministic: (%q,%q) vs (%q,%q)", s, p, sec, p2, sec2)
		}
		if len(p) > maxCode || len(sec) > maxCode {
			t.Errorf("Metaphone(%q) over cap: (%q,%q)", s, p, sec)
		}
		if sec != "" && sec == p {
			t.Errorf("Metaphone(%q) secondary duplicates primary: %q", s, p)
		}
		_ = Keys(s)      // must not panic
		_ = Sounds(s, s) // must not panic
	})
}

// FuzzSoundex asserts Soundex never panics, is deterministic, and is well-formed:
// either empty (no ASCII letters) or exactly one letter followed by three digits.
func FuzzSoundex(f *testing.F) {
	for _, s := range fuzzSeeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		got := Soundex(s)
		if got2 := Soundex(s); got2 != got {
			t.Fatalf("Soundex(%q) not deterministic: %q vs %q", s, got, got2)
		}
		if got == "" {
			return
		}
		if len(got) != 4 {
			t.Errorf("Soundex(%q) = %q, want 4 chars", s, got)
		}
		if got[0] < 'A' || got[0] > 'Z' {
			t.Errorf("Soundex(%q) = %q, first char not A–Z", s, got)
		}
		for _, d := range got[1:] {
			if d < '0' || d > '6' {
				t.Errorf("Soundex(%q) = %q, non-Soundex digit", s, got)
			}
		}
	})
}
