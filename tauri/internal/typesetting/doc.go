// Package typesetting is a vendored subset of github.com/go-text/typesetting
// v0.3.5 (Unlicense OR BSD-3-Clause, see LICENSE): the OpenType parser, the
// HarfBuzz port, scripts and grapheme segmentation that Tauri.Native's text
// measurement needs. It is vendored so go-support keeps no module
// dependencies and so it can carry these fixes:
//
//   - harfbuzz/set_digest.go: addRange skipped every filter once any one was
//     saturated, so pair-kerning subtables were wrongly skipped for glyphs
//     whose coverage range was added after that point (HarfBuzz saturates
//     each filter separately).
//   - font/glyphs.go: TIFF strike decoding is dropped with golang.org/x/image.
package typesetting
