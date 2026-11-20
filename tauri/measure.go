package tauri

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/i-am-the-slime/iris-go/tauri/internal/typesetting/font"
	ot "github.com/i-am-the-slime/iris-go/tauri/internal/typesetting/font/opentype"
	"github.com/i-am-the-slime/iris-go/tauri/internal/typesetting/harfbuzz"
	"github.com/i-am-the-slime/iris-go/tauri/internal/typesetting/language"
	"github.com/i-am-the-slime/iris-go/tauri/internal/typesetting/segmenter"
)

// MeasureText implements Tauri.Native.measureText: the advance width and the
// ink ascent/descent of a CSS font shorthand, exactly as the native host's
// Graphics::measure computes them (fontdb matching, per-grapheme fallback,
// HarfBuzz shaping, and tight glyph outline bounds). The results are the
// Metrics record's fields in label order.
func MeasureText(shorthand string, text string) (ascent float64, descent float64, width float64) {
	graphics, err := sharedGraphics()
	if err != nil {
		throw(err.Error())
	}
	metrics, err := graphics.measure(shorthand, text)
	if err != nil {
		throw(err.Error())
	}
	return metrics.ascent, metrics.descent, metrics.width
}

type textMetrics struct {
	width, ascent, descent float64
}

type requestedFont struct {
	size      float64
	weight    uint16
	style     faceStyle
	smallCaps bool
}

type graphics struct {
	fonts  *fontDatabase
	lock   sync.Mutex
	shaped map[*faceInfo]*shapingFace
}

type shapingFace struct {
	face     *font.Face
	font     *harfbuzz.Font
	features []harfbuzz.Feature
	upem     float64
	bounds   map[font.GID]glyphBounds
}

type glyphBounds struct {
	yMin, yMax int16
	present    bool
}

var (
	graphicsOnce  sync.Once
	graphicsValue *graphics
	graphicsError error
)

// sharedGraphics loads the font database on first use, like Host::graphics.
func sharedGraphics() (*graphics, error) {
	graphicsOnce.Do(func() {
		database, err := newFontDatabase(config.Fonts)
		if err != nil {
			graphicsError = err
			return
		}
		graphicsValue = &graphics{fonts: database, shaped: map[*faceInfo]*shapingFace{}}
	})
	return graphicsValue, graphicsError
}

func (graphics *graphics) measure(shorthand string, text string) (textMetrics, error) {
	graphics.lock.Lock()
	defer graphics.lock.Unlock()
	numericWeight, hasWeight, keywords := splitNumericWeight(shorthand)
	parsed, err := parseFontShorthand(keywords)
	if err != nil {
		return textMetrics{}, hostError(fmt.Sprintf("invalid CSS font %s: %s", rustDebugString(shorthand), err))
	}
	size, err := fontSize(parsed.size)
	if err != nil {
		return textMetrics{}, err
	}
	families, err := parseFontFamilies(parsed.family)
	if err != nil {
		return textMetrics{}, err
	}
	weight := numericWeight
	if !hasWeight {
		switch parsed.weight {
		case "", "normal":
			weight = 400
		case "bold", "bolder":
			weight = 700
		case "lighter":
			weight = 300
		default:
			value, err := strconv.ParseUint(parsed.weight, 10, 16)
			if err != nil {
				return textMetrics{}, hostError(err.Error())
			}
			weight = uint16(value)
		}
	}
	style := styleNormal
	switch parsed.style {
	case "italic":
		style = styleItalic
	case "oblique":
		style = styleOblique
	}
	stretch := uint16(5)
	if value, ok := stretchNumbers[parsed.stretch]; ok {
		stretch = value
	}
	base := graphics.fonts.query(fontQuery{families: families, weight: weight, stretch: stretch, style: style})
	if base == nil {
		sansSerif, _ := genericFamily("sans-serif")
		base = graphics.fonts.query(fontQuery{families: []string{sansSerif}, weight: weight, stretch: stretch, style: style})
	}
	if base == nil {
		base = graphics.fonts.faces[0]
	}
	requested := requestedFont{size: size, weight: weight, style: style, smallCaps: parsed.variant == "small-caps"}
	return graphics.shapeText(base, text, requested)
}

