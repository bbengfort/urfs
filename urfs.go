// Package urfs provides utilities for rapidly walking a file system
// directory using go concurrency in order to apply a function to discovered
// paths. The original application using this methodology was a uniform
// random sample of files in a directory, but this package has been expanded
// to include other utilities such as search and file size distribution.
package urfs

import (
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
)

// DefaultWorkers is the default number of worker threads to access system
// resources to prevent too many files open or max number of threads.
const DefaultWorkers = 5000

// DefaultBuffer is the size of the channels used to store paths and results.
const DefaultBuffer = 1000

//===========================================================================
// Initialization
//===========================================================================

// Initialize the package and random numbers, etc.
func init() {
	// Set the random seed to something different each time.
	rand.Seed(time.Now().Unix())
}

//===========================================================================
// File System Walker
//===========================================================================

// WalkFunc is a function that takes a path and does something with it,
// possibly returning an error which cancels the entire operation. It should
// also return a string, representing the path of anything being worked on,
// if the WalkFunc does nothing, it should return an empty string.
type WalkFunc func(path string) (string, error)

// FSWalker provides an API for walking a file system and applying a function
// concurrently to every path discovered. It is meant to handle much larger
// directories than ioutil.Walk. A set number of workers (by default 5000) is
// used to apply the function so that maximum files open or maximum thread
// limits are not reached, crashing the program.
type FSWalker struct {
	Workers    int             // number of workers that apply the func
	SkipHidden bool            // whether or not to skip hidden files and directories
	SkipDirs   bool            // whether or not to skip directories
	Match      string          // pattern to match files on (glob syntax)
	root       string          // root path currently being walked
	paths      chan string     // channel that discovered paths are passed to
	nPaths     uint64          // total number of paths discovered
	results    chan string     // paths that were operated on by the function
	nResults   uint64          // total number of results
	group      *errgroup.Group // group of threads being waited on
	ctx        context.Context // context of concurrent operation
	started    time.Time       // the time the last walk was started
	duration   time.Duration   // amount of time it took to walk and apply func
}

// Init the FSWalker and associated data structures.
func (fs *FSWalker) Init(ctx context.Context) {
	// Set up FSWalker defaults
	fs.Workers = DefaultWorkers
	fs.SkipHidden = true
	fs.SkipDirs = true
	fs.Match = "*"

	// Reset the required data structures
	fs.Reset(ctx)
}

// Reset the FSWalker and create required data structures.
func (fs *FSWalker) Reset(ctx context.Context) {
	if ctx == nil {
		// Create a new context
		ctx = context.Background()
		deadline, ok := fs.ctx.Deadline()
		if ok {
			ctx, _ = context.WithDeadline(ctx, deadline)
		}
	}

	fs.paths = make(chan string, DefaultBuffer)
	fs.results = make(chan string, DefaultBuffer)
	fs.group, fs.ctx = errgroup.WithContext(ctx)
	fs.nPaths = 0
	fs.nResults = 0
	fs.started = time.Time{}
	fs.duration = time.Duration(0)
}

// Walk the file systemfrom the path and apply the specified function.
// Can optionally pass a match pattern which uses glob-like syntax to match
// files and filter the paths being processed (if empty string is passed in,
// then the pattern is set to "*").
//
// NOTE: once walked, the FSWalker must be reinitialized to walk again.
func (fs *FSWalker) Walk(path string, walkFn WalkFunc) error {
	// Compute the duration of the walk
	fs.started = time.Now()
	defer func() { fs.duration = time.Since(fs.started) }()

	// Set the root path for the walk
	fs.root = path

	// Launch the goroutine that populates the paths
	fs.group.Go(fs.walk)

	// Create the worker function and allocate pool
	worker := fs.worker(walkFn)
	for w := 0; w < fs.Workers; w++ {
		fs.group.Go(worker)
	}

	// Wait for the workers to complete, then close the results channel
	go func() {
		fs.group.Wait()
		close(fs.results)
	}()

	// Start gathering the results
	for _ = range fs.results {
		fs.nResults++
	}

	return fs.group.Wait()
}

// Internal walk function that populates the paths channel.
func (fs *FSWalker) walk() error {
	// Ensure that the channel is closed when we've loaded all paths.
	defer close(fs.paths)

	// Walk through all the files in the directory specified, ignoring hidden
	// files and directories if required, matching the pattern if provided.
	return filepath.Walk(fs.root, fs.filterPaths)
}

// Internal filter paths function that is passed to filepath.Walk
func (fs *FSWalker) filterPaths(path string, info os.FileInfo, err error) error {
	// Propagate any errors
	if err != nil {
		return err
	}

	// Check to ensure that no mode bits are set
	if !info.Mode().IsRegular() {
		return nil
	}

	// Get the name of the file without the complete path
	name := info.Name()

	// Skip hidden files or directories if required.
	if fs.SkipHidden {
		if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "~") {
			return nil
		}
	}

	// Skip directories if required
	if fs.SkipDirs {
		if info.IsDir() {
			return nil
		}
	}

	// Check to see if the pattern matches the file
	match, err := filepath.Match(fs.Match, name)
	if err != nil {
		return err
	} else if !match {
		return nil
	}

	// Increment the total number of paths we've seen.
	atomic.AddUint64(&fs.nPaths, 1)

	select {
	case fs.paths <- path:
	case <-fs.ctx.Done():
		return fs.ctx.Err()
	}

	return nil
}

// Internal helper function that creates a worker function for the specified
// WalkFunc action to be applied to each path.
func (fs *FSWalker) worker(walkFn WalkFunc) func() error {
	return func() error {
		// Apply the function all paths in the channel
		for path := range fs.paths {
			// avoid race condition
			p := path

			// apply the walk function to the path and return errors
			r, err := walkFn(p)
			if err != nil {
				return err
			}

			// store the result and check the context
			if r != "" {

				select {
				case fs.results <- r:
				case <-fs.ctx.Done():
					return fs.ctx.Err()
				}
			}

		}
		return nil
	}
}
