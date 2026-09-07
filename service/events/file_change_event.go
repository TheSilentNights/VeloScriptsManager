package events

import (
	"github/TheSilentNights/VeloScriptsManager/service/utils"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type FileChangeEvent struct {
	Event
	watcher *fsnotify.Watcher
	Path    string
	mu      sync.Mutex
}

const FileChangeEventID = "file_change_event"

// map by path
var (
	registry   = make(map[string]*FileChangeEvent)
	registryMu sync.Mutex
)

func RegisterFileChangeEvent(path string, call func()) {

	subscriber := &Subscriber{
		call: call,
	}

	registryMu.Lock()
	event, ok := registry[path]
	if ok {
		event.mu.Lock()
		event.Event.subscribers = append(event.Event.subscribers, subscriber)
		event.mu.Unlock()
		registryMu.Unlock()
		return
	}

	event = &FileChangeEvent{
		Event: Event{
			ID:          utils.GenerateFileChangeEventId(),
			subscribers: []*Subscriber{subscriber},
		},
		Path: path,
	}
	registry[path] = event
	registryMu.Unlock()

	launchNewWatcher(path, event)
}

func launchNewWatcher(path string, fileChangeEvent *FileChangeEvent) {
	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		registryMu.Lock()
		delete(registry, path)
		registryMu.Unlock()
		return
	}

	//add a path
	if err := watcher.Add(path); err != nil {
		watcher.Close()
		return
	}

	go func() {
		for {
			select {
			case _, ok := <-watcher.Events:
				if !ok {
					return
				}

				fileChangeEvent.mu.Lock()
				subscribers := make([]*Subscriber, len(fileChangeEvent.subscribers))
				copy(subscribers, fileChangeEvent.subscribers)
				fileChangeEvent.mu.Unlock()

				for _, v := range subscribers {
					v.call()
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				//todo: handle error
				println(err)
			}
		}
	}()

}
