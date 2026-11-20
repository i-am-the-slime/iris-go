package tauri

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"testing"
)

// TestMeasureMatchesNativeHost compares measurements with those of the Rust
// host for a corpus of "font<TAB>text" lines (TAURI_MEASURE_CORPUS) and the
// host's "width ascent descent" answers (TAURI_MEASURE_EXPECTED), measured
// with the fonts listed in TAURI_MEASURE_FONTS ("family:weight:italic:path,…").
func TestMeasureMatchesNativeHost(t *testing.T) {
	corpus, expected, fonts := os.Getenv("TAURI_MEASURE_CORPUS"), os.Getenv("TAURI_MEASURE_EXPECTED"), os.Getenv("TAURI_MEASURE_FONTS")
	if corpus == "" || expected == "" || fonts == "" {
		t.Skip("no native-host oracle configured")
	}
	var assets []FontAsset
	for _, entry := range strings.Split(fonts, ",") {
		fields := strings.SplitN(entry, ":", 4)
		weight, _ := strconv.Atoi(fields[1])
		data, err := os.ReadFile(fields[3])
		if err != nil {
			t.Fatal(err)
		}
		assets = append(assets, FontAsset{Family: fields[0], Weight: uint16(weight), Italic: fields[2] == "1", Data: data})
	}
	Configure(Config{Fonts: assets})
	graphics, err := sharedGraphics()
	if err != nil {
		t.Fatal(err)
	}
	inputs, _ := os.ReadFile(corpus)
	answers, _ := os.ReadFile(expected)
	lines := strings.Split(strings.TrimSuffix(string(inputs), "\n"), "\n")
	results := bufio.NewScanner(strings.NewReader(string(answers)))
	failures := 0
	for _, line := range lines {
		results.Scan()
		want := results.Text()
		font, text, _ := strings.Cut(line, "\t")
		text = strings.ReplaceAll(text, `\n`, "\n")
		metrics, err := graphics.measure(font, text)
		got := ""
		if err != nil {
			got = "error " + err.Error()
		} else {
			got = rustFloat(metrics.width) + " " + rustFloat(metrics.ascent) + " " + rustFloat(metrics.descent)
		}
		if got != want {
			failures++
			if failures <= 40 {
				t.Errorf("%q %q: got %s, want %s", font, text, got, want)
			}
		}
	}
	t.Logf("%d of %d measurements differ", failures, len(lines))
}

func rustFloat(value float64) string {
	text := strconv.FormatFloat(value, 'f', -1, 64)
	if !strings.ContainsAny(text, ".eEn") {
		text += ".0"
	}
	return text
}
