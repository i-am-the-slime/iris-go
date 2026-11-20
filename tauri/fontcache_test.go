package tauri

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestFaceCacheMatchesFullScanOfSystemFonts(t *testing.T) {
	paths := systemFontPaths()
	if len(paths) == 0 {
		t.Skip("no system fonts")
	}
	cacheFile := filepath.Join(t.TempDir(), "faces.bin")
	want := describeDatabase(uncachedFaces(paths))
	expectFaces(t, "cold", cachedFaces(paths, cacheFile), want)
	written := readBytes(t, cacheFile)
	warm := cachedFaces(paths, cacheFile)
	expectFaces(t, "warm", warm, want)
	for _, face := range warm {
		if face.source.data != nil {
			t.Fatalf("warm cache opened %s", face.source.path)
		}
	}
	if !bytes.Equal(readBytes(t, cacheFile), written) {
		t.Fatal("an unchanged scan rewrote the cache")
	}
}

func TestFaceCacheMatchesFullScanOfFontTree(t *testing.T) {
	tree := newFontTree(t)
	tree.add("a.ttf", 0)
	tree.add("nested/deeper/b.TTC", tree.collection())
	tree.add("nested/c.otf", 1)
	tree.add(".ttf", 2)
	tree.write("notes.txt", []byte("not a font"))
	tree.write("broken.ttf", []byte("not a font either"))
	tree.symlink("linked.ttf", "a.ttf")
	tree.symlink("elsewhere", "nested")
	tree.expectCacheMatchesFullScan("cold")
	tree.expectCacheMatchesFullScan("warm")
}

func TestFaceCacheFollowsAddedRemovedAndChangedFonts(t *testing.T) {
	tree := newFontTree(t)
	tree.add("a.ttf", 0)
	tree.add("sub/b.ttf", 1)
	tree.expectCacheMatchesFullScan("initial")

	tree.add("sub/c.ttf", 2)
	tree.expectCacheMatchesFullScan("added")

	tree.remove("sub/b.ttf")
	tree.expectCacheMatchesFullScan("removed")
	tree.expectCacheHoldsOnlyTheCurrentFiles()

	tree.add("a.ttf", 2)
	tree.expectCacheMatchesFullScan("replaced in place")

	tree.symlink("sub/link.ttf", "a.ttf")
	tree.expectCacheMatchesFullScan("symlink added")
	tree.remove("sub/link.ttf")
	tree.symlink("sub/link.ttf", "sub/c.ttf")
	tree.expectCacheMatchesFullScan("symlink retargeted")

	tree.touch("sub/c.ttf")
	tree.expectCacheMatchesFullScan("touched")
	tree.expectCachedStampIsCurrent("sub/c.ttf")
}

func TestFaceCacheTrustsOnlyAnIdenticalStamp(t *testing.T) {
	tree := newFontTree(t)
	path := tree.add("a.ttf", 0)
	stamp, _ := statFont(path)
	poisoned := func(stamp fileStamp) []*faceInfo {
		writeFaceCache(tree.cacheFile, []cachedFile{{path: path, stamp: stamp, faces: []cachedFace{{families: []string{"Cached"}, weight: 400, stretch: 5}}}})
		return cachedFaces([]string{path}, tree.cacheFile)
	}
	if faces := poisoned(stamp); faces[0].families[0] != "Cached" {
		t.Fatal("an unchanged file was parsed again")
	}
	for name, changed := range map[string]func(*fileStamp){
		"size":   func(stamp *fileStamp) { stamp.size++ },
		"mtime":  func(stamp *fileStamp) { stamp.mtime++ },
		"ctime":  func(stamp *fileStamp) { stamp.ctime++ },
		"device": func(stamp *fileStamp) { stamp.device++ },
		"inode":  func(stamp *fileStamp) { stamp.inode++ },
	} {
		other := stamp
		changed(&other)
		if faces := poisoned(other); faces[0].families[0] == "Cached" {
			t.Fatalf("a file with a different %s was trusted", name)
		}
	}
}

