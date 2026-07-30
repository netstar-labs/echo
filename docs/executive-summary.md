# echo — executive summary

**What it is.** A small, dependency-free Go library that detects squats which
*sound* like a brand but are spelled far from it. `Metaphone` returns Double
Metaphone codes (primary plus an optional secondary), `Soundex` returns standard
NARA Soundex, `Sounds(a, b)` reports whether two labels share a phonetic code, and
`Keys(s)` returns the codes to bucket a brand list by. It is the phonetic sibling of
`snare` (edit distance) and `unmask` (glyph skeleton) — the same "reduce to a key,
then match" shape, over sound instead of spelling.

**Why it exists.** A homophone squat — `fone`/`phone`, `kwik`/`quick`,
`sekure`/`secure` — is invisible to the other two axes. It is several edits from its
target, so a typo budget rejects it, and it is written in ordinary Latin letters, so
a glyph skeleton has nothing to fold. The only principled route to the homophone
class is to compare *sound*, and that is the one thing echo does.

**What you get.**
- **Double Metaphone**, a port of Philips' reference algorithm — the full
  canonical rule set (silent initial letters, `PH`→F, `CK`→K, `TH`, the
  Slavic/Germanic/Romance special cases), with a primary and an optional secondary
  code so alternate pronunciations are covered.
- **Standard Soundex** — first letter plus three digits, with correct handling of
  the H/W bridge (`Ashcraft`→`A261`) and vowel-separated repeats (`Tymczak`→`T522`).
- **`Sounds` — any-to-any key match** over the Double Metaphone codes: a hit on
  either label's primary or secondary. Symmetric, deterministic.
- **`Keys` — the canonical index key**: bucket a brand list by `Keys(brand)` and a
  candidate resolves in O(1) via `Keys(candidate)`.
- **No dependencies, no state, no configuration** — standard library only, a pure
  function of its input, safe for concurrent use, and unaffected by locale.

**Where it fits.** echo is the *algorithm* layer, coupled to nothing. It takes a
caller-normalised label and returns phonetic key(s); it keys off ASCII letters only,
so scattered punctuation is transparent, but leet/homoglyph substitutions must be
folded upstream first (a digit replacing a letter is dropped, not read as it).
Consumers wire it in a few lines: a brand monitor adds a
`phonetic_suspect` term to its score, a CT-log tailer filters observed names against
a brand bucket, `twister`'s homophone generator round-trips each variant back to its
brand.

**A precision choice.** `Sounds` matches on the Double Metaphone keys only, not
Soundex. Soundex is a coarser four-slot bucket; folding it into the match would
raise recall at a steep cost in false positives, against the grain of a detector
whose job is precision. `Soundex` remains exported for callers who want that
bucketing on purpose.

**What it is not.** Not an edit-distance metric (that is `snare`), not a glyph
skeleton (that is `unmask`), not a resolver or classifier, and not — in v1 — a fuzzy
phonetic *distance* or a non-Latin coder. It answers one question — do these share a
phonetic code — and only that.
