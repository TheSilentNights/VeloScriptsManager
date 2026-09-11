package events

type Event struct {
	ID          string
	subscribers []*Subscriber
	status      string
}

type Subscriber struct {
	call func()
}
