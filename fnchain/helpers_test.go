package fnchain

import (
	"os"
	"testing"
)

func TestNewSingletonEngine(t *testing.T) {
	engineMap.Clear()
	e := NewSingletonEngine(os.Stdout)
	if e == nil {
		t.Error("engine was nil, initialization failed")
	}

	var check int
	engineMap.Range(func(key, value any) bool {
		check++
		return true
	})

	if check != 1 {
		t.Errorf("map had %d entries, it should only have a single entry in the cache", check)
	}
	e.Close()

	_ = NewSingletonEngine(os.Stdout)
	var checkDup int
	engineMap.Range(func(key, value any) bool {
		checkDup++
		return true
	})
	if checkDup != 1 {
		t.Errorf("map had %d entries, cache did not work", checkDup)
	}
}
