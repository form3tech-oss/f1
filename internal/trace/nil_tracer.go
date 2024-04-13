package trace

var _ Tracer = &NilTracer{}

func NewNilTracer() *NilTracer {
	return &NilTracer{}
}

type NilTracer struct{}

func (*NilTracer) ReceivedFromChannel(string) {
}

func (*NilTracer) SentToChannel(string) {
}

func (*NilTracer) SendingToChannel(string) {
}

func (*NilTracer) Event(string) {
}

func (*NilTracer) WorkerEvent(string, string) {
}

func (*NilTracer) IterationEvent(string, uint32) {
}
