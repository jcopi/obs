package fnchain

import (
	crand "crypto/rand"
	"io"
	"math/rand/v2"
	"sync"
	"time"
)

type Ctx struct {
	enc         encoder
	buf         *[]byte
	engine      *Engine
	eventParent uint64
	event       uint64
	typ         EvtType
	lvl         Level
}

type CtxMeta Ctx

var _ Event[*Ctx, *CtxMeta] = &Ctx{}

// Event implements Event.
func (c *Ctx) Event(lvl Level, typ EvtType) *Ctx {
	id, buf := c.engine.NewEvent()
	*buf = append(*buf, (*c.buf)...)
	return &Ctx{
		event:       id,
		eventParent: c.event,
		lvl:         lvl,
		typ:         typ,
		buf:         buf,
		enc:         c.enc,
		engine:      c.engine,
	}
}

// End implements Event.
func (c *Ctx) End() {
	c = c.With().Uint("type", uint(c.typ)).Str("lvl", LevelString(c.lvl)).Evt()
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

// Bool implements Event.
func (c *CtxMeta) Bool(key string, b bool) *CtxMeta {
	if c.enc.needsSeperator(*c.buf) {
		*c.buf = c.enc.appendSeperator(*c.buf)
	}
	*c.buf = c.enc.appendBool(*c.buf, key, b)
	return c
}

// Dur implements Event.
func (c *CtxMeta) Dur(key string, d time.Duration) *CtxMeta {
	if c.enc.needsSeperator(*c.buf) {
		*c.buf = c.enc.appendSeperator(*c.buf)
	}
	*c.buf = c.enc.appendDur(*c.buf, key, d)
	return c
}

// Err implements Event.
func (c *CtxMeta) Err(e error) *CtxMeta {
	if c.enc.needsSeperator(*c.buf) {
		*c.buf = c.enc.appendSeperator(*c.buf)
	}
	*c.buf = c.enc.appendErr(*c.buf, "error", e)
	return c
}

// Float32 implements Event.
func (c *CtxMeta) Float32(key string, f float32) *CtxMeta {
	if c.enc.needsSeperator(*c.buf) {
		*c.buf = c.enc.appendSeperator(*c.buf)
	}
	*c.buf = c.enc.appendFloat32(*c.buf, key, f)
	return c
}

// Float64 implements Event.
func (c *CtxMeta) Float64(key string, f float64) *CtxMeta {
	if c.enc.needsSeperator(*c.buf) {
		*c.buf = c.enc.appendSeperator(*c.buf)
	}
	*c.buf = c.enc.appendFloat64(*c.buf, key, f)
	return c
}

// Hex implements Event.
func (c *CtxMeta) Hex(key string, b []byte) *CtxMeta {
	if c.enc.needsSeperator(*c.buf) {
		*c.buf = c.enc.appendSeperator(*c.buf)
	}
	*c.buf = c.enc.appendHex(*c.buf, key, b)
	return c
}

// Int implements Event.
func (c *CtxMeta) Int(key string, i int) *CtxMeta {
	if c.enc.needsSeperator(*c.buf) {
		*c.buf = c.enc.appendSeperator(*c.buf)
	}
	*c.buf = c.enc.appendInt(*c.buf, key, i)
	return c
}

// Int64 implements Event.
func (c *CtxMeta) Int64(key string, i int64) *CtxMeta {
	if c.enc.needsSeperator(*c.buf) {
		*c.buf = c.enc.appendSeperator(*c.buf)
	}
	*c.buf = c.enc.appendInt64(*c.buf, key, i)
	return c
}

// Msg implements Event.
func (c *CtxMeta) Msg(msg string) *CtxMeta {
	if c.enc.needsSeperator(*c.buf) {
		*c.buf = c.enc.appendSeperator(*c.buf)
	}
	*c.buf = c.enc.appendStr(*c.buf, "msg", msg)
	return c
}

// Str implements Event.
func (c *CtxMeta) Str(key string, s string) *CtxMeta {
	if c.enc.needsSeperator(*c.buf) {
		*c.buf = c.enc.appendSeperator(*c.buf)
	}
	*c.buf = c.enc.appendStr(*c.buf, key, s)
	return c
}

// Strs implements Event.
func (c *CtxMeta) Strs(key string, strs []string) *CtxMeta {
	if c.enc.needsSeperator(*c.buf) {
		*c.buf = c.enc.appendSeperator(*c.buf)
	}
	*c.buf = c.enc.appendStrs(*c.buf, key, strs)
	return c
}

// Time implements Event.
func (c *CtxMeta) Time(key string, t time.Time) *CtxMeta {
	if c.enc.needsSeperator(*c.buf) {
		*c.buf = c.enc.appendSeperator(*c.buf)
	}
	*c.buf = c.enc.appendTime(*c.buf, key, t)
	return c
}

// Uint implements Event.
func (c *CtxMeta) Uint(key string, u uint) *CtxMeta {
	if c.enc.needsSeperator(*c.buf) {
		*c.buf = c.enc.appendSeperator(*c.buf)
	}
	*c.buf = c.enc.appendUint(*c.buf, key, u)
	return c
}

// Uint64 implements Event.
func (c *CtxMeta) Uint64(key string, u uint64) *CtxMeta {
	if c.enc.needsSeperator(*c.buf) {
		*c.buf = c.enc.appendSeperator(*c.buf)
	}
	*c.buf = c.enc.appendUint64(*c.buf, key, u)
	return c
}

type Engine struct {
	evtPool sync.Pool
	queue   chan *[]byte
	idsrc   *rand.Rand
}

func (e *Engine) NewEvent() (uint64, *[]byte) {
	b := e.evtPool.Get().(*[]byte)
	*b = (*b)[:0]
	return e.idsrc.Uint64(), b
}

func (e *Engine) QueueEvent(evt *[]byte) {
	// e.queue <- evt
	*evt = (*evt)[:0]
	e.evtPool.Put(evt)
}

func (e *Engine) ProcessEvents(w io.Writer) {
	for evt := range e.queue {
		n, err := w.Write(*evt)
		if err != nil && n == 0 {
			// The entire write failed, the log line can be retried
			e.queue <- evt
		} else if err != nil {
			// a short write occurred, maybe do something about the short-write
			// it's unclear what could be done in this case, this thread is not writing
		}

		*evt = (*evt)[:0]
		e.evtPool.Put(evt)
	}
}

func (e *Engine) Close() {
	close(e.queue)
}

const defaultEventCapacity = 64

func NewEngine() *Engine {
	chachaSeed := [32]byte{}
	crand.Read(chachaSeed[:])
	return &Engine{
		evtPool: sync.Pool{
			New: func() any {
				b := make([]byte, 0, defaultEventCapacity)
				return &b
			},
		},
		queue: make(chan *[]byte),
		idsrc: rand.New(rand.NewChaCha8(chachaSeed)),
	}
}
