package trace

type Tracer interface {
	ReceivedFromChannel(name string)
	SendingToChannel(name string)
	SentToChannel(name string)
	Event(message string, args ...any)
}
