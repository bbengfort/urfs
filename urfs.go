// Package urfs provides utilities for rapidly walking a file system
// directory using go concurrency in order to apply a function to discovered
// paths. The original application using this methodology was a uniform
// random sample of files in a directory, but this package has been expanded
// to include other utilities such as search and file size distribution.
package urfs

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
)

//===========================================================================
// Initialization
//===========================================================================

// Initialize the package and random numbers, etc.
func init() {
	// Set the random seed to something different each time.
	rand.Seed(time.Now().Unix())
}

// Sample the files contained in a source directory (src), copying them to a
// destination directory (dst) with some probability between 0 and 1 (size).
func Sample(ctx context.Context, src, dst string, size float64) error {
	// Compute the started time
	started := time.Now()

	// Create the dst directory if it doesn't exist
	if err := Mkdir(dst); err != nil {
		return err
	}

	// Create the wait group with the context
	g, ctx := errgroup.WithContext(ctx)

	// Create a buffered channel to collect paths on
	paths := make(chan string, 1000)
	results := make(chan string, 1000)
	var total uint64

	// Launch the group of goroutines
	g.Go(func() error {
		// Ensure the channel is closed when we've loaded all the paths.
		defer close(paths)

		// Walk through all the files in the directory specified, ignoring
		// hidden directories and folders and processing any discovered files.
		return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
			// Propagate any errors
			if err != nil {
				return err
			}

			// Check to ensure that no mode bits are set
			if !info.Mode().IsRegular() {
				return nil
			}

			// Skip any hidden files or directories
			if strings.HasPrefix(info.Name(), ".") || strings.HasPrefix(info.Name(), "~") {
				return nil
			}

			// Increment the total
			atomic.AddUint64(&total, 1)

			select {
			case paths <- path:
			case <-ctx.Done():
				return ctx.Err()
			}

			return nil
		})
	})

	// Create a worker function
	worker := func() error {
		for path := range paths {
			p := path
			if rand.Float64() <= size {
				// Get the relative path from the base
				rel, err := filepath.Rel(src, p)
				if err != nil {
					return err
				}

				// Create the new path to the destination
				drl := filepath.Join(dst, rel)

				// Create the directory if it doesn't exist
				if err = Mkdir(filepath.Dir(drl)); err != nil {
					return err
				}

				// Copy the file to the destination directory
				if err = CopyFile(drl, p, 0644); err != nil {
					return err
				}

				select {
				case results <- drl:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}
		return nil
	}

	// Allocate 10000 workers
	for w := 0; w < 10000; w++ {
		g.Go(worker)
	}

	// Wait for the group to complete then close the results channel
	go func() {
		g.Wait()
		close(results)
	}()

	// Gather the results
	sampled := 0
	for _ = range results {
		sampled++
	}

	pcent := float64(sampled) / float64(total) * 100.0
	fmt.Printf("sampled %d out of %d files (%0.1f%%) in %s\n", sampled, total, pcent, time.Since(started))
	return g.Wait()
}