// shapeText splits the text into runs of one fallback face and one script,
// keeping grapheme clusters whole, and shapes each run.
func (graphics *graphics) shapeText(base *faceInfo, text string, requested requestedFont) (textMetrics, error) {
	var metrics textMetrics
	runes := []rune(text)
	var segments segmenter.Segmenter
	segments.Init(runes)
	runStart := 0
	runFace := base
	runScript := language.Common
	graphemes := segments.GraphemeIterator()
	for graphemes.Next() {
		grapheme := graphemes.Grapheme()
		selected := graphics.fallback(base, grapheme.Text)
		script := runScript
		for _, r := range grapheme.Text {
			if found := language.LookupScript(r); found != language.Common && found != language.Inherited {
				script = found
				break
			}
		}
		if grapheme.Offset > runStart &&
			(selected != runFace || (script != runScript && runScript != language.Common)) {
			if err := graphics.shape(runFace, runes[runStart:grapheme.Offset], requested, &metrics); err != nil {
				return metrics, err
			}
			runStart = grapheme.Offset
		}
		runFace = selected
		runScript = script
	}
	if runStart < len(runes) {
		if err := graphics.shape(runFace, runes[runStart:], requested, &metrics); err != nil {
			return metrics, err
		}
	}
	return metrics, nil
}

func (graphics *graphics) supports(face *faceInfo, text []rune) bool {
	data := face.source.bytes()
	tables, ok := parseRawFace(data, face.index)
	if !ok || !tables.faceParses() {
		return false
	}
	cmap := tables["cmap"]
	for _, r := range text {
		// Joiners and variation selectors modify their surrounding glyphs.
		if r == 0x200C || r == 0x200D || (r >= 0xFE00 && r <= 0xFE0F) || (r >= 0xE0100 && r <= 0xE01EF) {
			continue
		}
		if cmap == nil {
			return false
		}
		if _, found := cmapLookup(cmap, r); !found {
			return false
		}
	}
	return true
}

func (graphics *graphics) fallback(base *faceInfo, text []rune) *faceInfo {
	if graphics.supports(base, text) {
		return base
	}
	for _, face := range graphics.fonts.faces {
		similar := face.style == base.style || face.weight == base.weight || face.stretch == base.stretch
		if similar && graphics.supports(face, text) {
			return face
		}
	}
	for _, face := range graphics.fonts.faces {
		if graphics.supports(face, text) {
			return face
		}
	}
	return base
}

// shape adds one run's advance and ink extents, in the run's font units
// scaled to the requested size.
func (graphics *graphics) shape(face *faceInfo, text []rune, requested requestedFont, metrics *textMetrics) error {
	shaping, err := graphics.shapingFace(face)
	if err != nil {
		return err
	}
	scale := requested.size / shaping.upem
	buffer := harfbuzz.NewBuffer()
	buffer.AddRunes(text, 0, -1)
	guessSegmentProperties(buffer)
	features := shaping.features
	if requested.smallCaps {
		smallCaps, _ := harfbuzz.ParseFeature("smcp")
		features = append(append([]harfbuzz.Feature{}, features...), smallCaps)
	}
	buffer.Shape(shaping.font, features)
	advance := int64(0)
	for index, info := range buffer.Info {
		position := buffer.Pos[index]
		advance += int64(position.XAdvance)
		bounds := shaping.boundsOf(font.GID(uint16(info.Glyph)))
		if bounds.present {
			metrics.ascent = math.Max(metrics.ascent, float64(int32(bounds.yMax)+position.YOffset)*scale)
			metrics.descent = math.Max(metrics.descent, -float64(int32(bounds.yMin)+position.YOffset)*scale)
		}
	}
	// The explicit conversion rounds the product, keeping arm64 from fusing
	// the multiply-add as Rust never does.
	metrics.width += float64(math.Abs(float64(advance)) * scale)
	return nil
}

// guessSegmentProperties is rustybuzz's: the first strong script and its
// direction, and no language (HarfBuzz would take the locale's).
func guessSegmentProperties(buffer *harfbuzz.Buffer) {
	buffer.Props.Language = "und"
	buffer.GuessSegmentProperties()
	buffer.Props.Language = ""
}

func (graphics *graphics) shapingFace(face *faceInfo) (*shapingFace, error) {
	if shaping, ok := graphics.shaped[face]; ok {
		return shaping, nil
	}
	data := face.source.bytes()
	var parsed *font.Face
	var err error
	if _, collection := facesInCollection(data); collection {
		var faces []*font.Face
		faces, err = font.ParseTTC(bytes.NewReader(data))
		if err == nil {
			if int(face.index) >= len(faces) {
				err = fmt.Errorf("face index out of range")
			} else {
				parsed = faces[face.index]
			}
		}
	} else {
		parsed, err = font.ParseTTF(bytes.NewReader(data))
	}
	if err != nil {
		return nil, hostError("font could not be shaped")
	}
	var features []harfbuzz.Feature
	for _, part := range strings.Split(face.features, ",") {
		if strings.TrimSpace(part) == "" {
			continue
		}
		feature, err := harfbuzz.ParseFeature(strings.TrimSpace(part))
		if err != nil {
			return nil, hostError("invalid OpenType feature: " + part)
		}
		features = append(features, feature)
	}
	shaping := &shapingFace{
		face:     parsed,
		font:     harfbuzz.NewFont(parsed),
		features: features,
		upem:     float64(parsed.Upem()),
		bounds:   map[font.GID]glyphBounds{},
	}
	graphics.shaped[face] = shaping
	return shaping, nil
}

