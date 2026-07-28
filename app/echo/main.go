// Command echo prints phonetic codes and tests whether two labels sound alike.
//
//	echo phon <word...>     # Double Metaphone (primary/secondary) + Soundex per word
//	echo sounds <a> <b>     # report whether a and b share a phonetic code
//	echo version
//
// With no word arguments, phon reads one word per line from stdin. sounds exits 0
// when the pair matches and 1 when it does not, so it composes in a shell test.
package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/netstar-labs/echo"
)

// stamped by the build via -ldflags -X.
var (
	version = "dev"
	build   = "none"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	var err error
	switch os.Args[1] {
	case "phon":
		err = phon(os.Args[2:])
	case "sounds":
		err = sounds(os.Args[2:])
	case "version", "-version", "--version", "-v":
		fmt.Printf("echo %s (%s)\n", version, build)
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "echo:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: echo <phon|sounds|version> [args...]")
	fmt.Fprintln(os.Stderr, "  echo phon <word...>   metaphone + soundex per word (stdin if no args)")
	fmt.Fprintln(os.Stderr, "  echo sounds <a> <b>   do a and b sound alike? (exit 0 if yes, 1 if no)")
	os.Exit(2)
}

// phon prints "word <tab> primary <tab> secondary <tab> soundex" for each word,
// taken from the arguments or, when there are none, one per line from stdin.
func phon(args []string) error {
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	emit := func(word string) {
		if word == "" {
			return
		}
		p, s := echo.Metaphone(word)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", word, p, s, echo.Soundex(word))
	}

	if len(args) > 0 {
		for _, word := range args {
			emit(word)
		}
		return nil
	}
	sc := bufio.NewScanner(os.Stdin)
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)
	for sc.Scan() {
		emit(strings.TrimSpace(sc.Text()))
	}
	return sc.Err()
}

// sounds reports whether two labels share a phonetic code, printing the verdict and
// exiting 1 on a non-match so it works as a shell predicate.
func sounds(args []string) error {
	if len(args) != 2 {
		return errors.New("sounds: need exactly two words")
	}
	a, b := args[0], args[1]
	if echo.Sounds(a, b) {
		fmt.Printf("%s\t%s\tsounds-like\n", a, b)
		return nil
	}
	fmt.Printf("%s\t%s\tno\n", a, b)
	os.Exit(1)
	return nil
}
