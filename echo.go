package echo

// Metaphone returns the Double Metaphone codes of s: a primary code and an
// optional secondary code for an alternate pronunciation. The secondary is "" when
// s has a single pronunciation (the common case). Each code is at most four
// characters, drawn from the phonetic alphabet the algorithm emits (letters plus
// '0' for the "th" sound and 'X' for "sh"/"ch").
//
// s is the caller-normalised label; Metaphone folds it to uppercase and keys off
// the ASCII letters A–Z only, ignoring digits, punctuation, and non-ASCII runes
// (English/Latin, v1). An input with no ASCII letters yields ("", "").
func Metaphone(s string) (primary, secondary string) {
	return doubleMetaphone(clean(s))
}

// Soundex returns the standard (NARA/American) Soundex code of s: the first letter
// followed by three digits. Same-coded letters that are adjacent — or separated
// only by H or W — collapse to one digit; a vowel (A E I O U Y) between them does
// not, so the pair is coded twice. The code is right-padded with zeros and capped
// at four characters. An input with no ASCII letters yields "".
func Soundex(s string) string {
	w := clean(s)
	if len(w) == 0 {
		return ""
	}
	var out [4]byte
	out[0] = w[0]
	prev := soundexCode(w[0]) // seed with the first letter so an immediate repeat collapses
	nDigits := 1
	for i := 1; i < len(w) && nDigits < 4; i++ {
		switch code := soundexCode(w[i]); {
		case w[i] == 'H' || w[i] == 'W':
			// Transparent: leave prev intact so codes across H/W still collapse.
		case code == 0:
			// Vowel: reset prev so a following same code is coded again.
			prev = 0
		default:
			if code != prev {
				out[nDigits] = '0' + code
				nDigits++
			}
			prev = code
		}
	}
	for nDigits < 4 {
		out[nDigits] = '0'
		nDigits++
	}
	return string(out[:])
}

// Sounds reports whether a and b share a phonetic code — the homophone test.
// It compares the Double Metaphone keys any-to-any: a match on either string's
// primary or secondary code (e.g. "fone"/"phone" both key "FN", "kwik"/"quick"
// both key "KK") is a hit. Two labels with no ASCII letters never match.
//
// Sounds is deliberately over Double Metaphone alone, not Soundex: Soundex is a
// coarser bucket and folding it in would trade the precision this detector exists
// to provide for false positives. Callers who want Soundex bucketing can compare
// [Soundex] directly.
func Sounds(a, b string) bool {
	ka, kb := Keys(a), Keys(b)
	for _, x := range ka {
		for _, y := range kb {
			if x == y {
				return true
			}
		}
	}
	return false
}

// Keys returns the distinct phonetic codes of s — the canonical keys to index by.
// It is the Double Metaphone primary, plus the secondary when it differs, and is
// exactly what [Sounds] matches on. Use it to bucket a brand list by sound
// (map[string][]string keyed on each of Keys(brand)) so a candidate resolves in
// O(1): look up each of Keys(candidate). An input with no ASCII letters yields nil.
func Keys(s string) []string {
	primary, secondary := Metaphone(s)
	if primary == "" {
		return nil
	}
	if secondary == "" {
		return []string{primary}
	}
	return []string{primary, secondary}
}

// clean folds s to uppercase and keeps only the ASCII letters A–Z, dropping digits,
// punctuation, and non-ASCII bytes. Both coders operate on this form, so any input —
// arbitrary bytes, invalid UTF-8, empty — is handled without allocation surprises or
// panics; multibyte runes are simply discarded (English/Latin scope, v1). Dropping
// spaces also disables Double Metaphone's space-boundary rules (the SAN/VAN/VON
// prefixes and the JOSE and -IER endings), which is inconsequential for single-token
// registrable labels — they contain no spaces — and is why the codes match the
// reference on every single-token word.
func clean(s string) []byte {
	b := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			c -= 'a' - 'A'
		}
		if c >= 'A' && c <= 'Z' {
			b = append(b, c)
		}
	}
	return b
}

// soundexCode maps a letter to its Soundex digit, or 0 for the uncoded letters
// (vowels A E I O U Y, plus H and W which the caller treats as transparent).
func soundexCode(c byte) byte {
	switch c {
	case 'B', 'F', 'P', 'V':
		return 1
	case 'C', 'G', 'J', 'K', 'Q', 'S', 'X', 'Z':
		return 2
	case 'D', 'T':
		return 3
	case 'L':
		return 4
	case 'M', 'N':
		return 5
	case 'R':
		return 6
	}
	return 0
}
