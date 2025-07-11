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

	tickFreq time.Duration
	tick     *time.Ticker
}

func newNotifyWatcher(config *Config, ch changeHandler) *notifyWatcher {
	return &notifyWatcher{
		changeHandler: ch,
		watched:       make(map[string]time.Time),
		excludes:      config.Excludes,
		tickFreq:      config.PollDuration,
		tick:          time.NewTicker(config.PollDuration),
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
			if !ok {
				return // Channel was closed (i.e. Watcher.Close() was called).
			}
			watcher.handleError(err)
		case e, ok := <-watcher.fs.Events:
			if !ok {
				return // Channel was closed (i.e. Watcher.Close() was called).
			}
			watcher.handleEvent(&e)
		}
	}
}

func (watcher *notifyWatcher) handleError(err error) {
	log.Printf("error: %v", err)
}

func (watcher *notifyWatcher) handleEvent(e *fsnotify.Event) {
	select {
	case <-watcher.tick.C:
		// reading from watcher.tick.C will tick at most once every watcher.tickFreq
		go watcher.handleFileChange(e.Name)

		// reset the ticker in case there is another event already queued
		// this ensures that we don't process events too frequently
		watcher.tick.Reset(watcher.tickFreq)
	default:
		// Skip processing if the event is too soon after the last one.
		return
	}
}
