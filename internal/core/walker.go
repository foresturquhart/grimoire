package core

import (
	"github.com/boyter/gocodewalker"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// Walker defines the behavior for walking files in a directory.
// Implementations return file paths via Files() and send errors to Errors().
type Walker interface {
	// Start initiates the file walking process.
	// If the walker cannot start (e.g., invalid directory), it returns an error.
	Start() error

	// Files returns a channel that emits file paths discovered during the walk.
	Files() <-chan string

	// Errors returns a channel that emits errors encountered during the walk.
	Errors() <-chan error
}

// DefaultWalker is an implementation of Walker that uses the gocodewalker library.
// It wraps gocodewalker.FileWalker and provides filtering for ignored paths.
type DefaultWalker struct {
	walker *gocodewalker.FileWalker

	// filesChan emits the final list of file paths that pass filtering.
	filesChan chan string

	// errorsChan emits errors encountered during the walk.
	errorsChan chan error

	// fileListQueue is the channel that gocodewalker writes *gocodewalker.File into.
	fileListQueue chan *gocodewalker.File

	// targetDir is the base directory from which we walk.
	targetDir string

	// outputFile is the output file which should be excluded.
	outputFile string

	// ignoredPathRegexes is a list of regex patterns representing ignored paths.
	ignoredPathRegexes []*regexp.Regexp

	// wg ensures our bridging goroutine completes before we close channels.
	wg sync.WaitGroup
}

// NewDefaultWalker configures and returns a new DefaultWalker.
func NewDefaultWalker(targetDir string, allowedFileExtensions []string, ignoredPathRegexes []*regexp.Regexp, outputFile string) *DefaultWalker {
	fileListQueue := make(chan *gocodewalker.File, 100)

	w := gocodewalker.NewFileWalker(targetDir, fileListQueue)
	w.AllowListExtensions = allowedFileExtensions
	w.CustomIgnore = []string{".grimoireignore"}

	return &DefaultWalker{
		walker:             w,
		filesChan:          make(chan string),
		errorsChan:         make(chan error),
		fileListQueue:      fileListQueue,
		targetDir:          targetDir,
		outputFile:         outputFile,
		ignoredPathRegexes: ignoredPathRegexes,
	}
}

// Start begins the file walking process.
func (dw *DefaultWalker) Start() error {
	// Set an error handler so gocodewalker sends errors to our channel.
	dw.walker.SetErrorHandler(func(err error) bool {
		dw.errorsChan <- err

		// Continue walking despite the error.
		return true
	})

	// Launch a goroutine to bridge *gocodewalker.File to string paths.
	dw.wg.Add(1)
	go func() {
		defer dw.wg.Done()

		// Close filesChan when we're done draining fileListQueue.
		defer close(dw.filesChan)

		for fileEntry := range dw.fileListQueue {
			// Exclude the current output file, if set.
			if dw.outputFile != "" && dw.outputFile == fileEntry.Location {
				continue
			}

			// Compute a relative path for filtering.
			relPath := strings.TrimPrefix(fileEntry.Location, dw.targetDir)
			relPath = strings.TrimPrefix(relPath, string(filepath.Separator))

			// Only emit file paths that are not ignored.
			if shouldInclude(relPath, dw.ignoredPathRegexes) {
				dw.filesChan <- relPath
			}
		}
	}()

	if err := dw.walker.Start(); err != nil {
		return err
	}

	return nil
}

// Files returns a read-only channel that emits file paths as strings.
func (dw *DefaultWalker) Files() <-chan string {
	return dw.filesChan
}

// Errors returns a read-only channel that emits errors encountered during the walk.
func (dw *DefaultWalker) Errors() <-chan error {
	return dw.errorsChan
}

// Close waits for the bridging goroutine to finish and then closes errorsChan.
// Consumers can call this after they've drained Files() to fully clean up.
func (dw *DefaultWalker) Close() {
	dw.wg.Wait()
	close(dw.errorsChan)
}

// shouldInclude checks if relPath is NOT matched by any of the ignored regex patterns.
// If none match, the path should be included in the results.
func shouldInclude(relPath string, ignoredRegexes []*regexp.Regexp) bool {
	for _, re := range ignoredRegexes {
		if re.MatchString(relPath) {
			return false
		}
	}
	return true
}
