package watcher

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type pollingWatcher struct {
	changeHandler
	paths    map[string]time.Time
	ticker   *time.Ticker
	excludes []string
}

func newPollingWatcher(config *Config, ch changeHandler) *pollingWatcher {
	return &pollingWatcher{
		changeHandler: ch,
		ticker:        time.NewTicker(config.PollDuration),
		paths:         make(map[string]time.Time),
		excludes:      config.Excludes,
	}
}

func (watcher *pollingWatcher) start() error {
	log.Println("Starting polling watcher")
	go func() {
		for range watcher.ticker.C {
			for _, path := range watcher.watchedPaths() {
				info, err := os.Stat(path)
				if err != nil {
					log.Println(err)
				}
				watcher.checkModified(info, path)
			}
		}
	}()
	return nil
}

func (watcher *pollingWatcher) stop() error {
	if watcher.ticker == nil {
		watcher.ticker.Stop()
	}
	return nil
}

func (watcher *pollingWatcher) checkModified(info os.FileInfo, path string) {
	if !info.ModTime().In(time.UTC).Equal(watcher.paths[path]) {
		watcher.updateModTime(path, info)
		go watcher.handleFileChange(path)
	}
}

func (watcher *pollingWatcher) addPath(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		if err := watcher.watchDir(path); err != nil {
			return nil, err
		}
	} else {
		watcher.updateModTime(path, info)
	}
	return watcher.watchedPaths(), nil
}

func (watcher *pollingWatcher) watchDir(root string) error {
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		for _, exclude := range watcher.excludes {
			if strings.Contains(path, exclude) {
				return nil
			}
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			watcher.updateModTime(path, info)
		}
		return nil
	})
	return err
}

func (watcher *pollingWatcher) updateModTime(path string, info fs.FileInfo) {
	watcher.paths[path] = info.ModTime().In(time.UTC)
}

func (watcher *pollingWatcher) watchedPaths() []string {
	paths := []string{}
	for p := range watcher.paths {
		paths = append(paths, p)
	}
	return paths
}
