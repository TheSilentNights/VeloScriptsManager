package events

import "github.com/fsnotify/fsnotify"

type FileChangeEvent struct {
	Event
	watcher *fsnotify.Watcher
	Path    string
}

const FileChangeEventID = "file_change_event"

var registry = make(map[string]*FileChangeEvent)

func RegisterFileChangeEvent(path string, call func()) {
	subscriber := &Subscriber{
		call: call,
	}

	if _, ok := registry[path]; ok {
		registry[path].Event.subscribers = append(registry[path].Event.subscribers, *subscriber)
		return
	}

	event := &FileChangeEvent{
		Event: Event{
			subscribers: make([]Subscriber, 0),
		},
		Path: path,
	}

	launchNewWatcher(path, event)

	registry[path] = event
}

func launchNewWatcher(path string, fileChangeEvent *FileChangeEvent) {
	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		panic(err)
	}

	//add a path
	if err := watcher.Add(path); err != nil {
		panic(err)
	}

	go func() {
		for {
			select {
			case _, ok := <-watcher.Events:
				if !ok {
					return
				}

				for _, v := range fileChangeEvent.subscribers {
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
