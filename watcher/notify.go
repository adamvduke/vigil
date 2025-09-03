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
	runChan  chan *fsnotify.Event
}

func newNotifyWatcher(config *Config, ch changeHandler) *notifyWatcher {
	return &notifyWatcher{
		changeHandler: ch,
		watched:       make(map[string]time.Time),
		excludes:      config.Excludes,
		tickFreq:      config.PollDuration,
		tick:          time.NewTicker(config.PollDuration),
		runChan:       make(chan *fsnotify.Event, 1),
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
	go watcher.watchLoop()
	go watcher.runLoop()

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

func (watcher *notifyWatcher) watchLoop() {
	for {
		select {
		case err, ok := <-watcher.fs.Errors:
			if !ok {
				return // Channel was closed (i.e. Watcher.Close() was called).
			}
			go watcher.handleError(err)
		case e, ok := <-watcher.fs.Events:
			if !ok {
				return // Channel was closed (i.e. Watcher.Close() was called).
			}
			go func() {
				watcher.runChan <- &e
			}()
		}
	}
}

func (watcher *notifyWatcher) handleError(err error) {
	log.Printf("error: %v", err)
}

func (watcher *notifyWatcher) runLoop() {
	for {
		events := []*fsnotify.Event{}

		// Drain the channel and collect all events
		drain := true
		for drain {
			select {
			case e := <-watcher.runChan:
				if e.Has(fsnotify.Write) {
					// Ignore chmod events
					events = append(events, e)
				}
				// Collect the event
			default:
				drain = false
			}
		}

		if len(events) > 0 {
			// Print all events
			for _, e := range events {
				log.Printf("Event: %s %s", e.Name, e.Op)
			}

			// Pick a random event
			randIdx := time.Now().UnixNano() % int64(len(events))
			go watcher.handleFileChange(events[randIdx].Name)
		}

		// Wait for next tick
		<-watcher.tick.C
	}
}
