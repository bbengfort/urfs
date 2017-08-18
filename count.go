package urfs

import (
	"fmt"
	"os"
	"sync/atomic"
)

// Count the number of files and the number of bytes in each of the specified
// paths. Returns a struct with the count and size that can compute the mean
// and human readable representation of the result.
func (fs *FSWalker) Count(print bool, paths ...string) ([]*DirSize, error) {
	sizes := make([]*DirSize, 0, len(paths))
	for _, path := range paths {
		size := &DirSize{Path: path}
		if err := fs.Walk(path, size.Update); err != nil {
			return nil, err
		}
		sizes = append(sizes, size)

		if print {
			fmt.Println(size.String())
		}

		fs.Reset(nil)
	}
	return sizes, nil
}

// DirSize holds the number of files and bytes in a given directory.
type DirSize struct {
	Path  string // path to the directory
	Files uint64 // number of files in the directory
	Bytes uint64 // number of bytes in the directory
}

// Update the directory info from the given path, synchronizing as necessary.
func (s *DirSize) Update(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	if info.IsDir() {
		return "", nil
	}

	size := info.Size()
	if size <= 0 {
		return "", nil
	}

	atomic.AddUint64(&s.Files, 1)
	atomic.AddUint64(&s.Bytes, uint64(size))
	return path, nil
}

// Mean returns the average number of bytes per file
func (s *DirSize) Mean() float64 {
	return float64(s.Bytes) / float64(s.Files)
}

// String returns a string representation of the size
func (s *DirSize) String() string {
	return fmt.Sprintf("%s: %d files %d bytes (%0.0f bytes/file)", s.Path, s.Files, s.Bytes, s.Mean())
}
