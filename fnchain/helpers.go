package fnchain

import (
	"context"
	"sync"
)

var (
	engineMap  sync.Map
	contextKey = "obs_context_key"
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

// func FromContextOrRoot[T any, M EventMeta[M, *T]](ctx context.Context, lvl Level, typ EvtType) Event[T, M] {
// 	evt := FromContextOrNil[T, M](ctx)
// 	if evt != nil {
// 		return evt
// 	}

// 	return NewDefaultSingletonEngine().RootEvent(lvl, typ)
// }

func FromContextOrNil[T any, M EventMeta[M, *T]](ctx context.Context) Event[T, M] {
	val := ctx.Value(contextKey)
	if val == nil {
		return nil
	}

	e, ok := val.(Event[T, M])
	if !ok {
		return nil
	}
	return e
}

func ContextWithEvent[T any, M EventMeta[M, *T]](ctx context.Context, e Event[T, M]) context.Context {
	return context.WithValue(ctx, contextKey, e)
}
