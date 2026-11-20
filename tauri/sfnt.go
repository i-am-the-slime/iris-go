package tauri

import (
	"encoding/binary"
	"unicode/utf16"
)

// The minimal OpenType reading the font database needs: face headers, names,
// style metadata and Unicode coverage. Each reader mirrors the ttf-parser
// 0.25 behaviour that fontdb and rustybuzz rely on in the native host, so a
// face is accepted, named and matched exactly as it is there.

const (
	tagCollection = 0x74746366 // 'ttcf'
	tagTrueType   = 0x00010000
	tagAppleTrue  = 0x74727565 // 'true'
	tagOpenType   = 0x4F54544F // 'OTTO'
)

type sfntTables map[string][]byte

func be16(data []byte, offset int) (uint16, bool) {
	if offset < 0 || offset+2 > len(data) {
		return 0, false
	}
	return binary.BigEndian.Uint16(data[offset:]), true
}

func be32(data []byte, offset int) (uint32, bool) {
	if offset < 0 || offset+4 > len(data) {
		return 0, false
	}
	return binary.BigEndian.Uint32(data[offset:]), true
}

// facesInCollection is ttf_parser::fonts_in_collection: nil for a single face.
func facesInCollection(data []byte) (uint32, bool) {
	magic, ok := be32(data, 0)
	if !ok || magic != tagCollection {
		return 0, false
	}
	count, ok := be32(data, 8)
	return count, ok
}

// parseRawFace is ttf_parser::RawFace::parse: the table directory of one face.
func parseRawFace(data []byte, index uint32) (sfntTables, bool) {
	offset := 0
	magic, ok := be32(data, 0)
	if !ok {
		return nil, false
	}
	if magic == tagCollection {
		count, ok := be32(data, 8)
		if !ok || index >= count {
			return nil, false
		}
		face, ok := be32(data, 12+4*int(index))
		if !ok {
			return nil, false
		}
		offset = int(face)
		if magic, ok = be32(data, offset); !ok {
			return nil, false
		}
	} else if index != 0 {
		return nil, false
	}
	if magic != tagTrueType && magic != tagAppleTrue && magic != tagOpenType {
		return nil, false
	}
	count, ok := be16(data, offset+4)
	if !ok || offset+12+16*int(count) > len(data) {
		return nil, false
	}
	tables := sfntTables{}
	for record := 0; record < int(count); record++ {
		at := offset + 12 + 16*record
		tag := string(data[at : at+4])
		start, _ := be32(data, at+8)
		length, _ := be32(data, at+12)
		end := uint64(start) + uint64(length)
		if end > uint64(len(data)) {
			continue
		}
		tables[tag] = data[start:end]
	}
	return tables, true
}

// faceParses is the part of ttf_parser::Face::parse that can reject a face:
// the required head, hhea and maxp tables.
func (tables sfntTables) faceParses() bool {
	head := tables["head"]
	if len(head) < 54 {
		return false
	}
	upem, _ := be16(head, 18)
	format, _ := be16(head, 50)
	if upem < 16 || upem > 16384 || format > 1 {
		return false
	}
	if len(tables["hhea"]) < 36 {
		return false
	}
	maxp := tables["maxp"]
	version, ok := be32(maxp, 0)
	glyphs, _ := be16(maxp, 4)
	return ok && (version == 0x00005000 || version == 0x00010000) && glyphs != 0
}

type nameRecord struct {
	platform, encoding, language, id uint16
	value                            []byte
}

// nameRecords iterates the name table like ttf-parser, which stops at the
// first record whose string lies outside the storage area.
func nameRecords(data []byte) []nameRecord {
	version, ok := be16(data, 0)
	if !ok {
		return nil
	}
	count, _ := be16(data, 2)
	storage, _ := be16(data, 4)
	position := 6
	if version == 1 {
		tags, ok := be16(data, 6)
		if !ok {
			return nil
		}
		position = 8 + 4*int(tags)
	} else if version != 0 {
		return nil
	}
	if position+12*int(count) > len(data) {
		return nil
	}
	recordsEnd := position + 12*int(count)
	storageStart := recordsEnd
	if storageStart < int(storage) {
		storageStart = int(storage)
	}
	if storageStart > len(data) {
		return nil
	}
	stored := data[storageStart:]
	var records []nameRecord
	for index := 0; index < int(count); index++ {
		at := position + 12*index
		platform, _ := be16(data, at)
		encoding, _ := be16(data, at+2)
		language, _ := be16(data, at+4)
		id, _ := be16(data, at+6)
		length, _ := be16(data, at+8)
		offset, _ := be16(data, at+10)
		end := int(offset) + int(length)
		if end > len(stored) {
			break
		}
		records = append(records, nameRecord{platform, encoding, language, id, stored[offset:end]})
	}
	return records
}

