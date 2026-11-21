package foreignvalue

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"
)

// benchDocument is the document of the Iris benchmark
// compiler-backend/go/benchmarks/idioms Bench.Idiom.Json, as JSON text.
func benchDocument(size int, indent string) string {
	items := make([]any, size)
	for i := range items {
		tags := make([]any, i%4+1)
		for j := range tags {
			tags[j] = "t" + strconv.Itoa((i+j)%13)
		}
		item := map[string]any{
			"id": i, "name": "item-" + strconv.Itoa(i) + "-" + "abcdefghijklmnopqrstuvwxyz"[:i%17],
			"price": float64(i%1000) / 8, "tags": tags,
			"dims": map[string]any{"w": float64(i%50) + 0.5, "h": float64(i%30) * 2.25},
		}
		if i%7 == 0 {
			item["note"] = "line\nbreak\t\"q\" café " + strconv.Itoa(i)
		}
		items[i] = item
	}
	text, _ := json.MarshalIndent(map[string]any{"owner": "bench \"owner\" naïve", "version": 3, "items": items}, "", indent)
	return string(text)
}

var compactDocument, prettyDocument = benchDocument(10000, ""), benchDocument(10000, "  ")

func BenchmarkParseJSON(b *testing.B) {
	b.SetBytes(int64(len(compactDocument)))
	for b.Loop() {
		ParseJSON(compactDocument)
	}
}

func BenchmarkParseJSONPretty(b *testing.B) {
	b.SetBytes(int64(len(prettyDocument)))
	for b.Loop() {
		ParseJSON(prettyDocument)
	}
}

func BenchmarkUnmarshalAny(b *testing.B) {
	b.SetBytes(int64(len(compactDocument)))
	for b.Loop() {
		var tree any
		json.Unmarshal([]byte(compactDocument), &tree)
	}
}

func BenchmarkStringify(b *testing.B) {
	value := ParseJSON(compactDocument)
	b.SetBytes(int64(len(compactDocument)))
	for b.Loop() {
		UnsafeStringify(value)
	}
}

func BenchmarkStringifyLongStrings(b *testing.B) {
	value := ParseJSON(`["` + strings.Repeat("a long run of plain text without anything to escape, ", 40) + `"]`)
	for b.Loop() {
		UnsafeStringify(value)
	}
}

func BenchmarkMarshalAny(b *testing.B) {
	var tree any
	json.Unmarshal([]byte(compactDocument), &tree)
	b.SetBytes(int64(len(compactDocument)))
	for b.Loop() {
		json.Marshal(tree)
	}
}
