package events

import (
	"github/TheSilentNights/VeloScriptsManager/service/utils"
	"sync"
	"time"
)

type TimeEvent struct {
	Event
	Interval int
	Repeat   bool
	timer    *time.Timer
}

const TimeEventID = "time_event"

var (
	timeEventRegistry     = make(map[string]*TimeEvent)
	timeEventRegistryLock sync.Mutex
)

func RegisterTimeEvent(afterSeconds int, repeat bool, call func()) {
	subscriber := &Subscriber{
		call: call,
	}

	timeEventRegistryLock.Lock()
	defer timeEventRegistryLock.Unlock()

	timer := time.NewTimer(time.Duration(afterSeconds) * time.Second)

	timeEvent := &TimeEvent{
		Event: Event{
			ID:          utils.GenerateTimeEventId(),
			subscribers: []*Subscriber{subscriber},
		},
		timer: timer,
	}

	timeEventRegistry[timeEvent.ID] = timeEvent

	go waitForTime(timeEvent)
}

func waitForTime(timeEvent *TimeEvent) {
	<-timeEvent.timer.C
	for _, subscriber := range timeEvent.Event.subscribers {
		subscriber.call()
	}
	if timeEvent.Repeat {
		timeEvent.timer.Reset(time.Duration(timeEvent.Interval) * time.Second)
	} else {
		timeEvent.timer.Stop()
	}
}

func GetTimeEventRegistry() map[string]*TimeEvent {
	timeEventRegistryLock.Lock()
	defer timeEventRegistryLock.Unlock()

	//copy registry
	timeEventRegistryCopy := make(map[string]*TimeEvent)
	for k, v := range timeEventRegistry {
		timeEventRegistryCopy[k] = v
	}

	return timeEventRegistryCopy
}