const (
	platformUnicode   = 0
	platformMacintosh = 1
	platformWindows   = 3
)

func (record nameRecord) isUnicode() bool {
	switch record.platform {
	case platformUnicode:
		return true
	case platformWindows:
		return record.encoding == 0 || record.encoding == 1
	}
	return false
}

func (record nameRecord) isMacRoman() bool {
	return record.platform == platformMacintosh && record.encoding == 0
}

func (record nameRecord) isEnglishUS() bool {
	switch record.platform {
	case platformWindows:
		return record.language == 0x0409
	case platformMacintosh:
		return record.encoding == 0 && record.language == 0
	}
	return false
}

func (record nameRecord) text() (string, bool) {
	if record.isUnicode() {
		if len(record.value)%2 != 0 {
			record.value = record.value[:len(record.value)-1]
		}
		units := make([]uint16, len(record.value)/2)
		for index := range units {
			units[index] = binary.BigEndian.Uint16(record.value[2*index:])
		}
		for index := 0; index < len(units); index++ {
			unit := units[index]
			if unit >= 0xD800 && unit < 0xDC00 {
				if index+1 >= len(units) || units[index+1] < 0xDC00 || units[index+1] >= 0xE000 {
					return "", false
				}
				index++
			} else if unit >= 0xDC00 && unit < 0xE000 {
				return "", false
			}
		}
		return string(utf16.Decode(units)), true
	}
	if record.isMacRoman() {
		runes := make([]rune, len(record.value))
		for index, b := range record.value {
			runes[index] = rune(macRoman[b])
		}
		return string(runes), true
	}
	return "", false
}

// familyNames is fontdb's parse_names: typographic family names, else the
// legacy family names; faces without them or a PostScript name are skipped.
func familyNames(tables sfntTables) ([]string, bool) {
	records := nameRecords(tables["name"])
	if records == nil {
		if _, present := tables["name"]; !present {
			return nil, false
		}
	}
	families := collectFamilies(16, records)
	if len(families) == 0 {
		families = collectFamilies(1, records)
	}
	if len(families) == 0 {
		return nil, false
	}
	for _, record := range records {
		if record.id == 6 && (record.isUnicode() || record.isMacRoman()) {
			_, ok := record.text()
			return families, ok
		}
	}
	return nil, false
}

func collectFamilies(id uint16, records []nameRecord) []string {
	var families []string
	english := false
	for _, record := range records {
		if record.id == id && record.isUnicode() {
			if family, ok := record.text(); ok {
				families = append(families, family)
				english = english || record.isEnglishUS()
			}
		}
	}
	if !english {
		for _, record := range records {
			if record.id == id && record.isMacRoman() {
				if family, ok := record.text(); ok {
					families = append(families, family)
					break
				}
			}
		}
	}
	return families
}

// styleMetadata is fontdb's parse_os2 plus the italic angle from post.
func styleMetadata(tables sfntTables) (style faceStyle, weight uint16, stretch uint16) {
	style, weight, stretch = styleNormal, 400, 5
	if os2, ok := tables["OS/2"]; ok {
		version, _ := be16(os2, 0)
		if length, known := os2Lengths[version]; known && len(os2) >= length {
			weight, _ = be16(os2, 4)
			width, _ := be16(os2, 6)
			if width >= 1 && width <= 9 {
				stretch = width
			}
			selection, _ := be16(os2, 62)
			switch {
			case selection&1 != 0:
				style = styleItalic
			case version >= 4 && selection&(1<<9) != 0:
				style = styleOblique
			}
		}
	}
	if post, ok := tables["post"]; ok {
		// fontdb compares post[4..8] with zero; a shorter table compares unequal.
		italic := len(post) < 8 || post[4] != 0 || post[5] != 0 || post[6] != 0 || post[7] != 0
		if style == styleNormal && italic {
			style = styleItalic
		}
	}
	return style, weight, stretch
}

var os2Lengths = map[uint16]int{0: 78, 1: 86, 2: 96, 3: 96, 4: 96, 5: 100}

