package tauri

import "github.com/i-am-the-slime/iris-go/promise"

// Browser APIs that a native host does not provide. The embedded JavaScript
// host has no DOM, OffscreenCanvas or FontFaceSet either, so the capability
// queries answer false exactly as they do there, and markgraf measures text
// through Tauri.Native.measureText instead. Every other operation needs a
// canvas or font set that does not exist.

// SupportsOffscreenCanvas implements Graphics.Canvas.Extra.supportsOffscreenCanvas.
func SupportsOffscreenCanvas() bool { return false }

// SupportsCanvasElement implements Graphics.Canvas.Extra.supportsCanvasElement.
func SupportsCanvasElement() bool { return false }

// HasFontSet implements Web.Font.Loading.hasFontSet.
func HasFontSet() bool { return false }

// SupportsFontFace implements Web.Font.Loading.supportsFontFace.
func SupportsFontFace() bool { return false }

// LoadFontImpl implements Web.Font.Loading.loadFontImpl: without a font set
// there is nothing to load, so the promise resolves at once.
func LoadFontImpl(_ string) any { return promise.Resolve(struct{}{}) }

// FontsReadyImpl implements Web.Font.Loading.fontsReadyImpl.
func FontsReadyImpl() any { return promise.Resolve(struct{}{}) }

// CheckFont implements Web.Font.Loading.checkFont, which needs a font set.
func CheckFont(_ string) bool { Unsupported("the browser font set"); return false }

// MeasureTextInkImpl implements Graphics.Canvas.Extra.measureTextInkImpl
// (record fields ascent, descent, width), which needs a 2D context.
func MeasureTextInkImpl(_ any, _ string) (float64, float64, float64) {
	Unsupported("canvas rendering")
	return 0, 0, 0
}

// GetContext2DNullable implements Graphics.Canvas.Extra.getContext2DNullable.
func GetContext2DNullable(_ any) *any {
	Unsupported("canvas rendering")
	return nil
}
