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
	fileChangeEventRegistry     = make(map[string]*FileChangeEvent)
	fileChangeEventRegistryLock sync.Mutex
)

func RegisterFileChangeEvent(path string, call func()) {

	subscriber := &Subscriber{
		call: call,
	}

	fileChangeEventRegistryLock.Lock()
	event, ok := fileChangeEventRegistry[path]
	if ok {
		event.mu.Lock()
		event.Event.subscribers = append(event.Event.subscribers, subscriber)
		event.mu.Unlock()
		fileChangeEventRegistryLock.Unlock()
		return
	}

	event = &FileChangeEvent{
		Event: Event{
			ID:          utils.GenerateFileChangeEventId(),
			subscribers: []*Subscriber{subscriber},
		},
		Path: path,
	}
	fileChangeEventRegistry[path] = event
	fileChangeEventRegistryLock.Unlock()

	launchNewWatcher(path, event)
}

func launchNewWatcher(path string, fileChangeEvent *FileChangeEvent) {
	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		fileChangeEventRegistryLock.Lock()
		delete(fileChangeEventRegistry, path)
		fileChangeEventRegistryLock.Unlock()
		return
	}

	//add a path
	if err := watcher.Add(path); err != nil {
		fileChangeEventRegistryLock.Lock()
		delete(fileChangeEventRegistry, path)
		fileChangeEventRegistryLock.Unlock()
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

				// make a copy of subscribers
				// to avoid race condition
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
				//TODO: handle error
				println(err)
			}
		}
	}()

}

func GetFileChangeEventRegistry() map[string]*FileChangeEvent {
	fileChangeEventRegistryLock.Lock()
	defer fileChangeEventRegistryLock.Unlock()

	//copy registry
	fileChangeEventRegistryCopy := make(map[string]*FileChangeEvent)
	for k, v := range fileChangeEventRegistry {
		fileChangeEventRegistryCopy[k] = v
	}

	return fileChangeEventRegistryCopy
}