// cmapLookup is ttf_parser::Face::glyph_index: the first Unicode subtable
// that maps the code point wins; iteration ends at an unparseable subtable.
func cmapLookup(cmap []byte, codePoint rune) (uint16, bool) {
	count, ok := be16(cmap, 2)
	if !ok {
		return 0, false
	}
	for index := 0; index < int(count); index++ {
		at := 4 + 8*index
		platform, ok1 := be16(cmap, at)
		encoding, ok2 := be16(cmap, at+2)
		offset, ok3 := be32(cmap, at+4)
		if !ok1 || !ok2 || !ok3 || int(offset) > len(cmap) {
			return 0, false
		}
		subtable := cmap[offset:]
		format, ok := be16(subtable, 0)
		if !ok || !cmapSubtableParses(format, subtable) {
			return 0, false
		}
		unicode := platform == platformUnicode ||
			(platform == platformWindows && encoding == 1) ||
			(platform == platformWindows && encoding == 10 && (format == 12 || format == 13))
		if !unicode {
			continue
		}
		if glyph, found := cmapSubtableLookup(format, subtable, uint32(codePoint)); found {
			return glyph, true
		}
	}
	return 0, false
}

func cmapSubtableParses(format uint16, data []byte) bool {
	switch format {
	case 0:
		return len(data) >= 6+256
	case 2:
		return len(data) >= 6+512
	case 4:
		segments, ok := be16(data, 6)
		return ok && len(data) >= 16+4*int(segments/2)*2
	case 6:
		count, ok := be16(data, 8)
		return ok && len(data) >= 10+2*int(count)
	case 8:
		return true
	case 10:
		count, ok := be32(data, 16)
		return ok && uint64(len(data)) >= 20+2*uint64(count)
	case 12, 13:
		count, ok := be32(data, 12)
		return ok && uint64(len(data)) >= 16+12*uint64(count)
	case 14:
		count, ok := be32(data, 6)
		return ok && uint64(len(data)) >= 10+11*uint64(count)
	}
	return false
}

func cmapSubtableLookup(format uint16, data []byte, codePoint uint32) (uint16, bool) {
	switch format {
	case 0:
		if codePoint >= 256 || data[6+codePoint] == 0 {
			return 0, false
		}
		return uint16(data[6+codePoint]), true
	case 4:
		return cmapFormat4(data, codePoint)
	case 6:
		if codePoint > 0xFFFF {
			return 0, false
		}
		first, _ := be16(data, 6)
		count, _ := be16(data, 8)
		if uint16(codePoint) < first || int(uint16(codePoint)-first) >= int(count) {
			return 0, false
		}
		return be16(data, 10+2*int(uint16(codePoint)-first))
	case 10:
		first, _ := be32(data, 12)
		count, _ := be32(data, 16)
		if codePoint < first || codePoint-first >= count {
			return 0, false
		}
		return be16(data, 20+2*int(codePoint-first))
	case 12:
		count, _ := be32(data, 12)
		low, high := 0, int(count)
		for low < high {
			middle := (low + high) / 2
			at := 16 + 12*middle
			start, _ := be32(data, at)
			end, _ := be32(data, at+4)
			switch {
			case start > codePoint:
				high = middle
			case end < codePoint:
				low = middle + 1
			default:
				glyph, _ := be32(data, at+8)
				id := uint64(glyph) + uint64(codePoint)
				if id > 0xFFFFFFFF || id < uint64(start) || id-uint64(start) > 0xFFFF {
					return 0, false
				}
				return uint16(id - uint64(start)), true
			}
		}
	case 13:
		count, _ := be32(data, 12)
		for group := 0; group < int(count); group++ {
			at := 16 + 12*group
			start, _ := be32(data, at)
			end, _ := be32(data, at+4)
			if codePoint >= start && codePoint <= end {
				glyph, _ := be32(data, at+8)
				if glyph > 0xFFFF {
					return 0, false
				}
				return uint16(glyph), true
			}
		}
	}
	return 0, false
}

func cmapFormat4(data []byte, codePoint uint32) (uint16, bool) {
	if codePoint > 0xFFFF {
		return 0, false
	}
	point := uint16(codePoint)
	doubled, _ := be16(data, 6)
	segments := int(doubled / 2)
	ends := 14
	starts := ends + 2*segments + 2
	deltas := starts + 2*segments
	offsets := deltas + 2*segments
	low, high := 0, segments
	for high > low {
		index := (low + high) / 2
		end, ok := be16(data, ends+2*index)
		if !ok {
			return 0, false
		}
		if end >= point {
			start, ok := be16(data, starts+2*index)
			if !ok {
				return 0, false
			}
			if start > point {
				high = index
				continue
			}
			rangeOffset, ok1 := be16(data, offsets+2*index)
			delta, ok2 := be16(data, deltas+2*index)
			if !ok1 || !ok2 {
				return 0, false
			}
			if rangeOffset == 0 {
				return point + delta, true
			}
			if rangeOffset == 0xFFFF {
				return 0, false
			}
			distance := (uint32(point) - uint32(start)) * 2
			if distance > 0xFFFF {
				return 0, false
			}
			position := uint16(offsets+2*index) + uint16(distance) + rangeOffset
			glyph, ok := be16(data, int(position))
			if !ok || glyph == 0 {
				return 0, false
			}
			signed := int16(glyph) + int16(delta)
			if signed < 0 {
				return 0, false
			}
			return uint16(signed), true
		}
		low = index + 1
	}
	return 0, false
}

