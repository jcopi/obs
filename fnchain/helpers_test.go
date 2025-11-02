package fnchain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSingletonEngineFromEncoder(t *testing.T) {
	enc := jsonEncoder{}
	engineMap.Clear()
	e := NewSingletonEngineFromEncoder(enc)
	assert.NotNil(t, e)

	var check int
	engineMap.Range(func(key, value any) bool {
		check++
		return true
	})

	assert.Equal(t, 1, check)
	e.Close()

	_ = NewSingletonEngineFromEncoder(enc)
	var checkDup int
	engineMap.Range(func(key, value any) bool {
		checkDup++
		return true
	})
	assert.Equal(t, 1, checkDup)
}

// func TestContextHelpers(t *testing.T) {
// 	ctx := context.Background()
// 	ev := FromContextOrNil[Ctx[jsonEncoder], *CtxMeta[jsonEncoder]](ctx)
// 	assert.Nil(t, ev)

// 	e := NewDefaultSingletonEngine()
// 	evt := e.RootEvent(DebugLevel, GenericEvent)
// 	ContextWithEvent(ctx, evt)
// }
