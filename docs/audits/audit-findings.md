# echo — A1 adversarial audit + least-code pass, pre-public

Full-repo audit ahead of public release, 2026-09-30: four dimensions (A simpler · B
dedup · C correctness+optimization · D doc-drift) plus a separate least-code pass.
Every candidate reproduced against the code before repair, then independently
re-verified by a second, adversarial pass whose only job was to try to refute it, then
re-verified a third time after the fix to confirm it actually holds with no regression.

## Verdict summary

| # | Finding | Dimension | Severity | Verdict | Status |
|---|---|---|---|---|---|
| 1 | `Metaphone("jose")` (and case variants) returned the wrong primary/secondary code — a missing disjunct from the reference algorithm's JOSE special case | C | **High** | CONFIRMED | **Fixed** |
| 2 | `done` closure had exactly one call site | A | Low | — | **Fixed** |
| 3 | `isSlavoGermanic` converted `[]byte` to `string` only to use the wrong stdlib package | A | Low | — | **Fixed** |
| 4 | Version/build stamp vars have no wiring anywhere in the repo | A | Low | — | Noted, not fixed |
| 5 | `clean()`'s hand-rolled loop vs. `strings.Map` | A | Very low | — | Not changed (already minimal) |

## Finding 1 — the JOSE bug (CONFIRMED, fixed)

**Root cause.** Double Metaphone's reference algorithm (`handleJ` in Apache Commons
Codec, verified against the live source, not from memory) special-cases the Spanish
name "JOSE": the whole-word case collapses to a single code (`H`), while "JOSE"
appearing as *part of* a longer name forks into `J`/`H`. The reference detects the
whole-word case via `index == 0 && charAt(value, index+4) == ' '` **or**
`value.length() == 4` **or** a `"SAN "` prefix. This port's `clean()` strips every
space in the input before matching begins — so the space-boundary check can never
fire for this library — and the port never added the `length == 4` fallback the
reference relies on for exactly that situation. The result: the whole-word "JOSE" case
was unreachable, and "jose" silently took the multi-word fork instead.

**Reproduced, before fix.**
```
Metaphone("jose") = primary="JS" secondary="HS"    // wrong: should be a single "HS" code
```
**Reproduced, after fix** (`|| len(w) == 4` added to the inner condition):
```
Metaphone("jose") = primary="HS" secondary=""      // correct
```
Confirmed the scope is exact by construction, not just by sampling: the added
disjunct only changes behavior when the cleaned input is exactly 4 bytes, so every
other word (`josephine`, `jones`, `sanjose`, `carlosjose`, `josefina`, `josie`, and
others tried across two independent verification passes) is provably byte-identical
to the pre-fix code path.

**Blast radius, checked explicitly.** `Sounds("jose", "hose")` returns `true` both
before and after the fix — before, by coincidence (the wrong secondary code still
matched); after, correctly (the right primary code matches). So the fix's practical
urgency rests on `Metaphone()`/`Keys()` correctness for any caller that inspects
primary and secondary individually, not on this one brand pair.

**Fix.** `metaphone.go`, the J-case JOSE/SAN branch: added `|| len(w) == 4` to the
inner condition, matching the reference algorithm's `value.length() == 4` disjunct.

## Findings 2–3 (least-code, fixed)

- **`done` closure** (`pri.Len() >= maxCode && sec.Len() >= maxCode`) had exactly one
  call site. Inlined as `for cur < n && (pri.Len() < maxCode || sec.Len() < maxCode)`
  at the loop — verified the negation is exact (De Morgan's law, checked against a
  36-cell truth table across `priLen`/`secLen` combinations, not just cited) and that a
  long input still terminates correctly at the 4-character cap.
- **`isSlavoGermanic`** converted its `[]byte` argument to a `string` solely to call
  `strings.IndexByte`/`strings.Contains` — an allocating conversion to reach the wrong
  package. Rewritten to operate directly on the byte slice via
  `bytes.IndexByte`/`bytes.Contains`. Verified byte-for-byte equivalent behavior across
  8 words spanning every branch (W-only, K-only, "CZ", "WITZ", and no-match cases).

## Finding 4 — the version stamp that's never set (noted, not fixed)

`app/echo/main.go`'s `version`/`build` vars default to `"dev"`/`"none"` with a comment
claiming they're "stamped by the build via `-ldflags -X`" — but no build script,
Makefile, or CI step in this repo ever passes `-ldflags`, so the CLI always prints
`echo dev (none)`. Left as-is rather than fixed: whether this repo should grow its own
`build/echo` packaging script (matching `snare`/`twister`'s convention) or is meant to
be built by an external packager is a scope/ownership decision, not a bug fix — flagged
here so it isn't silently shipped as a mystery to the next public reader.

## Finding 5 — `clean()` (considered, not changed)

A hand-rolled uppercase-and-filter loop could be a `strings.Map` call instead, but the
current form is already 13 lines, operates byte-wise (matching the doc's explicit claim
that non-ASCII bytes are dropped without rune decoding), and the stdlib version
wouldn't read more clearly. Not worth the churn.

## Re-validation gate

| Check | Result |
|---|---|
| `go build` / `go vet` / `gofmt -l` | clean |
| `go test -race -count=1 ./...` | all packages pass |
| `FuzzMetaphone`, `FuzzSoundex` (15–20s each, both before and after the fix) | ~4.4–8.3M execs, 0 crashers, 0 new failures |
| Independent skeptic pass on the JOSE claim, pre-fix, against the live reference algorithm source | CONFIRMED — fetched Apache Commons Codec's actual `handleJ`/`cleanInput`, not recalled from memory |
| Independent skeptic pass on the fix + both cleanups, post-fix, with its own fresh inputs | HOLDS, no regressions found |