var macRoman = [256]uint16{
	0x0000, 0x0001, 0x0002, 0x0003, 0x0004, 0x0005, 0x0006, 0x0007,
	0x0008, 0x0009, 0x000A, 0x000B, 0x000C, 0x000D, 0x000E, 0x000F,
	0x0010, 0x2318, 0x21E7, 0x2325, 0x2303, 0x0015, 0x0016, 0x0017,
	0x0018, 0x0019, 0x001A, 0x001B, 0x001C, 0x001D, 0x001E, 0x001F,
	0x0020, 0x0021, 0x0022, 0x0023, 0x0024, 0x0025, 0x0026, 0x0027,
	0x0028, 0x0029, 0x002A, 0x002B, 0x002C, 0x002D, 0x002E, 0x002F,
	0x0030, 0x0031, 0x0032, 0x0033, 0x0034, 0x0035, 0x0036, 0x0037,
	0x0038, 0x0039, 0x003A, 0x003B, 0x003C, 0x003D, 0x003E, 0x003F,
	0x0040, 0x0041, 0x0042, 0x0043, 0x0044, 0x0045, 0x0046, 0x0047,
	0x0048, 0x0049, 0x004A, 0x004B, 0x004C, 0x004D, 0x004E, 0x004F,
	0x0050, 0x0051, 0x0052, 0x0053, 0x0054, 0x0055, 0x0056, 0x0057,
	0x0058, 0x0059, 0x005A, 0x005B, 0x005C, 0x005D, 0x005E, 0x005F,
	0x0060, 0x0061, 0x0062, 0x0063, 0x0064, 0x0065, 0x0066, 0x0067,
	0x0068, 0x0069, 0x006A, 0x006B, 0x006C, 0x006D, 0x006E, 0x006F,
	0x0070, 0x0071, 0x0072, 0x0073, 0x0074, 0x0075, 0x0076, 0x0077,
	0x0078, 0x0079, 0x007A, 0x007B, 0x007C, 0x007D, 0x007E, 0x007F,
	0x00C4, 0x00C5, 0x00C7, 0x00C9, 0x00D1, 0x00D6, 0x00DC, 0x00E1,
	0x00E0, 0x00E2, 0x00E4, 0x00E3, 0x00E5, 0x00E7, 0x00E9, 0x00E8,
	0x00EA, 0x00EB, 0x00ED, 0x00EC, 0x00EE, 0x00EF, 0x00F1, 0x00F3,
	0x00F2, 0x00F4, 0x00F6, 0x00F5, 0x00FA, 0x00F9, 0x00FB, 0x00FC,
	0x2020, 0x00B0, 0x00A2, 0x00A3, 0x00A7, 0x2022, 0x00B6, 0x00DF,
	0x00AE, 0x00A9, 0x2122, 0x00B4, 0x00A8, 0x2260, 0x00C6, 0x00D8,
	0x221E, 0x00B1, 0x2264, 0x2265, 0x00A5, 0x00B5, 0x2202, 0x2211,
	0x220F, 0x03C0, 0x222B, 0x00AA, 0x00BA, 0x03A9, 0x00E6, 0x00F8,
	0x00BF, 0x00A1, 0x00AC, 0x221A, 0x0192, 0x2248, 0x2206, 0x00AB,
	0x00BB, 0x2026, 0x00A0, 0x00C0, 0x00C3, 0x00D5, 0x0152, 0x0153,
	0x2013, 0x2014, 0x201C, 0x201D, 0x2018, 0x2019, 0x00F7, 0x25CA,
	0x00FF, 0x0178, 0x2044, 0x20AC, 0x2039, 0x203A, 0xFB01, 0xFB02,
	0x2021, 0x00B7, 0x201A, 0x201E, 0x2030, 0x00C2, 0x00CA, 0x00C1,
	0x00CB, 0x00C8, 0x00CD, 0x00CE, 0x00CF, 0x00CC, 0x00D3, 0x00D4,
	0xF8FF, 0x00D2, 0x00DA, 0x00DB, 0x00D9, 0x0131, 0x02C6, 0x02DC,
	0x00AF, 0x02D8, 0x02D9, 0x02DA, 0x00B8, 0x02DD, 0x02DB, 0x02C7,
}
