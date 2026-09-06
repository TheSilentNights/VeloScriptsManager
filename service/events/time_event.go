package events

import (
	"github/TheSilentNights/VeloScriptsManager/service/utils"
	"sync"
	"time"
)

type TimeEvent struct {
	Event
	timer *time.Timer
}

const TimeEventID = "time_event"

var (
	timeEventRegistry     = make(map[string]*TimeEvent)
	timeEventRegistryLock sync.Mutex
)

func registerTimeEvent(afterSeconds int, call func()) {
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

	go func(timeEvent *TimeEvent) {
		<-timer.C
		for _, subscriber := range timeEvent.Event.subscribers {
			subscriber.call()
		}
		timeEvent.timer.Stop()
	}(timeEvent)

}
