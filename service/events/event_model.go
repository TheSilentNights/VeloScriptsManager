package events

type Event struct {
	ID          string
	subscribers []*Subscriber
}

type Subscriber struct {
	call func()
}
