package events

type Event struct {
	subscribers []Subscriber
}

type Subscriber struct {
	call func()
}
