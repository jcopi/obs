package fnchain

import (
	"sync"
)

var (
	engineMap sync.Map
)

func NewSingletonEngineFromEncoder[T encoder](enc T) *Engine[T] {
	key := any(*new(T))
	e, ok := engineMap.Load(key)

	if !ok {
		e = NewEngine[T]()
		engineMap.Store(key, e)
	}

	return e.(*Engine[T])
}

func NewDefaultSingletonEngine() *Engine[jsonEncoder] {
	return NewSingletonEngineFromEncoder(jsonEncoder{})
}
