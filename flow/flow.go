package flow

// An execution flow
type Flow[T any, Te Event[Te]] interface {
	Metadata[T]
	
	// Start a sub flow that is annotated to have come from this parent flow
	Flow() T
	// Create an event that is part of the flow
	Event() Te
}

type Event[T any] interface {
	Metadata[T]
	
	// Create an event with the current as the parent
	Event() T

	// Some functions to complete the event
	End()
}
