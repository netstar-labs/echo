// Runnable example: index a brand set by phonetic key, then flag observed labels
// that *sound* like a brand even when they are spelled arbitrarily far from it.
//
//	go run ./example/sounds
package main

import (
	"fmt"

	"github.com/netstar-labs/echo"
)

func main() {
	brands := []string{"paypal", "phone", "quick", "google"}

	// Bucket each brand under every one of its phonetic keys — this is the O(1)
	// match index the roadmap's Set describes, built from echo.Keys.
	index := map[string]string{}
	for _, b := range brands {
		for _, k := range echo.Keys(b) {
			index[k] = b
		}
	}

	for _, obs := range []string{"fone", "kwik", "gooogle", "paypal-login", "example"} {
		p, s := echo.Metaphone(obs)
		hit := ""
		for _, k := range echo.Keys(obs) {
			if b, ok := index[k]; ok {
				hit = b
				break
			}
		}
		fmt.Printf("%-14q metaphone=(%q,%q) soundex=%q", obs, p, s, echo.Soundex(obs))
		if hit != "" {
			fmt.Printf("  -> sounds like %q", hit)
		}
		fmt.Println()
	}
}
