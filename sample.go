package urfs

import (
	"fmt"
	"math/rand"
	"path/filepath"
)

// Sample the files contained in a source directory (src), copying them to a
// destination directory (dst) with some probability between 0 and 1 (size).
// To modify the behavior of the walk, pass in a FSWalker; if nil will use
// the default FSWalker.
func (fs *FSWalker) Sample(src, dst string, size float64) (string, error) {
	// Run the walk with our sampling function
	err := fs.Walk(src, func(path string) (string, error) {
		// If we're in the sample percent, perform the copy
		if rand.Float64() <= size {
			// Get the relative path from the base
			rel, err := filepath.Rel(src, path)
			if err != nil {
				return "", err
			}

			// Create the new path to the destination
			drl := filepath.Join(dst, rel)

			// Create the directory if it doesn't exist
			if err = Mkdir(filepath.Dir(drl)); err != nil {
				return "", err
			}

			// Copy the file to the destination directory
			if err = CopyFile(drl, path, 0644); err != nil {
				return "", err
			}

			// Return the path to the copied file
			return drl, nil
		}

		// No work was done so return empty string
		return "", nil
	})

	// If an error occured return it
	if err != nil {
		return "", err
	}

	// Otherwise return a statement of how much was sampled
	pcent := (float64(fs.nResults) / float64(fs.nPaths)) * 100.0
	result := fmt.Sprintf("sampled %d of %d files (%0.1f%%) in %s", fs.nResults, fs.nPaths, pcent, fs.duration)
	return result, nil
}
