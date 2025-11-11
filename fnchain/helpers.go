package fnchain

import (
	"context"
	"io"
	"os"
	"sync"
)

var (
	engineMap sync.Map
)

type contextKey struct{}

// NewSingletonEngine uses the writer as the singleton key
// this way there can only ever be a single engine using a particular writer instance
func NewSingletonEngine(w io.Writer) Engine {
	e, ok := engineMap.Load(w)

	if !ok {
		e = NewEngine(w)
		engineMap.Store(w, e)
	}

	return e.(Engine)
}

func NewDefaultSingletonEngine() Engine {
	return NewSingletonEngine(os.Stdout)
}

func FromContextOrRoot(ctx context.Context) *Ctx {
	val := ctx.Value(contextKey{})
	if val == nil {
		return nil
	}

	e, ok := val.(*Ctx)
	if !ok {
		return NewDefaultSingletonEngine().RootEvent(NoLevel, GenericEvent)
	}

	return e
}

func ContextWithEvent(ctx context.Context, e *Ctx) context.Context {
	return context.WithValue(ctx, contextKey{}, e.Event(NoLevel, GenericEvent))
}
