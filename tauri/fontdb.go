package tauri

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// The font database of the native host (fontdb 0.23): the application's font
// assets under their configured names, followed by every system face in
// fontdb's scan order. Matching, fallback order and tie-breaking all depend
// on that order, so the scan reproduces fontdb's directory walk exactly.

type faceStyle int

const (
	styleNormal faceStyle = iota
	styleItalic
	styleOblique
)

type faceInfo struct {
	families []string
	style    faceStyle
	weight   uint16
	stretch  uint16
	source   *fontSource
	index    uint32
	features string
}

type fontSource struct {
	path string
	once sync.Once
	data []byte
}

func (source *fontSource) bytes() []byte {
	source.once.Do(func() {
		if source.data == nil {
			source.data = mapFile(source.path)
		}
	})
	return source.data
}

type fontDatabase struct {
	faces []*faceInfo
}

// newFontDatabase mirrors Graphics::new: assets take their configured family
// (ahead of the font's own names), weight and style, then system fonts follow.
func newFontDatabase(assets []FontAsset) (*fontDatabase, error) {
	database := &fontDatabase{}
	for _, asset := range assets {
		source := &fontSource{data: asset.Data}
		faces := parseFaces(source)
		if len(faces) == 0 {
			return nil, hostError("invalid font asset: " + asset.Family)
		}
		for _, face := range faces {
			face.families = append([]string{asset.Family}, face.families...)
			face.weight = asset.Weight
			face.style = styleNormal
			if asset.Italic {
				face.style = styleItalic
			}
			face.features = asset.Features
			database.faces = append(database.faces, face)
		}
	}
	database.faces = append(database.faces, systemFaces()...)
	if len(database.faces) == 0 {
		return nil, hostError("no fonts are available")
	}
	return database, nil
}

// parseFaces is fontdb's per-source loop: every face of a collection that
// parses and is named.
func parseFaces(source *fontSource) []*faceInfo {
	data := source.bytes()
	count, collection := facesInCollection(data)
	if !collection {
		count = 1
	}
	var faces []*faceInfo
	for index := uint32(0); index < count; index++ {
		tables, ok := parseRawFace(data, index)
		if !ok {
			continue
		}
		families, ok := familyNames(tables)
		if !ok {
			continue
		}
		style, weight, stretch := styleMetadata(tables)
		faces = append(faces, &faceInfo{
			families: families,
			style:    style,
			weight:   weight,
			stretch:  stretch,
			source:   source,
			index:    index,
		})
	}
	return faces
}

// systemFaces is every system face in fontdb's scan order. The walk itself
// runs every time; the faces of files unchanged since the last process come
// from the on-disk face cache instead of being opened and parsed.
func systemFaces() []*faceInfo {
	return cachedFaces(systemFontPaths(), faceCacheFile())
}

