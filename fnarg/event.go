package fnarg

type Event[T any] interface {
	// Create a child event
	Event(lvl Level, typ EvtType) T
	With(meta ...Metadata) T
	End()
}
