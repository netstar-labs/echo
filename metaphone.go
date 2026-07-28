package echo

import "strings"

// maxCode caps each Double Metaphone code at four characters — the canonical
// length. The main loop also stops once both codes reach it.
const maxCode = 4

// doubleMetaphone computes the primary and secondary Double Metaphone codes of a
// cleaned (uppercase, A–Z only) word. It is a port of Lawrence Philips' reference
// algorithm (C/C++ User's Journal, June 2000): the same case analysis,
// the same primary/alternate branches, capped at four characters. The secondary
// code is returned only when it differs from the primary (callers get "" when a
// word has a single pronunciation).
//
// Both codes are built in lockstep: add(m) appends m to both, add2(m, alt) lets
// them diverge. Position rules read the cleaned word through getAt/stringAt, which
// are safe at any index, so no rule can index out of range.
func doubleMetaphone(w []byte) (primary, secondary string) {
	n := len(w)
	if n == 0 {
		return "", ""
	}
	last := n - 1
	slavoGermanic := isSlavoGermanic(w)

	var pri, sec strings.Builder
	add := func(m string) { pri.WriteString(m); sec.WriteString(m) }
	add2 := func(m, alt string) { pri.WriteString(m); sec.WriteString(alt) }
	done := func() bool { return pri.Len() >= maxCode && sec.Len() >= maxCode }

	cur := 0

	// Skip an initial silent letter: GN, KN, PN, WR, PS all lead with a mute.
	if stringAt(w, 0, 2, "GN", "KN", "PN", "WR", "PS") {
		cur = 1
	}
	// Initial 'X' is pronounced 'Z' (e.g. "Xavier") → code 'S'.
	if getAt(w, 0) == 'X' {
		add("S")
		cur = 1
	}

	for cur < n && !done() {
		switch c := w[cur]; c {

		case 'A', 'E', 'I', 'O', 'U', 'Y':
			// Vowels are coded only at the very start of the word.
			if cur == 0 {
				add("A")
			}
			cur++

		case 'B':
			// "-mb" (dumb, thumb) has its silent B swallowed by the M rule.
			add("P")
			if getAt(w, cur+1) == 'B' {
				cur += 2
			} else {
				cur++
			}

		case 'C':
			// Various Germanic spellings: "-ACH-" as 'K' (Bach), not 'X'.
			if cur > 1 && !isVowel(getAt(w, cur-2)) && stringAt(w, cur-1, 3, "ACH") &&
				getAt(w, cur+2) != 'I' && (getAt(w, cur+2) != 'E' || stringAt(w, cur-2, 6, "BACHER", "MACHER")) {
				add("K")
				cur += 2
				break
			}
			// Special case "caesar".
			if cur == 0 && stringAt(w, cur, 6, "CAESAR") {
				add("S")
				cur += 2
				break
			}
			// Italian "chianti".
			if stringAt(w, cur, 4, "CHIA") {
				add("K")
				cur += 2
				break
			}
			if stringAt(w, cur, 2, "CH") {
				// "michael"
				if cur > 0 && stringAt(w, cur, 4, "CHAE") {
					add2("K", "X")
					cur += 2
					break
				}
				// Greek roots, e.g. "chemistry", "chorus".
				if cur == 0 &&
					(stringAt(w, cur+1, 5, "HARAC", "HARIS") ||
						stringAt(w, cur+1, 3, "HOR", "HYM", "HIA", "HEM")) &&
					!stringAt(w, 0, 5, "CHORE") {
					add("K")
					cur += 2
					break
				}
				// Germanic, Greek, or otherwise 'K' for the 'ch' sound.
				if stringAt(w, 0, 4, "VAN ", "VON ") || stringAt(w, 0, 3, "SCH") ||
					stringAt(w, cur-2, 6, "ORCHES", "ARCHIT", "ORCHID") ||
					stringAt(w, cur+2, 1, "T", "S") ||
					((stringAt(w, cur-1, 1, "A", "O", "U", "E") || cur == 0) &&
						stringAt(w, cur+2, 1, "L", "R", "N", "M", "B", "H", "F", "V", "W", " ")) {
					add("K")
				} else if cur > 0 {
					if stringAt(w, 0, 2, "MC") {
						// e.g. "McHugh"
						add("K")
					} else {
						add2("X", "K")
					}
				} else {
					add("X")
				}
				cur += 2
				break
			}
			// e.g. "czerny"
			if stringAt(w, cur, 2, "CZ") && !stringAt(w, cur-2, 4, "WICZ") {
				add2("S", "X")
				cur += 2
				break
			}
			// e.g. "focaccia"
			if stringAt(w, cur+1, 3, "CIA") {
				add("X")
				cur += 3
				break
			}
			// double 'C', but not e.g. "McClellan"
			if stringAt(w, cur, 2, "CC") && !(cur == 1 && getAt(w, 0) == 'M') {
				// "bellocchio" but not "bacchus"
				if stringAt(w, cur+2, 1, "I", "E", "H") && !stringAt(w, cur+2, 2, "HU") {
					// "accident", "accede", "succeed"
					if (cur == 1 && getAt(w, cur-1) == 'A') ||
						stringAt(w, cur-1, 5, "UCCEE", "UCCES") {
						add("KS")
					} else {
						add("X")
					}
					cur += 3
					break
				}
				// Pierce's rule
				add("K")
				cur += 2
				break
			}
			if stringAt(w, cur, 2, "CK", "CG", "CQ") {
				add("K")
				cur += 2
				break
			}
			if stringAt(w, cur, 2, "CI", "CE", "CY") {
				// Italian vs. English
				if stringAt(w, cur, 3, "CIO", "CIE", "CIA") {
					add2("S", "X")
				} else {
					add("S")
				}
				cur += 2
				break
			}
			// else
			add("K")
			switch {
			case stringAt(w, cur+1, 2, " C", " Q", " G"): // "mac caffrey", "mac gregor"
				cur += 3
			case stringAt(w, cur+1, 1, "C", "K", "Q") && !stringAt(w, cur+1, 2, "CE", "CI"):
				cur += 2
			default:
				cur++
			}

		case 'D':
			if stringAt(w, cur, 2, "DG") {
				if stringAt(w, cur+2, 1, "I", "E", "Y") {
					// e.g. "edge"
					add("J")
					cur += 3
					break
				}
				// e.g. "edgar"
				add("TK")
				cur += 2
				break
			}
			if stringAt(w, cur, 2, "DT", "DD") {
				add("T")
				cur += 2
				break
			}
			add("T")
			cur++

		case 'F':
			add("F")
			if getAt(w, cur+1) == 'F' {
				cur += 2
			} else {
				cur++
			}

		case 'G':
			if getAt(w, cur+1) == 'H' {
				if cur > 0 && !isVowel(getAt(w, cur-1)) {
					add("K")
					cur += 2
					break
				}
				if cur < 3 && cur == 0 {
					// "ghislane", "ghiradelli"
					if getAt(w, cur+2) == 'I' {
						add("J")
					} else {
						add("K")
					}
					cur += 2
					break
				}
				// Parker's rule (with refinements), e.g. "hugh"
				if (cur > 1 && stringAt(w, cur-2, 1, "B", "H", "D")) ||
					(cur > 2 && stringAt(w, cur-3, 1, "B", "H", "D")) ||
					(cur > 3 && stringAt(w, cur-4, 1, "B", "H")) {
					cur += 2
					break
				}
				// e.g. "laugh", "McLaughlin", "cough", "rough"
				if cur > 2 && getAt(w, cur-1) == 'U' &&
					stringAt(w, cur-3, 1, "C", "G", "L", "R", "T") {
					add("F")
				} else if cur > 0 && getAt(w, cur-1) != 'I' {
					add("K")
				}
				cur += 2
				break
			}
			if getAt(w, cur+1) == 'N' {
				if cur == 1 && isVowel(getAt(w, 0)) && !slavoGermanic {
					add2("KN", "N")
				} else if !stringAt(w, cur+2, 2, "EY") && getAt(w, cur+1) != 'Y' && !slavoGermanic {
					add2("N", "KN")
				} else {
					add("KN")
				}
				cur += 2
				break
			}
			// "tagliaro"
			if stringAt(w, cur+1, 2, "LI") && !slavoGermanic {
				add2("KL", "L")
				cur += 2
				break
			}
			// -ges-, -gep-, -gel-, -gie- at the beginning
			if cur == 0 && (getAt(w, cur+1) == 'Y' ||
				stringAt(w, cur+1, 2, "ES", "EP", "EB", "EL", "EY", "IB", "IL", "IN", "IE", "EI", "ER")) {
				add2("K", "J")
				cur += 2
				break
			}
			// -ger-, -gy-
			if (stringAt(w, cur+1, 2, "ER") || getAt(w, cur+1) == 'Y') &&
				!stringAt(w, 0, 6, "DANGER", "RANGER", "MANGER") &&
				!stringAt(w, cur-1, 1, "E", "I") &&
				!stringAt(w, cur-1, 3, "RGY", "OGY") {
				add2("K", "J")
				cur += 2
				break
			}
			// Italian, e.g. "biaggi"
			if stringAt(w, cur+1, 1, "E", "I", "Y") || stringAt(w, cur-1, 4, "AGGI", "OGGI") {
				// obvious Germanic
				if stringAt(w, 0, 4, "VAN ", "VON ") || stringAt(w, 0, 3, "SCH") ||
					stringAt(w, cur+1, 2, "ET") {
					add("K")
				} else if stringAt(w, cur+1, 4, "IER ") {
					// French ending is always soft
					add("J")
				} else {
					add2("J", "K")
				}
				cur += 2
				break
			}
			add("K")
			if getAt(w, cur+1) == 'G' {
				cur += 2
			} else {
				cur++
			}

		case 'H':
			// Keep 'H' only at the start before a vowel, or between two vowels.
			if (cur == 0 || isVowel(getAt(w, cur-1))) && isVowel(getAt(w, cur+1)) {
				add("H")
				cur += 2
			} else {
				cur++
			}

		case 'J':
			// Spanish "jose", "san jacinto"
			if stringAt(w, cur, 4, "JOSE") || stringAt(w, 0, 4, "SAN ") {
				if (cur == 0 && getAt(w, cur+4) == ' ') || stringAt(w, 0, 4, "SAN ") {
					add("H")
				} else {
					add2("J", "H")
				}
				cur++
				break
			}
			if cur == 0 && !stringAt(w, cur, 4, "JOSE") {
				add2("J", "A") // "Yankelovich" / "Jankelowicz"
			} else if isVowel(getAt(w, cur-1)) && !slavoGermanic &&
				(getAt(w, cur+1) == 'A' || getAt(w, cur+1) == 'O') {
				// Spanish pronunciation, e.g. "bajador"
				add2("J", "H")
			} else if cur == last {
				add2("J", "")
			} else if !stringAt(w, cur+1, 1, "L", "T", "K", "S", "N", "M", "B", "Z") &&
				!stringAt(w, cur-1, 1, "S", "K", "L") {
				add("J")
			}
			if getAt(w, cur+1) == 'J' {
				cur += 2
			} else {
				cur++
			}

		case 'K':
			add("K")
			if getAt(w, cur+1) == 'K' {
				cur += 2
			} else {
				cur++
			}

		case 'L':
			if getAt(w, cur+1) == 'L' {
				// Spanish, e.g. "cabrillo", "gallegos"
				if (cur == n-3 && stringAt(w, cur-1, 4, "ILLO", "ILLA", "ALLE")) ||
					((stringAt(w, last-1, 2, "AS", "OS") || stringAt(w, last, 1, "A", "O")) &&
						stringAt(w, cur-1, 4, "ALLE")) {
					add2("L", "")
					cur += 2
					break
				}
				cur += 2
			} else {
				cur++
			}
			add("L")

		case 'M':
			// "-umb" silent B (thumb, dumb), and doubled M.
			if (stringAt(w, cur-1, 3, "UMB") && (cur+1 == last || stringAt(w, cur+2, 2, "ER"))) ||
				getAt(w, cur+1) == 'M' {
				cur += 2
			} else {
				cur++
			}
			add("M")

		case 'N':
			add("N")
			if getAt(w, cur+1) == 'N' {
				cur += 2
			} else {
				cur++
			}

		case 'P':
			if getAt(w, cur+1) == 'H' {
				add("F")
				cur += 2
				break
			}
			// "campbell", "raspberry"
			if stringAt(w, cur+1, 1, "P", "B") {
				cur += 2
			} else {
				cur++
			}
			add("P")

		case 'Q':
			add("K")
			if getAt(w, cur+1) == 'Q' {
				cur += 2
			} else {
				cur++
			}

		case 'R':
			// French final "-ier", e.g. "rogier" (silent R), but not "hochmeier".
			if cur == last && !slavoGermanic && stringAt(w, cur-2, 2, "IE") &&
				!stringAt(w, cur-4, 2, "ME", "MA") {
				add2("", "R")
			} else {
				add("R")
			}
			if getAt(w, cur+1) == 'R' {
				cur += 2
			} else {
				cur++
			}

		case 'S':
			// "island", "isle", "carlisle"
			if stringAt(w, cur-1, 3, "ISL", "YSL") {
				cur++
				break
			}
			// "sugar-"
			if cur == 0 && stringAt(w, cur, 5, "SUGAR") {
				add2("X", "S")
				cur++
				break
			}
			if stringAt(w, cur, 2, "SH") {
				if stringAt(w, cur+1, 4, "HEIM", "HOEK", "HOLM", "HOLZ") {
					add("S")
				} else {
					add("X")
				}
				cur += 2
				break
			}
			// Italian & Armenian, e.g. "sio", "sian"
			if stringAt(w, cur, 3, "SIO", "SIA") || stringAt(w, cur, 4, "SIAN") {
				if !slavoGermanic {
					add2("S", "X")
				} else {
					add("S")
				}
				cur += 3
				break
			}
			// German & anglicisations: -sm, -sn, -sl, -sw at start; -sz-
			if (cur == 0 && stringAt(w, cur+1, 1, "M", "N", "L", "W")) || stringAt(w, cur+1, 1, "Z") {
				add2("S", "X")
				if stringAt(w, cur+1, 1, "Z") {
					cur += 2
				} else {
					cur++
				}
				break
			}
			if stringAt(w, cur, 2, "SC") {
				// Schlesinger's rule
				if getAt(w, cur+2) == 'H' {
					// Dutch origin, e.g. "school", "schooner"
					if stringAt(w, cur+3, 2, "OO", "ER", "EN", "UY", "ED", "EM") {
						// "schermerhorn", "schenker"
						if stringAt(w, cur+3, 2, "ER", "EN") {
							add2("X", "SK")
						} else {
							add("SK")
						}
						cur += 3
						break
					}
					if cur == 0 && !isVowel(getAt(w, 3)) && getAt(w, 3) != 'W' {
						add2("X", "S")
					} else {
						add("X")
					}
					cur += 3
					break
				}
				if stringAt(w, cur+2, 1, "I", "E", "Y") {
					add("S")
					cur += 3
					break
				}
				add("SK")
				cur += 3
				break
			}
			// French final "-ais", "-ois" (silent S), e.g. "artois".
			if cur == last && stringAt(w, cur-2, 2, "AI", "OI") {
				add2("", "S")
			} else {
				add("S")
			}
			if stringAt(w, cur+1, 1, "S", "Z") {
				cur += 2
			} else {
				cur++
			}

		case 'T':
			if stringAt(w, cur, 4, "TION") {
				add("X")
				cur += 3
				break
			}
			if stringAt(w, cur, 3, "TIA", "TCH") {
				add("X")
				cur += 3
				break
			}
			if stringAt(w, cur, 2, "TH") || stringAt(w, cur, 3, "TTH") {
				// "thomas", "thames", or Germanic → hard 'T'
				if stringAt(w, cur+2, 2, "OM", "AM") ||
					stringAt(w, 0, 4, "VAN ", "VON ") || stringAt(w, 0, 3, "SCH") {
					add("T")
				} else {
					add2("0", "T")
				}
				cur += 2
				break
			}
			if stringAt(w, cur+1, 1, "T", "D") {
				cur += 2
			} else {
				cur++
			}
			add("T")

		case 'V':
			add("F")
			if getAt(w, cur+1) == 'V' {
				cur += 2
			} else {
				cur++
			}

		case 'W':
			if stringAt(w, cur, 2, "WR") {
				add("R")
				cur += 2
				break
			}
			if cur == 0 && (isVowel(getAt(w, cur+1)) || stringAt(w, cur, 2, "WH")) {
				// "Wasserman" should match "Vasserman"; "Uomo" should match "Womo".
				if isVowel(getAt(w, cur+1)) {
					add2("A", "F")
				} else {
					add("A")
				}
				cur++
				break
			}
			// "Arnow" should match "Arnoff"
			if (cur == last && isVowel(getAt(w, cur-1))) ||
				stringAt(w, cur-1, 5, "EWSKI", "EWSKY", "OWSKI", "OWSKY") ||
				stringAt(w, 0, 3, "SCH") {
				add2("", "F")
				cur++
				break
			}
			// Polish, e.g. "filipowicz"
			if stringAt(w, cur, 4, "WICZ", "WITZ") {
				add2("TS", "FX")
				cur += 4
				break
			}
			cur++ // otherwise silent

		case 'X':
			// French "breaux" ends silent.
			if !(cur == last && (stringAt(w, cur-3, 3, "IAU", "EAU") || stringAt(w, cur-2, 2, "AU", "OU"))) {
				add("KS")
			}
			if stringAt(w, cur+1, 1, "C", "X") {
				cur += 2
			} else {
				cur++
			}

		case 'Z':
			// Chinese pinyin, e.g. "zhao"
			if getAt(w, cur+1) == 'H' {
				add("J")
				cur += 2
				break
			}
			if stringAt(w, cur+1, 2, "ZO", "ZI", "ZA") ||
				(slavoGermanic && cur > 0 && getAt(w, cur-1) != 'T') {
				add2("S", "TS")
			} else {
				add("S")
			}
			if getAt(w, cur+1) == 'Z' {
				cur += 2
			} else {
				cur++
			}

		default:
			cur++
		}
	}

	primary = truncate(pri.String())
	secondary = truncate(sec.String())
	if secondary == primary {
		secondary = ""
	}
	return primary, secondary
}