// systemFontPaths is fontdb's load_system_fonts: the font files of the
// platform's font directories, in the order fontdb visits them.
func systemFontPaths() []string {
	var paths []string
	scan := &directoryScan{seen: map[string]bool{}, files: &paths}
	home := os.Getenv("HOME")
	switch runtime.GOOS {
	case "darwin":
		scan.directory("/Library/Fonts")
		scan.directory("/System/Library/Fonts")
		for _, entry := range readDirectory("/System/Library/AssetsV2") {
			if strings.HasPrefix(entry.Name(), "com_apple_MobileAsset_Font") {
				scan.directory(filepath.Join("/System/Library/AssetsV2", entry.Name()))
			}
		}
		scan.directory("/Network/Library/Fonts")
		if home != "" {
			scan.directory(filepath.Join(home, "Library/Fonts"))
		}
	case "windows":
		root := os.Getenv("SYSTEMROOT")
		if root == "" {
			scan.directory(`C:\Windows\Fonts\`)
		} else {
			scan.directory(filepath.Join(root, "Fonts"))
		}
		if profile := os.Getenv("USERPROFILE"); profile != "" {
			scan.directory(filepath.Join(profile, `AppData\Local\Microsoft\Windows\Fonts`))
			scan.directory(filepath.Join(profile, `AppData\Roaming\Microsoft\Windows\Fonts`))
		}
	default:
		scan.directory("/usr/share/fonts/")
		scan.directory("/usr/local/share/fonts/")
		if home != "" {
			scan.directory(filepath.Join(home, ".fonts"))
			scan.directory(filepath.Join(home, ".local/share/fonts"))
		}
	}
	return paths
}

// uncachedFaces opens and parses every file, as fontdb does.
func uncachedFaces(paths []string) []*faceInfo {
	parsed := make([][]*faceInfo, len(paths))
	parseInParallel(len(paths), func(index int) {
		parsed[index] = parseFaces(&fontSource{path: paths[index]})
	})
	var faces []*faceInfo
	for _, file := range parsed {
		faces = append(faces, file...)
	}
	return faces
}

// parseInParallel runs parse for 0..count-1 on GOMAXPROCS workers.
func parseInParallel(count int, parse func(index int)) {
	var group sync.WaitGroup
	work := make(chan int)
	for worker := 0; worker < runtime.GOMAXPROCS(0); worker++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for index := range work {
				parse(index)
			}
		}()
	}
	for index := 0; index < count; index++ {
		work <- index
	}
	close(work)
	group.Wait()
}

// directoryScan is fontdb's load_fonts_dir_impl with its symlink handling:
// plain entries are deduplicated by path only once a symlink has been seen,
// symlinks by their canonical target.
type directoryScan struct {
	seen  map[string]bool
	files *[]string
}

func (scan *directoryScan) directory(directory string) {
	for _, entry := range readDirectory(directory) {
		path := filepath.Join(directory, entry.Name())
		var directoryEntry, regular bool
		if entry.Type()&os.ModeSymlink == 0 {
			if len(scan.seen) != 0 {
				if scan.seen[path] {
					continue
				}
				scan.seen[path] = true
			}
			directoryEntry, regular = entry.IsDir(), entry.Type().IsRegular()
		} else {
			stat, err := os.Stat(path)
			if err != nil {
				continue
			}
			canonical, err := filepath.EvalSymlinks(path)
			if err != nil {
				continue
			}
			if absolute, err := filepath.Abs(canonical); err == nil {
				canonical = absolute
			}
			if scan.seen[canonical] {
				continue
			}
			scan.seen[canonical] = true
			path = canonical
			directoryEntry, regular = stat.IsDir(), stat.Mode().IsRegular()
		}
		if regular {
			switch fontExtension(path) {
			case "ttf", "ttc", "TTF", "TTC", "otf", "otc", "OTF", "OTC":
				*scan.files = append(*scan.files, path)
			}
		} else if directoryEntry {
			scan.directory(path)
		}
	}
}

// fontExtension is Rust's Path::extension: none for dotfiles like ".ttf".
func fontExtension(path string) string {
	name := filepath.Base(path)
	dot := strings.LastIndexByte(name, '.')
	if dot <= 0 {
		return ""
	}
	return name[dot+1:]
}

// readDirectory lists entries in the order the OS returns them, as Rust's
// read_dir does (os.ReadDir would sort them).
func readDirectory(directory string) []os.DirEntry {
	handle, err := os.Open(directory)
	if err != nil {
		return nil
	}
	defer handle.Close()
	entries, _ := handle.ReadDir(-1)
	return entries
}

// genericFamily is fontdb's default family for each CSS generic family.
func genericFamily(name string) (string, bool) {
	switch name {
	case "serif":
		return "Times New Roman", true
	case "sans-serif":
		return "Arial", true
	case "cursive":
		return "Comic Sans MS", true
	case "fantasy":
		if runtime.GOOS == "darwin" {
			return "Papyrus", true
		}
		return "Impact", true
	case "monospace":
		return "Courier New", true
	}
	return "", false
}

type fontQuery struct {
	families []string
	weight   uint16
	stretch  uint16
	style    faceStyle
}

// query is fontdb's Database::query: the first family with any candidate
// decides, by CSS font matching among that family's faces.
func (database *fontDatabase) query(query fontQuery) *faceInfo {
	for _, family := range query.families {
		var candidates []*faceInfo
		for _, face := range database.faces {
			for _, name := range face.families {
				if name == family {
					candidates = append(candidates, face)
					break
				}
			}
		}
		if len(candidates) != 0 {
			if best := bestMatch(candidates, query); best != nil {
				return best
			}
		}
	}
	return nil
}

// bestMatch is fontdb's find_best_match (CSS Fonts 3 §5.2 step 4).
func bestMatch(candidates []*faceInfo, query fontQuery) *faceInfo {
	set := candidates
	stretch := query.stretch
	if !anyFace(set, func(face *faceInfo) bool { return face.stretch == stretch }) {
		narrower := func(face *faceInfo) bool { return face.stretch < query.stretch }
		wider := func(face *faceInfo) bool { return face.stretch > query.stretch }
		distance := func(face *faceInfo) int { return absInt(int(face.stretch) - int(query.stretch)) }
		if query.stretch <= 5 {
			stretch = closest(set, narrower, distance, wider)
		} else {
			stretch = closest(set, wider, distance, narrower)
		}
	}
	set = filterFaces(set, func(face *faceInfo) bool { return face.stretch == stretch })

	preference := map[faceStyle][3]faceStyle{
		styleItalic:  {styleItalic, styleOblique, styleNormal},
		styleOblique: {styleOblique, styleItalic, styleNormal},
		styleNormal:  {styleNormal, styleOblique, styleItalic},
	}[query.style]
	found := false
	var style faceStyle
	for _, wanted := range preference {
		if anyFace(set, func(face *faceInfo) bool { return face.style == wanted }) {
			style, found = wanted, true
			break
		}
	}
	if !found {
		return nil
	}
	set = filterFaces(set, func(face *faceInfo) bool { return face.style == style })

	weight := query.weight
	has := func(value uint16) bool {
		return anyFace(set, func(face *faceInfo) bool { return face.weight == value })
	}
	switch {
	case has(weight):
	case weight >= 400 && weight < 450 && has(500):
		weight = 500
	case weight >= 450 && weight <= 500 && has(400):
		weight = 400
	case weight <= 500:
		weight = closestWeight(set, query.weight, false)
	default:
		weight = closestWeight(set, query.weight, true)
	}
	set = filterFaces(set, func(face *faceInfo) bool { return face.weight == weight })
	if len(set) == 0 {
		return nil
	}
	return set[0]
}

// closest picks the first face minimising the distance among those passing
// the preferred filter, else among all faces (Rust's min_by_key keeps the
// first minimum). The second filter only documents fontdb's fallback order.
func closest(set []*faceInfo, preferred func(*faceInfo) bool, distance func(*faceInfo) int, _ func(*faceInfo) bool) uint16 {
	if face := minimum(filterFaces(set, preferred), distance); face != nil {
		return face.stretch
	}
	return minimum(set, distance).stretch
}

// closestWeight is the weight half of step 4c: thinner first (or fatter
// first when heavier), then the other direction. fontdb subtracts in u16, so
// the fallback direction compares wrapped differences.
func closestWeight(set []*faceInfo, weight uint16, heavier bool) uint16 {
	if heavier {
		if face := minimum(filterFaces(set, func(face *faceInfo) bool { return face.weight >= weight }),
			func(face *faceInfo) int { return int(face.weight - weight) }); face != nil {
			return face.weight
		}
		return minimum(set, func(face *faceInfo) int { return int(weight - face.weight) }).weight
	}
	if face := minimum(filterFaces(set, func(face *faceInfo) bool { return face.weight <= weight }),
		func(face *faceInfo) int { return int(weight - face.weight) }); face != nil {
		return face.weight
	}
	return minimum(set, func(face *faceInfo) int { return int(face.weight - weight) }).weight
}

func minimum(set []*faceInfo, key func(*faceInfo) int) *faceInfo {
	var best *faceInfo
	bestKey := 0
	for _, face := range set {
		if value := key(face); best == nil || value < bestKey {
			best, bestKey = face, value
		}
	}
	return best
}

func anyFace(set []*faceInfo, predicate func(*faceInfo) bool) bool {
	for _, face := range set {
		if predicate(face) {
			return true
		}
	}
	return false
}

func filterFaces(set []*faceInfo, predicate func(*faceInfo) bool) []*faceInfo {
	var kept []*faceInfo
	for _, face := range set {
		if predicate(face) {
			kept = append(kept, face)
		}
	}
	return kept
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