func TestCorruptFaceCacheFallsBackToFullScan(t *testing.T) {
	tree := newFontTree(t)
	tree.add("a.ttf", 0)
	tree.add("b.ttc", tree.collection())
	tree.expectCacheMatchesFullScan("valid")
	valid := readBytes(t, tree.cacheFile)
	flipped := bytes.Clone(valid)
	flipped[len(flipped)/2] ^= 0xff
	paths := tree.paths()
	corruptions := map[string][]byte{
		"empty":        {},
		"garbage":      []byte("garbage"),
		"magic only":   faceCacheMagic,
		"truncated":    valid[:len(valid)/2],
		"flipped byte": flipped,
		"other key":    encodeFaceCache(readFaceCache(tree.cacheFile), faceCacheKey()+"elsewhere"),
		"huge count":   withChecksum(append(appendString(bytes.Clone(faceCacheMagic), faceCacheKey()), 0xff, 0xff, 0xff, 0xff, 0x0f)),
		"trailing":     withChecksum(append(bytes.Clone(valid[:len(valid)-4]), 0)),
		"bad paths":    encodeFaceCache([]cachedFile{{path: paths[0]}}, faceCacheKey()),
	}
	for name, corrupt := range corruptions {
		if err := os.WriteFile(tree.cacheFile, corrupt, 0o644); err != nil {
			t.Fatal(err)
		}
		tree.expectCacheMatchesFullScan(name)
		if readFaceCache(tree.cacheFile) == nil {
			t.Fatalf("%s: the cache was not rebuilt", name)
		}
	}
}

func TestUnwritableFaceCacheStillScans(t *testing.T) {
	tree := newFontTree(t)
	tree.add("a.ttf", 0)
	tree.write("blocker", []byte("a file where the cache directory should be"))
	paths := tree.paths()
	expectFaces(t, "unwritable", cachedFaces(paths, filepath.Join(tree.root, "blocker", "faces.bin")), describeDatabase(uncachedFaces(paths)))
	expectFaces(t, "no cache directory", cachedFaces(paths, ""), describeDatabase(uncachedFaces(paths)))
}

func TestConcurrentFaceCacheWritersLeaveAWholeCache(t *testing.T) {
	tree := newFontTree(t)
	tree.add("a.ttf", 0)
	tree.add("b.ttf", 1)
	tree.add("c.ttc", tree.collection())
	paths := tree.paths()
	want := describeDatabase(uncachedFaces(paths))
	versions := [][]cachedFile{
		{{path: paths[0], faces: []cachedFace{{families: []string{"One"}}}}},
		{{path: paths[1], faces: []cachedFace{{families: []string{"Two", "Deux"}}}}, {path: paths[2]}},
		fullCache(tree.t, paths),
	}
	writeFaceCache(tree.cacheFile, versions[0])
	var group sync.WaitGroup
	failures := make(chan string, 1000)
	for worker := 0; worker < 16; worker++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for round := 0; round < 20; round++ {
				switch round % 3 {
				case 0:
					expectFacesTo(failures, "racing scan", cachedFaces(paths, tree.cacheFile), want)
				case 1:
					writeFaceCache(tree.cacheFile, versions[(worker+round)%2])
				default:
					if files := readFaceCache(tree.cacheFile); !isWholeVersion(files, versions) {
						failures <- "read a missing, torn or foreign cache"
					}
				}
			}
		}()
	}
	group.Wait()
	close(failures)
	for failure := range failures {
		t.Fatal(failure)
	}
	tree.expectCacheMatchesFullScan("after the race")
	leftovers, _ := filepath.Glob(tree.cacheFile + ".*")
	if len(leftovers) != 0 {
		t.Fatalf("temporary files left behind: %v", leftovers)
	}
}

