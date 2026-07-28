# echo

Phonetic squat detection for Go — **reduce a label to a phonetic code and match on
the sound, not the spelling**, so `fone`/`phone`, `kwik`/`quick`, `zoom`/`zoem` all
collide. Pure Go, **zero dependencies**, embedded rule tables. echo is the
**phonetic** third axis of the matcher family: `twist` measures edits, `unmask`
folds glyph look-alikes, echo folds sound — the one axis no edit- or glyph-metric
can reach, because a homophone can be spelled arbitrarily far from its target.

```go
echo.Metaphone("phone")            // ("FN", "")      — Double Metaphone: primary, optional secondary
echo.Metaphone("Schmidt")          // ("XMT", "SMT")  — a word with two pronunciations forks
echo.Soundex("Ashcraft")           // "A261"          — standard NARA Soundex
echo.Sounds("kwik", "quick")       // true            — both key "KK"
echo.Sounds("paypal", "microsoft") // false
echo.Keys("Schmidt")               // ["XMT" "SMT"]   — the keys to index a brand list by
```

## Where it fits

| Axis | Repo | Distance |
|---|---|---|
| edits | [`twist`](https://github.com/netstar-labs/twist) | Damerau-Levenshtein |
| glyphs | [`unmask`](https://github.com/netstar-labs/unmask) | UTS-39 skeleton |
| **sound** | **`echo`** | **Double Metaphone / Soundex** |

A brand monitor scores a candidate on all three and combines the terms; echo is one
term. Feed it the caller-normalised registrable label — echo keys off the ASCII
letters only, so scattered punctuation is transparent (`p.h.o.n.e` keys like
`phone`) — but fold leet/homoglyph substitutions upstream first, since a digit
replacing a letter is dropped, not read as that letter.

## Quick start

```go
s := echo.Sounds("fone", "phone")     // true

// Index a brand set by sound, then match candidates in O(1):
index := map[string]string{}
for _, b := range brands {
    for _, k := range echo.Keys(b) {  // one label can have two keys
        index[k] = b
    }
}
for _, k := range echo.Keys(candidate) {
    if brand, ok := index[k]; ok { /* candidate sounds like brand */ }
}
```

```sh
go run ./app/echo phon phone fone quick kwik    # metaphone + soundex per word
go run ./app/echo sounds fone phone             # exit 0 if they sound alike, 1 if not
printf 'night\nnite\n' | go run ./app/echo phon # words on stdin
```

## The precision choice

`Sounds` matches on the **Double Metaphone** keys any-to-any (primary or secondary,
either side). It deliberately does **not** fold `Soundex` into the match: Soundex is
a coarser four-slot bucket, and mixing it in would trade the precision this detector
exists to provide for false positives. `Soundex` is exported as its own coder for
callers who want that bucketing explicitly.

## Documentation

- **Start here** — [docs/introduction.md](docs/introduction.md) ·
  [docs/executive-summary.md](docs/executive-summary.md)
- **Deep dive** — [docs/architecture.md](docs/architecture.md)
- **Operations** — [docs/userguide.md](docs/userguide.md)
- **Examples** — [example/README.md](example/README.md)

## Layout

| File | Purpose |
|---|---|
| [echo.go](echo.go) | `Metaphone`, `Soundex`, `Sounds`, `Keys`, and the shared cleaner |
| [metaphone.go](metaphone.go) | the Double Metaphone engine — a port of Philips' reference (single-token labels) |
| [doc.go](doc.go) | package doc — the third-axis metaphor and the precision choice |
| [app/echo/](app/echo/main.go) | the CLI — `phon` · `sounds` · `version` |

## Scope (v1)

English/Latin phonetic rules (Double Metaphone), exact key match (any-to-any). Out
of scope: language detection, non-Latin phonetics, and a fuzzy phonetic *distance* —
all noted as later refinements, not the smallest capability. See
[docs/architecture.md](docs/architecture.md) § "Deliberately out".

Go module `github.com/netstar-labs/echo`. Standard library only. Build standalone
with `GOWORK=off`.
