package fnchain

type Event[T any, M EventMeta[M, T]] interface {
	With() M

	// Create an event with the current as the parent
	Event(lvl Level, typ EvtType) T

	// Some functions to complete the event
	End()
}

type EventMeta[T any, E any] interface {
	Metadata[T]
	Evt() E
}
