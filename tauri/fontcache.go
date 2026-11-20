package tauri

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"hash/crc32"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sync"
)

// The face cache keeps the parsed metadata of every system font file, keyed
// by the file's stat, the way fontconfig caches its scans. The directory walk
// still runs on every start, so the cached result always follows the current
// listing and its order; only the per-file parse is skipped for files whose
// size, mtime, ctime, device and inode are unchanged. Any failure reading,
// decoding or writing the cache only costs the full parse.

// faceCacheVersion must change whenever the cache format or the metadata
// parsed by parseFaces changes.
const faceCacheVersion = 1

var faceCacheMagic = []byte("PBGFACE\x00")

// fileStamp is everything about a font file that changes when it is edited,
// replaced or swapped out: size and mtime everywhere, and on unix also the
// ctime (bumped by any write or metadata change, even with a restored mtime)
// and the device and inode (changed by a replace-by-rename).
type fileStamp struct {
	size, mtime, ctime, device, inode uint64
}

// cachedFile is one font file's parse: its faces in collection order.
type cachedFile struct {
	path  string
	stamp fileStamp
	faces []cachedFace
}

type cachedFace struct {
	families []string
	style    faceStyle
	weight   uint16
	stretch  uint16
	index    uint32
}

// cachedFaces is uncachedFaces(paths), reusing the cached parse of every file
// whose stamp is unchanged. Unchanged files are not opened: their sources
// load lazily, as with any other face.
func cachedFaces(paths []string, cacheFile string) []*faceInfo {
	previous := readFaceCache(cacheFile)
	known := make(map[string]cachedFile, len(previous))
	for _, file := range previous {
		known[file.path] = file
	}
	stamps := make([]fileStamp, len(paths))
	storable := make([]bool, len(paths))
	statInParallel(len(paths), func(index int) {
		stamps[index], storable[index] = statFont(paths[index])
	})
	files := make([]cachedFile, len(paths))
	sources := make([]*fontSource, len(paths))
	var stale []int
	for index, path := range paths {
		stamp := stamps[index]
		if cached, hit := known[path]; storable[index] && hit && cached.stamp == stamp {
			files[index], sources[index] = cached, &fontSource{path: path}
			continue
		}
		files[index] = cachedFile{path: path, stamp: stamp}
		stale = append(stale, index)
	}
	parseInParallel(len(stale), func(position int) {
		index := stale[position]
		sources[index] = &fontSource{path: paths[index]}
		files[index].faces = describeFaces(parseFaces(sources[index]))
	})
	if len(stale) != 0 || !samePaths(previous, paths) {
		writeFaceCache(cacheFile, storableFiles(files, storable))
	}
	return facesOf(files, sources)
}

// statInParallel strides count stats over GOMAXPROCS workers. Each stat is
// a microsecond or two of syscall, so handing them out one by one through a
// channel as parseInParallel does would cost as much as the stats.
func statInParallel(count int, stat func(index int)) {
	workers := runtime.GOMAXPROCS(0)
	var group sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for index := worker; index < count; index += workers {
				stat(index)
			}
		}()
	}
	group.Wait()
}

func describeFaces(faces []*faceInfo) []cachedFace {
	described := make([]cachedFace, len(faces))
	for position, face := range faces {
		described[position] = cachedFace{families: face.families, style: face.style, weight: face.weight, stretch: face.stretch, index: face.index}
	}
	return described
}

// facesOf rebuilds the face list in walk order, every face of a file sharing
// the file's source as parseFaces does.
func facesOf(files []cachedFile, sources []*fontSource) []*faceInfo {
	var faces []*faceInfo
	for index, file := range files {
		for _, face := range file.faces {
			faces = append(faces, &faceInfo{
				families: face.families,
				style:    face.style,
				weight:   face.weight,
				stretch:  face.stretch,
				source:   sources[index],
				index:    face.index,
			})
		}
	}
	return faces
}

func samePaths(files []cachedFile, paths []string) bool {
	return slices.EqualFunc(files, paths, func(file cachedFile, path string) bool { return file.path == path })
}

// storableFiles drops files that could not be stamped, so they are parsed
// again next time rather than trusted.
func storableFiles(files []cachedFile, storable []bool) []cachedFile {
	var kept []cachedFile
	for index, file := range files {
		if storable[index] {
			kept = append(kept, file)
		}
	}
	return kept
}

// faceCacheKey is every input of the walk besides the file system itself, so
// a cache written for a different scan is never read.
func faceCacheKey() string {
	return string(binary.AppendUvarint(nil, faceCacheVersion)) + "\x00" + runtime.GOOS +
		"\x00HOME=" + os.Getenv("HOME") +
		"\x00SYSTEMROOT=" + os.Getenv("SYSTEMROOT") +
		"\x00USERPROFILE=" + os.Getenv("USERPROFILE")
}

// faceCacheFile is the cache for the current key under the user cache
// directory; each key has its own file so differing environments do not
// overwrite each other's cache. It is empty when there is no cache directory.
func faceCacheFile() string {
	directory, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	digest := sha256.Sum256([]byte(faceCacheKey()))
	return filepath.Join(directory, "iris-go", "tauri-fonts", "faces-"+hex.EncodeToString(digest[:12])+".bin")
}

