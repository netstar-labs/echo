# Meet echo — the sound, not the spelling

A phishing kit does not have to *look* like a brand to fool a reader — it can
*sound* like one. `fone.com` for `phone`, `kwik-loans` for `quick`, `zoem` for
`zoom`, `sekure` for `secure`. Read aloud they are the target; on the page they are
a string an edit-distance metric scores as far away (three, four, five edits) and a
glyph-skeleton metric never touches at all, because every character is a perfectly
ordinary Latin letter. The homophone is a squat class with its own door, and echo is
the key.

## What it actually is

echo is a pure-Go, zero-dependency phonetic-key library. It reduces a label to a
**phonetic code** — a short string that captures how the label *sounds* — so two
labels that sound alike map to the same code, and asking "do these sound alike?"
becomes a string comparison. It ships two coders. `Metaphone` is a port of
Lawrence Philips' **Double Metaphone**: it returns a primary code and, when a word
has a plausible second pronunciation, an optional secondary — so `Schmidt` keys as
both `XMT` and `SMT`, covering the anglicised and Germanic sounds at once. `Soundex`
is the classic NARA/American Soundex, a first letter plus three digits. `Sounds(a,
b)` answers the homophone question directly, and `Keys(s)` returns the codes to
index a brand list by.

## Why the phonetic axis is its own thing

The matcher family has three axes, and each one catches what the others cannot.
`snare` measures **edits** — insert, delete, substitute, transpose — and catches
`paypa1`, `gooogle`, `micros0ft`. `unmask` folds **glyphs** to a Unicode skeleton
and catches `pаypаl` with a Cyrillic а. Neither can catch `fone`: it is three edits
from `phone` (well past a typo budget) and every glyph is a legitimate Latin letter
(nothing for the skeleton to fold). Only the sound collides. That is the whole
argument for echo: a homophone can be spelled *arbitrarily* far from its target
while sounding identical, so no spelling-based distance will ever reach it. A brand
monitor runs all three axes and combines the terms; echo is the term that closes the
phonetic door.

## The scope it keeps

echo answers one question — *do these two labels share a phonetic code* — and
refuses the neighbouring ones. It is a pure function: label in, key(s) out, no
resolution, no scoring, no language model, no state. It keys off the ASCII letters
only: scattered punctuation is transparent, but a digit replacing a letter is dropped
(fold leet/homoglyphs upstream first), and any input — Unicode, invalid UTF-8, empty —
is handled without a panic. The rules are English/Latin (Double Metaphone's rule set),
documented as such; it does exact key matching, not a fuzzy phonetic *distance*.
That discipline is what keeps it a three-line drop-in for a detector, an identity
service, or a CT-log filter, and unit-testable against published reference vectors
in isolation.

*Read next:* [executive-summary.md](executive-summary.md) ·
[architecture.md](architecture.md) · [userguide.md](userguide.md) ·
[../example/README.md](../example/README.md)
