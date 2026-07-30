// Package echo detects squats that *sound* alike but are edit-distance far —
// fone/phone, kwik/quick, zoom/zoem — by reducing a label to a phonetic code and
// matching on the code rather than the spelling. It is the phonetic third axis of
// the netstar matcher family: snare measures edits, unmask folds glyph look-alikes
// (the UTS-39 skeleton), and echo folds sound. A brand monitor scores a candidate
// on all three and combines the terms; echo is the one no edit- or glyph-metric can
// reach, because a homophone can be spelled arbitrarily far from its target.
//
// Two coders, one question. [Metaphone] is a port of Lawrence Philips'
// Double Metaphone (the primary key plus an optional secondary for an alternate
// pronunciation), and [Soundex] is the standard NARA/American Soundex. [Sounds]
// answers the homophone question directly — do two labels share a phonetic code —
// and [Keys] returns the codes to index a brand list by, so a candidate resolves in
// O(1) against a pre-built bucket.
//
// # The matcher, and its precision
//
// [Sounds] compares the Double Metaphone keys any-to-any: a hit on either label's
// primary or secondary code. fone and phone both key "FN"; kwik and quick both key
// "KK". It deliberately does NOT fold Soundex into the match — Soundex is a coarser,
// four-slot bucket, and mixing it in would trade the precision this detector exists
// to provide for false positives. Soundex is exported as its own coder for callers
// who want that bucketing explicitly.
//
// # Scope (v1)
//
// echo is pure Go, standard library only, deterministic, and side-effect free — a
// pure function from label to phonetic key(s). It operates on the caller-normalised
// registrable label and keys off the ASCII letters A–Z only: digits, punctuation,
// and non-ASCII runes are dropped (not folded). Scattered punctuation is therefore
// transparent — p.h.o.n.e keys the same as phone — but a digit standing in for a
// letter is removed, not read as that letter: g00gle keys "KL", not "google", and
// Sounds("g00gle","google") is false. Fold leet and homoglyph substitutions upstream
// (e.g. via unmask) before phonetic keying — echo keys sound, not spelling. The rule
// set is English/Latin (Double Metaphone's rules); language
// detection and non-Latin phonetics are out of scope for v1, as is a fuzzy phonetic
// *distance* — echo does exact key matching (any-to-any), which is the smallest real
// capability. A phonetic edit distance is a later refinement, not this.
package echo
