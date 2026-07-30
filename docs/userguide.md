# echo — user guide

## Library

```go
import "github.com/netstar-labs/echo"

echo.Metaphone("phone")   // ("FN", "")      primary, optional secondary
echo.Metaphone("Schmidt") // ("XMT", "SMT")  a two-pronunciation word forks
echo.Metaphone("p.h.o.n.e") // ("FN", "")      punctuation dropped → keyed as "phone"
echo.Soundex("Robert")    // "R163"
echo.Soundex("Ashcraft")  // "A261"
echo.Sounds("fone", "phone")        // true   both key "FN"
echo.Sounds("kwik", "quick")        // true   both key "KK"
echo.Sounds("paypal", "microsoft")  // false
echo.Keys("Schmidt")                // ["XMT" "SMT"]
```

### The four calls

- **`Metaphone(s) (primary, secondary string)`** — Double Metaphone. The secondary
  is `""` unless `s` has a plausible second pronunciation. Each code is at most four
  characters over the phonetic alphabet: letters, plus `0` for the *th* sound and `X`
  for the *sh*/*ch* sound.
- **`Soundex(s) string`** — standard NARA Soundex: one letter and three digits, or
  `""` for an input with no ASCII letters.
- **`Sounds(a, b) bool`** — do `a` and `b` share a Double Metaphone key? Symmetric,
  and false when either side has no phonetic content.
- **`Keys(s) []string`** — the distinct phonetic keys of `s` (primary, plus the
  secondary when different), or `nil` for no phonetic content. This is what `Sounds`
  matches on and what you index a brand list by.

### Contract and cost

- **echo normalises nothing.** The caller supplies the already-normalised
  registrable label — lowercase, eTLD+1, IDNA-decode, whatever the domain rules
  require. echo keys off the ASCII letters `A–Z` only: digits, punctuation, and
  non-ASCII runes are *dropped* (not folded): scattered punctuation is transparent
  (`p-h-o-n-e` keys like `phone`), but a digit standing in for a coding letter changes
  the key — `g00gle` keys `KL`, not `google`, so fold leet/homoglyphs upstream first
  (`ph0ne` keys like `phone` only because the `0` replaced a *vowel*). Any input
  (Unicode, invalid UTF-8, empty) is handled without a panic.
- **English/Latin, v1.** The rule set is Double Metaphone's; non-Latin phonetics and
  language detection are out of scope.
- **Deterministic and stateless.** No configuration, no locale dependence, no shared
  state — every function is safe for concurrent calls.
- **`Sounds` is Double-Metaphone-only** by design (precision over recall); use
  `Soundex` directly if you want the coarser bucket.

### Indexing a brand set

`Keys` is the O(1) index primitive: bucket each brand under every one of its keys,
then look up each of a candidate's keys.

```go
index := map[string]string{}
for _, b := range brands {
    for _, k := range echo.Keys(b) {
        index[k] = b
    }
}

func soundsLikeBrand(candidate string) (string, bool) {
    for _, k := range echo.Keys(candidate) {
        if b, ok := index[k]; ok {
            return b, true
        }
    }
    return "", false
}
```

## CLI

Build: `go build -o echo ./app/echo` (standalone: `GOWORK=off`).

```sh
# metaphone (primary/secondary) + soundex per word
echo phon phone fone quick kwik
#   phone   FN          P500
#   fone    FN          F500
#   quick   KK          Q200
#   kwik    KK          K200

# words one per line on stdin when no arguments are given
printf 'night\nnite\n' | echo phon
#   night   NT          N230
#   nite    NT          N300

# do two labels sound alike?  exit 0 if yes, 1 if no — usable as a shell predicate
echo sounds fone phone   ;  # -> "fone  phone  sounds-like",  exit 0
echo sounds paypal microsoft  ;  # -> "paypal  microsoft  no",  exit 1

echo version
```

Output of `phon` is one tab-separated line per word: `word <tab> primary <tab>
secondary <tab> soundex` (the secondary column is blank when the word has one
pronunciation).

| Command | Args | Meaning |
|---|---|---|
| `phon` | `<word...>` | print `word<tab>primary<tab>secondary<tab>soundex`; words from args or stdin |
| `sounds` | `<a> <b>` | print whether `a` and `b` sound alike; exit 0 on a match, 1 on a non-match |
| `version` | — | binary version + build revision |

## Reference vectors

echo is pinned against published reference outputs (`echo_test.go`): the Soundex
golden set (`Robert`→`R163`, `Ashcraft`→`A261`, `Tymczak`→`T522`, `Pfister`→`P236`,
`Honeyman`→`H555`, …) and canonical Double Metaphone codes (`Smith`→`SM0`/`XMT`,
`Schmidt`→`XMT`/`SMT`, `Michael`→`MKL`/`MXL`, …), plus the homophone properties that
are the library's purpose. `FuzzMetaphone` and `FuzzSoundex` assert no input ever
panics and every output stays well-formed.

## When to reach past echo

echo does exact phonetic-key matching. If a consumer needs *graded* phonetic
similarity (how close, not just same-or-not), the upgrade is a phonetic edit
distance over the codes — deferred until then. Non-Latin phonetics need a
language-specific coder (Kölner Phonetik, Beider-Morse), a separate axis. Edit
distance and glyph confusables are out of scope by design — that is `snare` and
`unmask`; see [architecture.md](architecture.md) § "Deliberately out (YAGNI)".