func TestConcurrentProcessesShareOneFaceCache(t *testing.T) {
	if root := os.Getenv("FACE_CACHE_CHILD_ROOT"); root != "" {
		var paths []string
		(&directoryScan{seen: map[string]bool{}, files: &paths}).directory(root)
		cachedFaces(paths, os.Getenv("FACE_CACHE_CHILD_FILE"))
		return
	}
	tree := newFontTree(t)
	tree.add("a.ttf", 0)
	tree.add("b.ttc", tree.collection())
	tree.add("c.ttf", 1)
	var children []*exec.Cmd
	for child := 0; child < 12; child++ {
		command := exec.Command(os.Args[0], "-test.run=^TestConcurrentProcessesShareOneFaceCache$")
		command.Env = append(os.Environ(), "FACE_CACHE_CHILD_ROOT="+tree.root, "FACE_CACHE_CHILD_FILE="+tree.cacheFile)
		if err := command.Start(); err != nil {
			t.Fatal(err)
		}
		children = append(children, command)
	}
	for _, command := range children {
		if err := command.Wait(); err != nil {
			t.Fatal(err)
		}
	}
	if !reflect.DeepEqual(normalized(readFaceCache(tree.cacheFile)), normalized(fullCache(t, tree.paths()))) {
		t.Fatal("racing processes left a cache other than a full scan's")
	}
	tree.expectCacheMatchesFullScan("after the processes")
}

func TestFaceCacheKeyCoversTheScanInputs(t *testing.T) {
	key := faceCacheKey()
	for _, variable := range []string{"HOME", "SYSTEMROOT", "USERPROFILE"} {
		t.Setenv(variable, os.Getenv(variable)+"/elsewhere")
		if faceCacheKey() == key || faceCacheFile() == "" {
			t.Fatalf("%s does not change the cache key", variable)
		}
		key = faceCacheKey()
	}
}

// faceDescription is everything about a face that matching depends on, and
// which face before it shares its file.
type faceDescription struct {
	families               []string
	style                  faceStyle
	weight, stretch        uint16
	path                   string
	index                  uint32
	features               string
	sharesSourceWithFaceAt int
}

func describeDatabase(faces []*faceInfo) []faceDescription {
	described := make([]faceDescription, len(faces))
	firstOfSource := map[*fontSource]int{}
	for position, face := range faces {
		if _, seen := firstOfSource[face.source]; !seen {
			firstOfSource[face.source] = position
		}
		described[position] = faceDescription{
			families: face.families, style: face.style, weight: face.weight, stretch: face.stretch,
			path: face.source.path, index: face.index, features: face.features,
			sharesSourceWithFaceAt: firstOfSource[face.source],
		}
	}
	return described
}

func expectFaces(t *testing.T, name string, faces []*faceInfo, want []faceDescription) {
	t.Helper()
	failures := make(chan string, 1)
	expectFacesTo(failures, name, faces, want)
	close(failures)
	for failure := range failures {
		t.Fatal(failure)
	}
}

func expectFacesTo(failures chan<- string, name string, faces []*faceInfo, want []faceDescription) {
	if got := describeDatabase(faces); !reflect.DeepEqual(got, want) {
		failures <- name + ": cached faces differ from a full scan"
	}
}

func isWholeVersion(files []cachedFile, versions [][]cachedFile) bool {
	for _, version := range versions {
		if reflect.DeepEqual(normalized(files), normalized(version)) {
			return true
		}
	}
	return false
}

// fullCache is the cache a scan of paths writes, as racing scans write it.
func fullCache(t *testing.T, paths []string) []cachedFile {
	cacheFile := filepath.Join(t.TempDir(), "faces.bin")
	cachedFaces(paths, cacheFile)
	return readFaceCache(cacheFile)
}

// normalized treats empty and nil face lists alike, as decoding does.
func normalized(files []cachedFile) []cachedFile {
	out := make([]cachedFile, len(files))
	for index, file := range files {
		out[index] = cachedFile{path: file.path, stamp: file.stamp}
		for _, face := range file.faces {
			out[index].faces = append(out[index].faces, face)
		}
	}
	return out
}

func withChecksum(body []byte) []byte {
	return binary.LittleEndian.AppendUint32(body, crc32.ChecksumIEEE(body))
}

