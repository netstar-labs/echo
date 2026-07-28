# echo examples

| Example | What it shows | Run |
|---|---|---|
| [sounds](sounds/main.go) | indexing a brand set by phonetic key (`echo.Keys`) and flagging observed labels that *sound* like a brand — `fone`→`phone`, `kwik`→`quick`, `gooogle`→`google` | `go run ./example/sounds` |

The phonetic key is the sound, not the spelling: a homophone can be spelled
arbitrarily far from its target, so a brand list is bucketed by `Keys(brand)` and a
candidate resolves in O(1) by looking up each of `Keys(candidate)`. See
[../doc.go](../doc.go).