// boundsOf is ttf-parser's glyph_bounding_box: the extremes of every outline
// point, control points included, truncated toward zero.
func (shaping *shapingFace) boundsOf(glyph font.GID) glyphBounds {
	if bounds, ok := shaping.bounds[glyph]; ok {
		return bounds
	}
	var bounds glyphBounds
	if outline, ok := shaping.face.GlyphDataOutline(glyph); ok && len(outline.Segments) != 0 {
		yMin, yMax := float32(math.MaxFloat32), float32(-math.MaxFloat32)
		for _, segment := range outline.Segments {
			count := 1
			switch segment.Op {
			case ot.SegmentOpQuadTo:
				count = 2
			case ot.SegmentOpCubeTo:
				count = 3
			}
			for _, point := range segment.Args[:count] {
				yMin = float32(math.Min(float64(yMin), float64(point.Y)))
				yMax = float32(math.Max(float64(yMax), float64(point.Y)))
			}
		}
		low, okLow := truncateInt16(yMin)
		high, okHigh := truncateInt16(yMax)
		bounds = glyphBounds{yMin: low, yMax: high, present: okLow && okHigh}
	}
	shaping.bounds[glyph] = bounds
	return bounds
}

func truncateInt16(value float32) (int16, bool) {
	truncated := math.Trunc(float64(value))
	if truncated < math.MinInt16 || truncated > math.MaxInt16 {
		return 0, false
	}
	return int16(truncated), true
}

var stretchNumbers = map[string]uint16{
	"ultra-condensed": 1, "extra-condensed": 2, "condensed": 3, "semi-condensed": 4,
	"semi-expanded": 6, "expanded": 7, "extra-expanded": 8, "ultra-expanded": 9,
}

// splitNumericWeight lifts a numeric CSS weight (any value 1–1000) out of the
// shorthand before the keyword parser sees it, as the native host does.
func splitNumericWeight(shorthand string) (uint16, bool, string) {
	tokens := strings.Fields(shorthand)
	sizeAt := len(tokens)
	for index, token := range tokens {
		if (token[0] >= '0' && token[0] <= '9' || token[0] == '.') && !isRustNumber(token) {
			sizeAt = index
			break
		}
	}
	for index := 0; index < sizeAt; index++ {
		if isRustNumber(tokens[index]) {
			value, _ := strconv.ParseFloat(tokens[index], 64)
			value = math.Min(math.Max(math.Round(value), 1), 1000)
			rest := append(append([]string{}, tokens[:index]...), tokens[index+1:]...)
			return uint16(value), true, strings.Join(rest, " ")
		}
	}
	return 0, false, shorthand
}

// isRustNumber is `str::parse::<f64>().is_ok()`.
func isRustNumber(token string) bool {
	lower := strings.ToLower(strings.TrimLeft(token, "+-"))
	switch lower {
	case "inf", "infinity", "nan":
		return true
	}
	if strings.ContainsAny(token, "_xXpP") || strings.HasPrefix(lower, "0x") {
		return false
	}
	_, err := strconv.ParseFloat(token, 64)
	return err == nil || strings.Contains(err.Error(), "value out of range")
}

type fontShorthand struct {
	style, variant, weight, stretch, size, family string
}

// parseFontShorthand is svgtypes' FontShorthand::from_str.
func parseFontShorthand(text string) (fontShorthand, error) {
	var parsed fontShorthand
	position := skipSpaces(text, 0)
	for round := 0; round < 4; round++ {
		end := position
		for end < len(text) && isASCIIIdent(text[end]) {
			end++
		}
		ident := text[position:end]
		switch ident {
		case "normal":
		case "small-caps":
			parsed.variant = ident
		case "italic", "oblique":
			parsed.style = ident
		case "bold", "bolder", "lighter", "100", "200", "300", "400", "500", "600", "700", "800", "900":
			parsed.weight = ident
		case "ultra-condensed", "extra-condensed", "condensed", "semi-condensed",
			"semi-expanded", "expanded", "extra-expanded", "ultra-expanded":
			parsed.stretch = ident
		default:
			round = 4
			continue
		}
		position = skipSpaces(text, end)
	}
	start := position
	if position >= len(text) {
		return parsed, fmt.Errorf("unexpected end of stream")
	}
	if text[position] >= '0' && text[position] <= '9' {
		end, ok := lengthEnd(text, position)
		if !ok {
			return parsed, fmt.Errorf("invalid number at position %d", position+1)
		}
		position = end
	} else {
		end := position
		for end < len(text) && isASCIIIdent(text[end]) {
			end++
		}
		switch text[position:end] {
		case "xx-small", "x-small", "small", "medium", "large", "x-large", "xx-large", "larger", "smaller":
			position = end
		default:
			return parsed, fmt.Errorf("unexpected data at position %d", start+1)
		}
	}
	parsed.size = text[start:position]
	position = skipSpaces(text, position)
	if position >= len(text) {
		return parsed, fmt.Errorf("unexpected end of stream")
	}
	if text[position] == '/' {
		position = skipSpaces(text, position+1)
		end, ok := lengthEnd(text, position)
		if !ok {
			return parsed, fmt.Errorf("invalid number at position %d", position+1)
		}
		position = skipSpaces(text, end)
	}
	if position >= len(text) {
		return parsed, fmt.Errorf("unexpected end of stream")
	}
	parsed.family = text[position:]
	return parsed, nil
}