func readBytes(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// fontTree is a font directory under the test's control, filled with copies
// of system fonts, and a face cache for it.
type fontTree struct {
	t         *testing.T
	root      string
	cacheFile string
	fonts     []string
}

func newFontTree(t *testing.T) *fontTree {
	fonts := singleFaceSystemFonts(3)
	if len(fonts) < 3 {
		t.Skip("needs three system fonts to copy")
	}
	directory := t.TempDir()
	return &fontTree{t: t, root: filepath.Join(directory, "fonts"), cacheFile: filepath.Join(directory, "cache", "faces.bin"), fonts: fonts}
}

func (tree *fontTree) paths() []string {
	var paths []string
	(&directoryScan{seen: map[string]bool{}, files: &paths}).directory(tree.root)
	return paths
}

func (tree *fontTree) expectCacheMatchesFullScan(name string) {
	tree.t.Helper()
	paths := tree.paths()
	expectFaces(tree.t, name, cachedFaces(paths, tree.cacheFile), describeDatabase(uncachedFaces(paths)))
}

func (tree *fontTree) expectCacheHoldsOnlyTheCurrentFiles() {
	tree.t.Helper()
	if !samePaths(readFaceCache(tree.cacheFile), tree.paths()) {
		tree.t.Fatal("the cache holds other files than the scan")
	}
}

func (tree *fontTree) expectCachedStampIsCurrent(name string) {
	tree.t.Helper()
	path := filepath.Join(tree.root, name)
	current, _ := statFont(path)
	for _, file := range readFaceCache(tree.cacheFile) {
		if file.path == path && file.stamp == current {
			return
		}
	}
	tree.t.Fatalf("the cache does not hold the current stamp of %s", name)
}

// add copies the font'th system font to name, replacing any file there by
// writing over it in place.
func (tree *fontTree) add(name string, font int) string {
	return tree.write(name, readBytes(tree.t, tree.fonts[font]))
}

// collection is the index of the first multi-face system collection, which
// the tree then copies like any other font.
func (tree *fontTree) collection() int {
	for _, path := range systemFontPaths() {
		if strings.HasSuffix(strings.ToLower(path), ".ttc") && len(parseFaces(&fontSource{path: path})) > 1 {
			tree.fonts = append(tree.fonts, path)
			return len(tree.fonts) - 1
		}
	}
	tree.t.Skip("no system font collection to copy")
	return 0
}

func (tree *fontTree) write(name string, data []byte) string {
	path := filepath.Join(tree.root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		tree.t.Fatal(err)
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		tree.t.Fatal(err)
	}
	if _, err := file.Write(data); err != nil {
		tree.t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		tree.t.Fatal(err)
	}
	return path
}

func (tree *fontTree) symlink(name string, target string) {
	if err := os.Symlink(filepath.Join(tree.root, target), filepath.Join(tree.root, name)); err != nil {
		tree.t.Fatal(err)
	}
}

func (tree *fontTree) remove(name string) {
	if err := os.Remove(filepath.Join(tree.root, name)); err != nil {
		tree.t.Fatal(err)
	}
}

func (tree *fontTree) touch(name string) {
	later := time.Now().Add(time.Hour)
	if err := os.Chtimes(filepath.Join(tree.root, name), later, later); err != nil {
		tree.t.Fatal(err)
	}
}

// singleFaceSystemFonts finds system fonts with one face and distinct
// families, small ones first so the trees stay cheap to build.
func singleFaceSystemFonts(count int) []string {
	var fonts []string
	seen := map[string]bool{}
	for _, path := range systemFontPaths() {
		info, err := os.Stat(path)
		if err != nil || info.Size() > 512<<10 {
			continue
		}
		faces := parseFaces(&fontSource{path: path})
		if len(faces) != 1 || seen[faces[0].families[0]] {
			continue
		}
		seen[faces[0].families[0]] = true
		fonts = append(fonts, path)
		if len(fonts) == count {
			break
		}
	}
	return fonts
}
