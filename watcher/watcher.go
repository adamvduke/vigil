package watcher

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type changeHandler interface {
	handleFileChange(string)
}

type watcher struct {
	changeHandler changeHandler
	watched       map[string]time.Time
}

func newWatcher(c changeHandler) *watcher {
	return &watcher{
		watched:       make(map[string]time.Time),
		changeHandler: c,
	}
}

func (watcher *watcher) startPolling(pollDuration time.Duration) error {
	go watcher.changeHandler.handleFileChange("")
	for {
		for _, path := range watcher.watchedPaths() {
			info, err := os.Stat(path)
			if err != nil {
				return err
			}
			watcher.checkModified(info, path)
		}
		time.Sleep(pollDuration)
	}
}

func (watcher *watcher) checkModified(info os.FileInfo, path string) {
	if !info.ModTime().In(time.UTC).Equal(watcher.watched[path]) {
		watcher.watchPath(path, info)
		if watcher.changeHandler != nil {
			go watcher.changeHandler.handleFileChange(path)
		}
	}
}

func (watcher *watcher) watch(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		if err := watcher.watchDir(path); err != nil {
			return nil, err
		}
	} else {
		watcher.watchPath(path, info)
	}

	return watcher.watchedPaths(), nil
}

func (watcher *watcher) watchedPaths() []string {
	paths := []string{}
	for p := range watcher.watched {
		paths = append(paths, p)
	}
	return paths
}

func (watcher *watcher) watchDir(root string) error {
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && !strings.Contains(path, ".git") {
			info, err := d.Info()
			if err != nil {
				return err
			}
			watcher.watchPath(path, info)
		}

		return nil
	})

	return err
}

func (watcher *watcher) watchPath(path string, info fs.FileInfo) {
	watcher.watched[path] = info.ModTime().In(time.UTC)
}
