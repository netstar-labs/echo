package echo

import "testing"

func BenchmarkMetaphone(b *testing.B) {
	for _, s := range []string{"phone", "Schmidt", "verylongbrandname", "Tymczak"} {
		b.Run(s, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = Metaphone(s)
			}
		})
	}
}

func BenchmarkSoundex(b *testing.B) {
	for _, s := range []string{"Robert", "Ashcraft", "verylongbrandname"} {
		b.Run(s, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = Soundex(s)
			}
		})
	}
}

func BenchmarkSounds(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Sounds("fone", "phone")
	}
}
