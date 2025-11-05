package fnchain

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSingletonEngine(t *testing.T) {
	engineMap.Clear()
	e := NewSingletonEngine(os.Stdout)
	assert.NotNil(t, e)

	var check int
	engineMap.Range(func(key, value any) bool {
		check++
		return true
	})

	assert.Equal(t, 1, check)
	e.Close()

	_ = NewSingletonEngine(os.Stdout)
	var checkDup int
	engineMap.Range(func(key, value any) bool {
		checkDup++
		return true
	})
	assert.Equal(t, 1, checkDup)
}

func TestContextHelpers(t *testing.T) {
	ctx := context.Background()
	ev := FromContextOrNil(ctx)
	assert.Nil(t, ev)

	e := NewDefaultSingletonEngine()
	evt := e.RootEvent(DebugLevel, GenericEvent)

	weCtx := ContextWithEvent(ctx, evt)
	actual := FromContextOrNil(weCtx)
	assert.Equal(t, evt, actual)
}
