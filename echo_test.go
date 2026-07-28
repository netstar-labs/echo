package echo

import (
	"reflect"
	"testing"
)

// TestSoundexGolden pins the standard (NARA) Soundex against the canonical
// reference vectors. Ashcraft (H bridges two 2-codes → collapse) and Tymczak
// (a vowel between two 2-codes → coded twice, T522 not T520) are the two rows that
// separate a correct implementation from a naive one; Pfister (first letter P and
// following F share code 1 → the F is dropped) exercises the first-letter seed.
func TestSoundexGolden(t *testing.T) {
	cases := map[string]string{
		"Robert":   "R163",
		"Rupert":   "R163",
		"Rubin":    "R150",
		"Ashcraft": "A261",
		"Tymczak":  "T522",
		"Pfister":  "P236",
		"Honeyman": "H555",
	}
	for in, want := range cases {
		if got := Soundex(in); got != want {
			t.Errorf("Soundex(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSoundexEmpty(t *testing.T) {
	for _, in := range []string{"", "123", "!!!", "……"} {
		if got := Soundex(in); got != "" {
			t.Errorf("Soundex(%q) = %q, want empty", in, got)
		}
	}
}

// TestMetaphoneGolden pins Double Metaphone against words whose canonical codes are
// well established, covering the load-bearing rules: PH→F and CK→K (implicit in the
// homophone tests), TH→"0"/"T" (Thumb), silent initial letters (Knight/Wright/
// Pneumatic/Psalm), silent -MB (Dumb/Thumb), DG→J (Edge) vs DG→TK (Edgar), the
// Germanic -ACH- (Bach), CC→KS (accident), and the primary/secondary split
// (Smith, Schmidt, Michael).
//
// NOTE on "Thompson": the brand roadmap lists "TMSN", but canonical Double Metaphone
// has no silent-P rule for "MPS", so the P is coded and the four-char cap yields
// "TMPS". This asserts the algorithm's actual (correct) output; see the repo report
// for the discrepancy flagged to the audit.
func TestMetaphoneGolden(t *testing.T) {
	type code struct{ primary, secondary string }
	cases := map[string]code{
		"Smith":     {"SM0", "XMT"},
		"Schmidt":   {"XMT", "SMT"},
		"Michael":   {"MKL", "MXL"},
		"Edge":      {"AJ", ""},
		"Edgar":     {"ATKR", ""},
		"Pneumatic": {"NMTK", ""},
		"Wright":    {"RT", ""},
		"Knight":    {"NT", ""},
		"Psalm":     {"SLM", ""},
		"Thumb":     {"0M", "TM"},
		"Dumb":      {"TM", ""},
		"Bach":      {"PK", ""},
		"accident":  {"AKST", ""},
		"Xavier":    {"SF", "SFR"},
		"Thompson":  {"TMPS", ""}, // canonical output; roadmap's "TMSN" is from another coder
		// Rare-rule pins (audit follow-up): soft-CH, initial silent-GH, soft-G before
		// a front vowel, double-Z, and the Polish -SKI + Slavic SZ branches.
		"chef":      {"XF", ""},
		"ghost":     {"KST", ""},
		"gina":      {"KN", "JN"},
		"buzz":      {"PS", ""},
		"szymanski": {"SMNS", "XMNS"},
	}
	for in, want := range cases {
		p, s := Metaphone(in)
		if p != want.primary || s != want.secondary {
			t.Errorf("Metaphone(%q) = (%q, %q), want (%q, %q)", in, p, s, want.primary, want.secondary)
		}
	}
}

// TestMetaphoneSecondaryEmpty confirms the secondary code is "" for a word with a
// single pronunciation, and non-empty (and distinct) when the word forks.
func TestMetaphoneSecondaryEmpty(t *testing.T) {
	if _, s := Metaphone("phone"); s != "" {
		t.Errorf("Metaphone(phone) secondary = %q, want empty", s)
	}
	if p, s := Metaphone("Schmidt"); s == "" || s == p {
		t.Errorf("Metaphone(Schmidt) = (%q, %q); want a distinct non-empty secondary", p, s)
	}
}

func TestMetaphoneEmpty(t *testing.T) {
	for _, in := range []string{"", "123", "!!!", "……"} {
		if p, s := Metaphone(in); p != "" || s != "" {
			t.Errorf("Metaphone(%q) = (%q, %q), want empty", in, p, s)
		}
	}
}

// TestSoundsHomophones is the library's reason to exist: labels that sound alike but
// are spelled arbitrarily far apart must share a phonetic code. Each true pair is a
// different rule — PH→F (fone/phone), Q/KW and CK→K (kwik/quick), silent GH
// (night/nite), vowel-collapse (bare/bear, sea/see) — and clearly-unrelated labels
// must not match.
func TestSoundsHomophones(t *testing.T) {
	match := [][2]string{
		{"fone", "phone"},
		{"kwik", "quick"},
		{"night", "nite"},
		{"bare", "bear"},
		{"sea", "see"},
		{"catherine", "kathryn"}, // TH→0, both fork identically
		{"google", "gooogle"},    // repeated-vowel typo still sounds identical
	}
	for _, p := range match {
		if !Sounds(p[0], p[1]) {
			ka, kb := Keys(p[0]), Keys(p[1])
			t.Errorf("Sounds(%q, %q) = false; keys %v vs %v", p[0], p[1], ka, kb)
		}
	}

	noMatch := [][2]string{
		{"paypal", "microsoft"}, // clearly different
		{"paypal", "paypai"},    // an -l/-i visual swap is not a sound-alike
		{"amazon", "google"},
	}
	for _, p := range noMatch {
		if Sounds(p[0], p[1]) {
			t.Errorf("Sounds(%q, %q) = true; want false (keys %v vs %v)", p[0], p[1], Keys(p[0]), Keys(p[1]))
		}
	}
}

// TestSoundsSymmetric asserts Sounds is order-independent.
func TestSoundsSymmetric(t *testing.T) {
	pairs := [][2]string{
		{"fone", "phone"}, {"paypal", "microsoft"}, {"kwik", "quick"}, {"a", "b"},
	}
	for _, p := range pairs {
		if Sounds(p[0], p[1]) != Sounds(p[1], p[0]) {
			t.Errorf("Sounds(%q,%q)=%v but Sounds(%q,%q)=%v — not symmetric",
				p[0], p[1], Sounds(p[0], p[1]), p[1], p[0], Sounds(p[1], p[0]))
		}
	}
}

// TestSoundsEmpty: a label with no phonetic content never matches, even against
// itself, so an empty/garbage candidate cannot spuriously flag a brand.
func TestSoundsEmpty(t *testing.T) {
	for _, p := range [][2]string{{"", ""}, {"", "phone"}, {"123", "phone"}, {"!!!", "!!!"}} {
		if Sounds(p[0], p[1]) {
			t.Errorf("Sounds(%q, %q) = true, want false", p[0], p[1])
		}
	}
}

// TestKeys checks the canonical-key helper: one key when the word has a single
// pronunciation, two distinct keys when it forks, nil for no phonetic content.
func TestKeys(t *testing.T) {
	if got := Keys("phone"); !reflect.DeepEqual(got, []string{"FN"}) {
		t.Errorf("Keys(phone) = %v, want [FN]", got)
	}
	if got := Keys("Schmidt"); !reflect.DeepEqual(got, []string{"XMT", "SMT"}) {
		t.Errorf("Keys(Schmidt) = %v, want [XMT SMT]", got)
	}
	if got := Keys(""); got != nil {
		t.Errorf("Keys(\"\") = %v, want nil", got)
	}
}

// TestMetaphoneCorpus drives the engine's language-specific special cases —
// Slavic (WICZ/WITZ/CZ), Spanish (JOSE/SAN/-LLO/CIA), French (-IER/-OIS/-EAUX),
// Germanic/Greek (SCH/-ACH-/CH-), Italian (GG/SIO), and pinyin (ZH) — and asserts
// the engine invariants over each: deterministic, both codes within the four-char
// cap, a non-empty primary of code letters, and a secondary that never merely
// duplicates the primary. It does not pin exact literals for the rarer rules (whose
// canonical codes are less universally published) — those are covered by property
// tests; this guards that the branches run and stay well-formed.
func TestMetaphoneCorpus(t *testing.T) {
	corpus := []string{
		"Filipowicz", "Horowitz", "Czerny", "Tchaikovsky", // Slavic
		"Jose", "San Jacinto", "cabrillo", "gallegos", "focaccia", "Chianti", // Spanish/Italian
		"Rogier", "artois", "breaux", "Beaux", "Villier", // French
		"Schmidt", "Schneider", "school", "schooner", "Bacher", "Machen", // Germanic
		"chorus", "chemistry", "Michael", "Charac", "architect", // Greek
		"biaggi", "reggie", "Focaccia", "Sicilia", // Italian soft/hard G, SIO
		"Zhao", "Zhang", // pinyin
		"McHugh", "McCarthy", "Hugh", "laugh", "cough", "through", // GH
		"Caesar", "Xavier", "Yankelovich", "island", "Wasserman", "Arnow",
		"campbell", "raspberry", "accident", "succeed", "bacchus",
		"Thomas", "Thames", "nation", "Ashcraft", "Tymczak",
	}
	for _, w := range corpus {
		p, s := Metaphone(w)
		if p2, s2 := Metaphone(w); p2 != p || s2 != s {
			t.Errorf("Metaphone(%q) not deterministic: (%q,%q) vs (%q,%q)", w, p, s, p2, s2)
		}
		if len(p) > maxCode || len(s) > maxCode {
			t.Errorf("Metaphone(%q) = (%q,%q) exceeds %d-char cap", w, p, s, maxCode)
		}
		if p == "" {
			t.Errorf("Metaphone(%q) has empty primary", w)
		}
		if s != "" && s == p {
			t.Errorf("Metaphone(%q) secondary duplicates primary: %q", w, p)
		}
	}
}

// TestNonLetterHandling pins the real contract: non-letters are DROPPED, not folded.
// Scattered punctuation is therefore transparent, but a digit standing in for a coding
// letter changes the key — echo keys sound, not spelling, so the caller folds
// leet/homoglyphs upstream first.
func TestNonLetterHandling(t *testing.T) {
	// Punctuation is dropped, so a punctuation-scattered label keys like the clean one.
	if !Sounds("p.h.o.n.e", "phone") {
		t.Errorf("Sounds(p.h.o.n.e, phone) = false; punctuation should be dropped (keys %v vs %v)",
			Keys("p.h.o.n.e"), Keys("phone"))
	}
	if Soundex("R-O-B-E-R-T") != Soundex("Robert") {
		t.Errorf("Soundex should drop punctuation: %q vs %q", Soundex("R-O-B-E-R-T"), Soundex("Robert"))
	}
	// A digit replacing a coding letter is dropped, NOT folded: g00gle cleans to "ggle"
	// (key "KL"), not "google", so it does not sound like the brand. Guards the doc claim
	// that echo does not do leet-folding — normalize leet upstream.
	if Sounds("g00gle", "google") {
		t.Errorf("Sounds(g00gle, google) = true; digits are dropped not folded, so keys should differ (%v vs %v)",
			Keys("g00gle"), Keys("google"))
	}
}