// lengthEnd skips an SVG number and its unit.
func lengthEnd(text string, position int) (int, bool) {
	end := numberEnd(text, position)
	if end == position {
		return position, false
	}
	for _, unit := range []string{"%", "em", "ex", "px", "in", "cm", "mm", "pt", "pc"} {
		if strings.HasPrefix(text[end:], unit) {
			return end + len(unit), true
		}
	}
	return end, true
}

func numberEnd(text string, position int) int {
	end := position
	if end < len(text) && (text[end] == '+' || text[end] == '-') {
		end++
	}
	digits := end
	for end < len(text) && text[end] >= '0' && text[end] <= '9' {
		end++
	}
	if end < len(text) && text[end] == '.' {
		end++
		for end < len(text) && text[end] >= '0' && text[end] <= '9' {
			end++
		}
	}
	if end == digits {
		return position
	}
	if end < len(text) && (text[end] == 'e' || text[end] == 'E') &&
		!(end+1 < len(text) && (text[end+1] == 'm' || text[end+1] == 'x')) {
		exponent := end + 1
		if exponent < len(text) && (text[exponent] == '+' || text[exponent] == '-') {
			exponent++
		}
		start := exponent
		for exponent < len(text) && text[exponent] >= '0' && text[exponent] <= '9' {
			exponent++
		}
		if exponent > start {
			end = exponent
		}
	}
	return end
}

func skipSpaces(text string, position int) int {
	for position < len(text) && (text[position] == ' ' || text[position] == '\t' || text[position] == '\n' || text[position] == '\r' || text[position] == '\f') {
		position++
	}
	return position
}

func isASCIIIdent(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_'
}

// fontSize is the native host's font_size: absolute CSS lengths in px.
func fontSize(value string) (float64, error) {
	end := numberEnd(value, 0)
	number, err := strconv.ParseFloat(value[:end], 64)
	if err != nil || end == 0 {
		return 0, hostError("invalid number")
	}
	unit := value[end:]
	scale := 0.0
	switch unit {
	case "", "px":
		scale = 1
	case "pt":
		scale = 96.0 / 72.0
	case "pc":
		scale = 16
	case "in":
		scale = 96
	case "cm":
		scale = 96.0 / 2.54
	case "mm":
		scale = 96.0 / 25.4
	default:
		return 0, hostError("native font measurements require an absolute CSS font size")
	}
	size := number * scale
	if size > 0 && !math.IsInf(float64(float32(size)), 0) && float32(size) > 0 {
		return size, nil
	}
	return 0, hostError("font size must be positive and finite in native coordinates")
}

// parseFontFamilies is svgtypes' parse_font_families, resolving generic
// families to fontdb's defaults.
func parseFontFamilies(text string) ([]string, error) {
	var families []string
	position := 0
	for position < len(text) {
		position = skipSpaces(text, position)
		if position >= len(text) {
			return nil, hostError("unexpected end of stream")
		}
		var family string
		if quote := text[position]; quote == '\'' || quote == '"' {
			end := strings.IndexByte(text[position+1:], quote)
			if end < 0 {
				return nil, hostError("unexpected end of stream")
			}
			family = text[position+1 : position+1+end]
			position += end + 2
		} else {
			var idents []string
			for position < len(text) && text[position] != ',' {
				end := position
				for end < len(text) {
					r, width := utf8.DecodeRuneInString(text[end:])
					if r < 0x80 && !isASCIIIdent(byte(r)) {
						break
					}
					end += width
				}
				if end == position {
					return nil, hostError(fmt.Sprintf("invalid ident at position %d", position+1))
				}
				idents = append(idents, text[position:end])
				position = skipSpaces(text, end)
			}
			family = strings.Join(idents, " ")
			if generic, ok := genericFamily(family); ok {
				family = generic
			}
		}
		if family != "" {
			families = append(families, family)
		}
		if position < len(text) {
			if text[position] != ',' {
				break
			}
			position++
		}
	}
	return families, nil
}

// rustDebugString renders a string the way Rust's {:?} does for plain text.
func rustDebugString(value string) string {
	return strconv.Quote(value)
}