// readFaceCache returns the cached files in walk order, or nothing when the
// cache is missing, corrupt, truncated or written for another key.
func readFaceCache(cacheFile string) (files []cachedFile) {
	if cacheFile == "" {
		return nil
	}
	defer func() {
		if recover() != nil {
			files = nil
		}
	}()
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil
	}
	return decodeFaceCache(data, faceCacheKey())
}

// writeFaceCache replaces the cache atomically: a complete temporary file is
// renamed over it, so concurrent readers and writers only ever see a whole
// cache from one of the writers.
func writeFaceCache(cacheFile string, files []cachedFile) {
	if cacheFile == "" {
		return
	}
	defer func() { _ = recover() }()
	directory := filepath.Dir(cacheFile)
	if os.MkdirAll(directory, 0o755) != nil {
		return
	}
	temporary, err := os.CreateTemp(directory, filepath.Base(cacheFile)+".*.tmp")
	if err != nil {
		return
	}
	_, writeErr := temporary.Write(encodeFaceCache(files, faceCacheKey()))
	closeErr := temporary.Close()
	if writeErr != nil || closeErr != nil || os.Rename(temporary.Name(), cacheFile) != nil {
		_ = os.Remove(temporary.Name())
	}
}

// The cache file is the magic, the key, the files, and a CRC-32 of all that.
// Strings and counts are uvarint-prefixed, stamps fixed 64-bit little endian.

func encodeFaceCache(files []cachedFile, key string) []byte {
	out := append([]byte(nil), faceCacheMagic...)
	out = appendString(out, key)
	out = binary.AppendUvarint(out, uint64(len(files)))
	for _, file := range files {
		out = appendString(out, file.path)
		for _, value := range [...]uint64{file.stamp.size, file.stamp.mtime, file.stamp.ctime, file.stamp.device, file.stamp.inode} {
			out = binary.LittleEndian.AppendUint64(out, value)
		}
		out = binary.AppendUvarint(out, uint64(len(file.faces)))
		for _, face := range file.faces {
			out = binary.AppendUvarint(out, uint64(len(face.families)))
			for _, family := range face.families {
				out = appendString(out, family)
			}
			out = append(out, byte(face.style))
			out = binary.LittleEndian.AppendUint16(out, face.weight)
			out = binary.LittleEndian.AppendUint16(out, face.stretch)
			out = binary.LittleEndian.AppendUint32(out, face.index)
		}
	}
	return binary.LittleEndian.AppendUint32(out, crc32.ChecksumIEEE(out))
}

func appendString(out []byte, text string) []byte {
	return append(binary.AppendUvarint(out, uint64(len(text))), text...)
}

func decodeFaceCache(data []byte, key string) []cachedFile {
	if len(data) < len(faceCacheMagic)+4 || !bytes.HasPrefix(data, faceCacheMagic) {
		return nil
	}
	body, sum := data[:len(data)-4], data[len(data)-4:]
	if crc32.ChecksumIEEE(body) != binary.LittleEndian.Uint32(sum) {
		return nil
	}
	in := &cacheReader{data: body, position: len(faceCacheMagic)}
	if in.string() != key {
		return nil
	}
	files := make([]cachedFile, in.count(len(body)))
	for index := range files {
		file := &files[index]
		file.path = in.string()
		file.stamp = fileStamp{size: in.uint64(), mtime: in.uint64(), ctime: in.uint64(), device: in.uint64(), inode: in.uint64()}
		file.faces = make([]cachedFace, in.count(len(body)))
		for position := range file.faces {
			face := &file.faces[position]
			face.families = make([]string, in.count(len(body)))
			for family := range face.families {
				face.families[family] = in.string()
			}
			face.style = faceStyle(in.byte())
			face.weight = in.uint16()
			face.stretch = in.uint16()
			face.index = in.uint32()
		}
		if in.failed {
			return nil
		}
	}
	if in.failed || in.position != len(body) {
		return nil
	}
	return files
}

// cacheReader decodes a cache body; once anything is out of bounds it fails
// and yields zeros, and the caller discards the whole cache.
type cacheReader struct {
	data     []byte
	position int
	failed   bool
}

func (in *cacheReader) take(size int) []byte {
	if in.failed || size < 0 || size > len(in.data)-in.position {
		in.failed = true
		return nil
	}
	taken := in.data[in.position : in.position+size]
	in.position += size
	return taken
}

// count is a length prefix, bounded by the size of the cache so a corrupt
// count cannot allocate more than the file could describe.
func (in *cacheReader) count(limit int) int {
	value, size := binary.Uvarint(in.data[min(in.position, len(in.data)):])
	if in.failed || size <= 0 || value > uint64(limit) {
		in.failed = true
		return 0
	}
	in.position += size
	return int(value)
}

func (in *cacheReader) string() string {
	return string(in.take(in.count(len(in.data))))
}

func (in *cacheReader) byte() byte {
	if taken := in.take(1); taken != nil {
		return taken[0]
	}
	return 0
}

func (in *cacheReader) uint16() uint16 {
	if taken := in.take(2); taken != nil {
		return binary.LittleEndian.Uint16(taken)
	}
	return 0
}

func (in *cacheReader) uint32() uint32 {
	if taken := in.take(4); taken != nil {
		return binary.LittleEndian.Uint32(taken)
	}
	return 0
}

func (in *cacheReader) uint64() uint64 {
	if taken := in.take(8); taken != nil {
		return binary.LittleEndian.Uint64(taken)
	}
	return 0
}
