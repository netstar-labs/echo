# echo — architecture

A flat library (`package echo`) at the repo root, one concern per file, plus a thin
CLI under `app/echo`. The pipeline is: label → clean to uppercase ASCII letters →
phonetic coder (Double Metaphone or Soundex) → key(s) → any-to-any match.

## Data flow

```
label ─▶ clean (uppercase, keep A–Z) ─▶ w []byte
                                          │
                    ┌─────────────────────┼─────────────────────┐
                    ▼                                            ▼
         doubleMetaphone(w)                                  Soundex(w)
         (primary, secondary)                          letter + 3 digits
                    │                                            │
              Keys(s) = {primary [, secondary]}                  │
                    ▼                                            ▼
   Sounds(a,b) = Keys(a) ∩ Keys(b) ≠ ∅               (coarser bucket, exported)
```

## The cleaner (`echo.go`)

Both coders start at `clean`: it folds the input to uppercase and keeps only the
ASCII letters `A–Z`, dropping digits, punctuation, and every non-ASCII byte in one
pass. Three consequences fall out of that one decision:

- **Punctuation transparency — but not leet folding.** `p.h.o.n.e` cleans to `PHONE`,
  so scattered punctuation keys the same as `phone`. A digit standing in for a letter
  is *dropped, not folded*, though: `g00gle` cleans to `ggle` (key `KL`), not `google`,
  and `Sounds("g00gle","google")` is false. Fold leet/homoglyph substitutions upstream
  (e.g. via `unmask`) before phonetic keying — echo keys sound, not spelling.
- **Total input safety.** Because the coders only ever see `A–Z`, arbitrary bytes,
  invalid UTF-8, and the empty string are handled with no special-casing and no
  panic; a multibyte rune is simply discarded (the English/Latin v1 scope).
- **Determinism and locale-independence.** The output depends only on the bytes, not
  on any Unicode table version or locale, so it is stable across releases — unlike
  `unmask`, whose skeleton tracks the Unicode confusables data.

## Double Metaphone (`metaphone.go`)

`doubleMetaphone` is a port of Lawrence Philips' reference algorithm (C/C++
User's Journal, June 2000). It walks the cleaned word with a cursor and a large
switch on the current letter; each case advances the cursor by the number of letters
its rule consumes and appends to **two** codes built in lockstep:

- `add(m)` appends `m` to both the primary and the secondary code.
- `add2(m, alt)` lets them diverge — the mechanism behind the secondary key, e.g.
  `CH` → `add2("X", "K")` so `Michael` keys `MKL`/`MXL`.

Position rules read the word through two safe accessors — `getAt(i)` returns `0`
past either end, and `stringAt(start, len, opts...)` is `false` for any
out-of-range window — so **no rule can index out of bounds**, at the start of the
word or the end, regardless of input. The `isSlavoGermanic` heuristic (the word
contains `W`, `K`, `CZ`, or `WITZ`) steers the Slavic/Germanic branches.

Both codes are capped at **four characters** (`maxCode`), the canonical length, and
the main loop stops once both reach it. At the end the secondary is cleared to `""`
whenever it equals the primary, so callers get a secondary only when the word
genuinely forks — the common case is a single key.

The rule set is the full canonical one: silent initial `GN/KN/PN/WR/PS` and initial
`X`→S; `PH`→F, `CK/CG/CQ`→K, `TH`→`0`/`T`; the `-MB` silent B; `DG`→J vs `TK`; the
Germanic `-ACH-`, the Greek `CH-`, the Italian soft/hard `G` and `CC`, the French
`-IER`/`-OIS`/`-EAUX` silent finals, and the Spanish `J`/`LL` cases.

### On the `0` and `X` symbols

