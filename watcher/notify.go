package watcher

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type notifyWatcher struct {
	changeHandler
	watched  map[string]time.Time
	fs       *fsnotify.Watcher
	excludes []string
}

func newNotifyWatcher(config *Config, ch changeHandler) *notifyWatcher {
	return &notifyWatcher{
		changeHandler: ch,
		watched:       make(map[string]time.Time),
		excludes:      config.Excludes,
	}
}

func (watcher *notifyWatcher) start() error {
	log.Println("Starting notify watcher")
	if watcher.fs != nil {
		return watcher.fs.Close()
	}
	fs, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	watcher.fs = fs
	go watcher.monitorLoop()

	return nil
}

func (watcher *notifyWatcher) stop() {
	watcher.fs.Close()
}

func (watcher *notifyWatcher) addPath(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		err = filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			return watcher.watchDir(path)
		})
		if err != nil {
			return nil, err
		}
	} else {
		watcher.fs.Add(path)
	}

	return watcher.watchedPaths(), nil
}

func (watcher *notifyWatcher) watchDir(path string) error {
	for _, exclude := range watcher.excludes {
		if strings.Contains(path, exclude) {
			return nil
		}
	}
	return watcher.fs.Add(path)
}

func (watcher *notifyWatcher) watchedPaths() []string {
	return watcher.fs.WatchList()
}

func (watcher *notifyWatcher) monitorLoop() {
	for {
		select {
		case err, ok := <-watcher.fs.Errors:
			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				return
			}
			log.Printf("error: %v", err)
			continue
		case e, ok := <-watcher.fs.Events:
			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				return
			}
			if !e.Has(fsnotify.Chmod) {
				go watcher.handleFileChange(e.Name)
			}
		}
	}
}
