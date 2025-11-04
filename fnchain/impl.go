package fnchain

import (
	"time"
)

const rootEventParent uint64 = 0

type Ctx struct {
	engine      Engine
	buf         []byte
	eventParent uint64
	eventID     uint64
	typ         EvtType
	lvl         Level
}

type CtxMeta Ctx

var _ Event[*Ctx, *CtxMeta] = &Ctx{}

// Event implements Event.
func (c *Ctx) Event(lvl Level, typ EvtType) *Ctx {
	id, buf := c.engine.initEvent(len(c.buf))

	// There is a need to
	next := Ctx{
		eventID:     id,
		eventParent: c.eventID,
		lvl:         lvl,
		typ:         typ,
		engine:      c.engine,
	}

	if len(c.buf) > 0 {
		next.buf = append(buf, c.buf...)
	} else {
		next.buf = (jsonEncoder{}).appendStart(buf)
	}

	return &next
}

// End implements Event.
func (c *Ctx) End() {

	c.buf = (jsonEncoder{}).appendKnownKeyID(c.buf, "@id", c.eventID)
	c.buf = (jsonEncoder{}).appendKnownKeyID(c.buf, "@pid", c.eventParent)
	c.buf = (jsonEncoder{}).appendKnownKeyType(c.buf, "@type", c.typ)
	c.buf = (jsonEncoder{}).appendKnownKeyLevel(c.buf, "level", c.lvl)
	c.buf = (jsonEncoder{}).appendEnd(c.buf)

	c.engine.QueueEvent(c.buf)
	c.buf = nil
}

func (c *Ctx) With() *CtxMeta {
	return (*CtxMeta)(c)
}

// Evt implements EventMeta
func (c *CtxMeta) Evt() *Ctx {
	return (*Ctx)(c)
}

// Base64 implements Event.
func (c *CtxMeta) Base64(key string, b []byte) *CtxMeta {
	c.buf = (jsonEncoder{}).appendB64(c.buf, key, b)
	return c
}

// Bool implements Event.
func (c *CtxMeta) Bool(key string, b bool) *CtxMeta {

	c.buf = (jsonEncoder{}).appendBool(c.buf, key, b)
	return c
}

// Dur implements Event.
func (c *CtxMeta) Dur(key string, d time.Duration) *CtxMeta {

	c.buf = (jsonEncoder{}).appendDur(c.buf, key, d)
	return c
}

// Err implements Event.
func (c *CtxMeta) Err(e error) *CtxMeta {

	c.buf = (jsonEncoder{}).appendErr(c.buf, "error", e)
	return c
}

// Float32 implements Event.
func (c *CtxMeta) Float32(key string, f float32) *CtxMeta {

	c.buf = (jsonEncoder{}).appendFloat32(c.buf, key, f)
	return c
}

// Float64 implements Event.
func (c *CtxMeta) Float64(key string, f float64) *CtxMeta {

	c.buf = (jsonEncoder{}).appendFloat64(c.buf, key, f)
	return c
}

// Hex implements Event.
func (c *CtxMeta) Hex(key string, b []byte) *CtxMeta {

	c.buf = (jsonEncoder{}).appendHex(c.buf, key, b)
	return c
}

// Int implements Event.
func (c *CtxMeta) Int(key string, i int) *CtxMeta {

	c.buf = (jsonEncoder{}).appendInt(c.buf, key, i)
	return c
}

// Int64 implements Event.
func (c *CtxMeta) Int64(key string, i int64) *CtxMeta {

	c.buf = (jsonEncoder{}).appendInt64(c.buf, key, i)
	return c
}

// Msg implements Event.
func (c *CtxMeta) Msg(msg string) *CtxMeta {

	c.buf = (jsonEncoder{}).appendStr(c.buf, "msg", msg)
	return c
}

// Str implements Event.
func (c *CtxMeta) Str(key string, s string) *CtxMeta {

	c.buf = (jsonEncoder{}).appendStr(c.buf, key, s)
	return c
}

// Strs implements Event.
func (c *CtxMeta) Strs(key string, strs []string) *CtxMeta {

	c.buf = (jsonEncoder{}).appendStrs(c.buf, key, strs)
	return c
}

// Time implements Event.
func (c *CtxMeta) Time(key string, t time.Time) *CtxMeta {

	c.buf = (jsonEncoder{}).appendTime(c.buf, key, t)
	return c
}

// Uint implements Event.
func (c *CtxMeta) Uint(key string, u uint) *CtxMeta {

	c.buf = (jsonEncoder{}).appendUint(c.buf, key, u)
	return c
}

// Uint64 implements Event.
func (c *CtxMeta) Uint64(key string, u uint64) *CtxMeta {

	c.buf = (jsonEncoder{}).appendUint64(c.buf, key, u)
	return c
}