Double Metaphone borrows two non-letter symbols: `0` (zero) stands for the *th*
sound (θ), and `X` stands for the *sh*/*ch* sound. They are ordinary characters in
the key — `Thumb` → `0M`/`TM`, `Schmidt` → `XMT`/`SMT` — and compare like any other.

## Soundex (`echo.go`)

`Soundex` implements the standard NARA/American algorithm: retain the first letter,
then map subsequent letters to digits (`B,F,P,V`→1; `C,G,J,K,Q,S,X,Z`→2; `D,T`→3;
`L`→4; `M,N`→5; `R`→6), collapse runs of the same digit, right-pad with zeros, and
cap at four characters. The two rules that separate a correct implementation from a
naive one:

- **H and W are transparent.** They carry no digit *and* do not reset the previous
  code, so two same-coded letters they bridge collapse to one digit. This is why
  `Ashcraft` is `A261`, not `A226`: the `s` and `c` (both 2) with an `h` between
  them count once.
- **Vowels reset; the first letter seeds.** A vowel (`A,E,I,O,U,Y`) carries no digit
  but *does* reset the previous code, so a same-coded letter after it is coded again
  — `Tymczak` → `T522`, not `T520`. And the previous-code state is seeded from the
  first (retained) letter, so an immediately following same-coded letter drops:
  `Pfister` → `P236` (the `f`, code 1, collapses into the retained `P`).

## Matching (`echo.go`)

`Keys(s)` is the primary code plus the secondary when it differs — the distinct
phonetic keys, and the canonical thing to index by. `Sounds(a, b)` compares
`Keys(a)` against `Keys(b)` any-to-any and returns true on the first shared key; it
is symmetric, deterministic, and returns false when either side has no phonetic
content (so an empty or all-digit candidate never spuriously flags a brand).

`Sounds` matches over the Double Metaphone keys **only**. Soundex is a deliberately
coarse bucket — many unrelated words share a Soundex code — so folding it into the
homophone test would raise recall at a steep precision cost. It is exported as its
own coder for callers who want that bucketing explicitly.

## Complexity

Let *n* be the label length in ASCII letters (small for host/brand labels).

- **`clean`** is O(*n*), one pass, one allocation (the byte slice).
- **`doubleMetaphone`** is O(*n*) — the cursor advances by at least one per switch
  iteration, each rule is O(1) — and produces a bounded (≤4-char) result.
- **`Soundex`** is O(*n*) with a fixed 4-byte output: one small heap allocation for
  that output string (the cleaned slice itself does not escape).
- **`Sounds`/`Keys`** are O(*n*) plus a constant-size (≤2×2) key comparison.

For the intended workload — short, caller-normalised labels — every call is
sub-microsecond; see `bench_test.go`.

## Deliberately out (YAGNI)

The scope is the smallest real capability; three tempting extensions are recorded as
future upgrades rather than pre-built:

- **No fuzzy phonetic distance.** v1 is exact key match (any-to-any). A phonetic
  *edit distance* (Levenshtein over the codes) is a later refinement — build it when
  a consumer needs graded phonetic similarity, not now.
- **No non-Latin phonetics / language detection.** The rule set is English/Latin
  (Double Metaphone). Language-specific coders (e.g. Kölner Phonetik, Beider-Morse)
  are a separate axis, deferred until a non-Latin corpus needs them.
- **No built-in Set type.** `Keys` is the indexing primitive; a consumer builds its
  own `map[string][]brand` bucket in three lines and owns its target projection and
  normalisation, exactly as `snare` and `unmask` consumers do.

## Consumer wiring (future work)

echo imports none of its consumers; they wire it in a few lines, each owning its own
target projection and query normalisation:

| Consumer | Signal | Wiring |
|---|---|---|
| Brand monitor | a `phonetic_suspect` term | normalise (eTLD+1), bucket brands by `Keys`, then match each of `Keys(candidate)`. Adding a scored term to a trained model triggers a retrain + model-version bump on the consumer side, not echo. |
| CT-log tailer (`vigil`) | phonetic filter on observed SANs | score each observed name against the brand bucket alongside the edit and glyph terms. |
| `twister` homophone generator | generate → detect round-trip | every homophone variant `twister` emits for a brand must `Sounds` back to that brand — the differential test the two share. |

## Layout

| File | Purpose |
|---|---|
| [echo.go](../echo.go) | `Metaphone`, `Soundex`, `Sounds`, `Keys`, the `clean` cleaner, and the Soundex digit map |
| [metaphone.go](../metaphone.go) | `doubleMetaphone` — the Philips port — and the `getAt`/`stringAt`/`isVowel` accessors |
| [doc.go](../doc.go) | package doc — the third-axis metaphor and the precision choice |
| [app/echo/](../app/echo/main.go) | the CLI — `phon` · `sounds` · `version` |
