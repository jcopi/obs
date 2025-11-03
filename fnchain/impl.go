package fnchain

import (
	"time"
)

const rootEventParent uint64 = 0

type Ctx[E encoder] struct {
	engine      Engine[E]
	buf         []byte
	eventParent uint64
	eventID     uint64
	typ         EvtType
	lvl         Level
}

type CtxMeta[E encoder] Ctx[E]

var _ Event[Ctx[jsonEncoder], *CtxMeta[jsonEncoder]] = &Ctx[jsonEncoder]{}

// Event implements Event.
func (c *Ctx[T]) Event(lvl Level, typ EvtType) Ctx[T] {
	id, buf := c.engine.initEvent(len(c.buf))

	// There is a need to
	next := Ctx[T]{
		eventID:     id,
		eventParent: c.eventID,
		lvl:         lvl,
		typ:         typ,
		engine:      c.engine,
	}

	if len(c.buf) > 0 {
		next.buf = append(buf, c.buf...)
	} else {
		next.buf = append(buf, '{')
	}

	return next
}

// End implements Event.
func (c *Ctx[T]) End() {
	var enc T

	c.buf = enc.appendID(c.buf, "@id", c.eventID)
	c.buf = enc.appendID(c.buf, "@pid", c.eventParent)
	c.buf = enc.appendUint(c.buf, "@type", uint(c.typ))
	c.buf = enc.appendStr(c.buf, "level", LevelString(c.lvl))
	c.buf = enc.appendEnd(c.buf)

	c.engine.QueueEvent(c.buf)
	c.buf = nil
}

func (c *Ctx[T]) With() *CtxMeta[T] {
	return (*CtxMeta[T])(c)
}

// Evt implements EventMeta
func (c *CtxMeta[T]) Evt() *Ctx[T] {
	return (*Ctx[T])(c)
}

// Bool implements Event.
func (c *CtxMeta[T]) Bool(key string, b bool) *CtxMeta[T] {
	var enc T

	c.buf = enc.appendBool(c.buf, key, b)
	return c
}

// Dur implements Event.
func (c *CtxMeta[T]) Dur(key string, d time.Duration) *CtxMeta[T] {
	var enc T

	c.buf = enc.appendDur(c.buf, key, d)
	return c
}

// Err implements Event.
func (c *CtxMeta[T]) Err(e error) *CtxMeta[T] {
	var enc T

	c.buf = enc.appendErr(c.buf, "error", e)
	return c
}

// Float32 implements Event.
func (c *CtxMeta[T]) Float32(key string, f float32) *CtxMeta[T] {
	var enc T

	c.buf = enc.appendFloat32(c.buf, key, f)
	return c
}

// Float64 implements Event.
func (c *CtxMeta[T]) Float64(key string, f float64) *CtxMeta[T] {
	var enc T

	c.buf = enc.appendFloat64(c.buf, key, f)
	return c
}

// Hex implements Event.
func (c *CtxMeta[T]) Hex(key string, b []byte) *CtxMeta[T] {
	var enc T

	c.buf = enc.appendHex(c.buf, key, b)
	return c
}

// Int implements Event.
func (c *CtxMeta[T]) Int(key string, i int) *CtxMeta[T] {
	var enc T

	c.buf = enc.appendInt(c.buf, key, i)
	return c
}

// Int64 implements Event.
func (c *CtxMeta[T]) Int64(key string, i int64) *CtxMeta[T] {
	var enc T

	c.buf = enc.appendInt64(c.buf, key, i)
	return c
}

// Msg implements Event.
func (c *CtxMeta[T]) Msg(msg string) *CtxMeta[T] {
	var enc T

	c.buf = enc.appendStr(c.buf, "msg", msg)
	return c
}

// Str implements Event.
func (c *CtxMeta[T]) Str(key string, s string) *CtxMeta[T] {
	var enc T

	c.buf = enc.appendStr(c.buf, key, s)
	return c
}

// Strs implements Event.
func (c *CtxMeta[T]) Strs(key string, strs []string) *CtxMeta[T] {
	var enc T

	c.buf = enc.appendStrs(c.buf, key, strs)
	return c
}

// Time implements Event.
func (c *CtxMeta[T]) Time(key string, t time.Time) *CtxMeta[T] {
	var enc T

	c.buf = enc.appendTime(c.buf, key, t)
	return c
}

// Uint implements Event.
func (c *CtxMeta[T]) Uint(key string, u uint) *CtxMeta[T] {
	var enc T

	c.buf = enc.appendUint(c.buf, key, u)
	return c
}

// Uint64 implements Event.
func (c *CtxMeta[T]) Uint64(key string, u uint64) *CtxMeta[T] {
	var enc T

	c.buf = enc.appendUint64(c.buf, key, u)
	return c
}