// truncate caps a code at the canonical four characters.
func truncate(s string) string {
	if len(s) > maxCode {
		return s[:maxCode]
	}
	return s
}

// isSlavoGermanic reports the Slavic/Germanic heuristic that steers several rules:
// the word contains W, K, "CZ", or "WITZ".
func isSlavoGermanic(w []byte) bool {
	s := string(w)
	return strings.IndexByte(s, 'W') >= 0 || strings.IndexByte(s, 'K') >= 0 ||
		strings.Contains(s, "CZ") || strings.Contains(s, "WITZ")
}

// getAt returns the byte at i, or 0 when i is out of range, so position rules can
// probe freely past either end of the word.
func getAt(w []byte, i int) byte {
	if i < 0 || i >= len(w) {
		return 0
	}
	return w[i]
}

// stringAt reports whether the length-run of w starting at start equals any opt.
// An out-of-range window (including a negative start) is never a match, which is
// exactly the "nothing there" semantics the rules assume.
func stringAt(w []byte, start, length int, opts ...string) bool {
	if start < 0 || start+length > len(w) {
		return false
	}
	seg := w[start : start+length]
	for _, o := range opts {
		if string(seg) == o {
			return true
		}
	}
	return false
}

// isVowel reports whether b is one of A E I O U Y (Double Metaphone treats Y as a
// vowel). getAt's zero sentinel is not a vowel, so probes past the ends are safe.
func isVowel(b byte) bool {
	switch b {
	case 'A', 'E', 'I', 'O', 'U', 'Y':
		return true
	}
	return false
}
