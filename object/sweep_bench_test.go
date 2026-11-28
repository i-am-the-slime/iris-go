package object

import "testing"

// sweepWords are words of one to eight letters from a sixteen-letter
// alphabet, as the benchmarks' sweep draws them: mostly distinct.
func sweepWords(count int) []string {
	state := uint64(23)
	next := func() uint64 {
		state ^= state << 13
		state ^= state >> 7
		state ^= state << 17
		return state
	}
	words := make([]string, count)
	for index := range words {
		letters := make([]byte, 1+next()%8)
		for position := range letters {
			letters[position] = byte('a' + next()%16)
		}
		words[index] = string(letters)
	}
	return words
}

func BenchmarkSweepInsertOwned(b *testing.B) {
	words := sweepWords(20000)
	for b.Loop() {
		m := Lend(Empty[int]())
		for _, word := range words {
			m = InsertOwned(word, 1, m)
		}
	}
}

func BenchmarkSweepInsertGoMap(b *testing.B) {
	words := sweepWords(20000)
	for b.Loop() {
		m := map[string]int{}
		for _, word := range words {
			m[word] = 1
		}
	}
}

func BenchmarkSweepPoke(b *testing.B) {
	words := sweepWords(10000)
	for b.Loop() {
		m := New[int]()
		for _, word := range words {
			Poke(word, 1, m)
		}
	}
}

func BenchmarkSweepPokePresized(b *testing.B) {
	words := sweepWords(10000)
	for b.Loop() {
		m := &struct {
			Index map[string]int
			Ints  int
			Keys  []string
			Vals  []int
		}{
			Index: make(map[string]int, len(words)),
			Keys:  make([]string, 0, len(words)),
			Vals:  make([]int, 0, len(words)),
		}
		for _, word := range words {
			Poke(word, 1, m)
		}
	}
}

func BenchmarkSweepPresizedGoMap(b *testing.B) {
	words := sweepWords(10000)
	for b.Loop() {
		m := make(map[string]int, len(words))
		for _, word := range words {
			m[word] = 1
		}
	}
}

func BenchmarkSweepFromFoldableWith(b *testing.B) {
	words := sweepWords(10000)
	for b.Loop() {
		m := New[int]()
		for _, word := range words {
			slot, found := Slot(m, word)
			if found {
				m.Vals[slot] += 1
			} else {
				Poke(word, 1, m)
			}
		}
	}
}

func BenchmarkSweepFoldlKeysUnsafeIndex(b *testing.B) {
	m := New[int]()
	for _, word := range sweepWords(10000) {
		Poke(word, 1, m)
	}
	for b.Loop() {
		total := 0
		for _, key := range Keys(m) {
			total += UnsafeIndex(m, key)
		}
		_ = total
	}
}

func BenchmarkSweepFoldlGoMap(b *testing.B) {
	m := map[string]int{}
	for _, word := range sweepWords(10000) {
		m[word] = 1
	}
	for b.Loop() {
		total := 0
		for _, value := range m {
			total += value
		}
		_ = total
	}
}

func BenchmarkSweepPokeGoMap(b *testing.B) {
	words := sweepWords(10000)
	for b.Loop() {
		m := map[string]int{}
		for _, word := range words {
			m[word] = 1
		}
	}
}

func BenchmarkSweepFoldlEntries(b *testing.B) {
	m := New[int]()
	for _, word := range sweepWords(10000) {
		Poke(word, 1, m)
	}
	for b.Loop() {
		total := 0
		for _, value := range Entries(m) {
			total += value
		}
		_ = total
	}
}

func BenchmarkSweepInsertIntoShared(b *testing.B) {
	m := New[int]()
	for _, word := range sweepWords(10000) {
		Poke(word, 1, m)
	}
	for b.Loop() {
		_ = Insert("a new key", 2, m)
	}
}

func BenchmarkSweepFromFoldableWithPokeAbsent(b *testing.B) {
	words := sweepWords(10000)
	for b.Loop() {
		m := New[int]()
		for _, word := range words {
			slot, found := Slot(m, word)
			if found {
				m.Vals[slot] += 1
			} else {
				PokeAbsent(word, 1, m)
			}
		}
	}
}
