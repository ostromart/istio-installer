package fswatch

import (
	"log"
	"os"
	"path/filepath"

	"time"

	"github.com/fsnotify/fsnotify"
)

const (
	// debounceDelay is the amount of time that must elapse after a file changes without any other file changes before
	// a signal is sent.
	debounceDelay = time.Second
)

// WatchDirRecursively recursively watches the given directory path. If any file under dir changes, a debounced signal
// is sent on the returned channel. The debounce delay prevents flooding when multiple files are updated at once e.g.
// during upgrade.
func WatchDirRecursively(dir string) (<-chan struct{}, error) {
	subdirs, err := recursiveSubdirs(dir)
	if err != nil {
		return nil, err
	}

	// Raw notifies that must be debounced.
	notifyRaw := make(chan struct{})
	for _, d := range subdirs {
		watchDir(d, notifyRaw)
	}

	// Debounced notifies.
	notify := make(chan struct{})
	go func() {
		var timer *time.Timer
		for {
			select {
			case <-notifyRaw:
				if timer == nil {
					timer = time.NewTimer(debounceDelay)
					break
				}
				timer.Reset(debounceDelay)
			case <-timer.C:
				notify <- struct{}{}
			}
		}
	}()

	return notify, nil
}

func watchDir(dir string, notify chan<- struct{}) error {
	var err error
	doneInit := make(chan struct{}, 1)
	go func() {
		for {
			// All errors are likely to be same reason here, just pick one.
			watcher, err := fsnotify.NewWatcher()
			if err != nil {
				log.Print(err)
				return
			}
			defer watcher.Close()

			done := make(chan struct{}, 1)
			go func() {
				for {
					select {
					case event, ok := <-watcher.Events:
						if !ok {
							log.Printf("watcher channel closed for %s", dir)
							done <- struct{}{}
							return
						}
						if event.Op&fsnotify.Write == fsnotify.Write {
							log.Println("modified file:", event.Name)
							select {
							case notify <- struct{}{}:
							default:
							}
						}
					case err, ok := <-watcher.Errors:
						if !ok {
							log.Printf("watcher channel closed for %s", dir)
							done <- struct{}{}
							return
						}
						log.Printf("error for watcher on %s:", dir, err)
					}
				}
			}()

			err = watcher.Add(dir)
			if err != nil {
				log.Print(err)
				return
			}
			// Don't block the second time around.
			select {
			case doneInit <- struct{}{}:
			default:
			}
			// Wait in case we need to restart the watcher.
			<-done
		}
	}()
	// Wait for any init type of errors.
	<-doneInit
	return err
}

func recursiveSubdirs(dir string) ([]string, error) {
	var dirs []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			dirs = append(dirs, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return dirs, nil
}
