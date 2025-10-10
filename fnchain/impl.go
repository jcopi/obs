package fnchain

import (
	"bufio"
	crand "crypto/rand"
	"encoding/binary"
	"io"
	"math/rand/v2"
	"os"
	"sync"
	"time"
	"unsafe"
)

const maxEventBufferSize = 4 * 1024 * 1024 // 4 MiB
// sync.Pool requires that each member has roughly the same memory cost in-order to operate efficiently
// if there is any single member that is significantly larger than the mean it should not be added to the pool and
// allowed to be GC'd immediately.
const minEventBufferSize = 64 // 64 B.
// There is little value in adding a buffer smaller than this to the pool.
// It's likely to require allocations to increase the size so it's better to just allocate a larger buffer up front
const defaultEventBufferSize = 128 // 128 B
const eventQueueSize = 64
const rootEventParent uint64 = 0

type Ctx[E encoder] struct {
	engine      *Engine[E]
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

type Engine[T encoder] struct {
	evtPool   sync.Pool
	queue     chan []byte
	idsrc     *rand.Rand
	maxBuffer int
}

func (e *Engine[T]) RootEvent(lvl Level, typ EvtType) Ctx[T] {
	id, buf := e.initEvent(defaultEventBufferSize)
	return Ctx[T]{
		engine:      e,
		buf:         buf,
		eventParent: rootEventParent,
		eventID:     id,
		typ:         typ,
		lvl:         lvl,
	}
}

func (e *Engine[T]) initEvent(knownBufferSize int) (uint64, []byte) {
	b := e.getBuffer(max(knownBufferSize, defaultEventBufferSize))
	return e.idsrc.Uint64(), b[:0]
}

// TODO: QueueEvent and ProcessEvents needs some testing. The assumption was that writing to a
// buffered channel that actually did the syscall writes would produce *consistent* (and minimal)
// latency when writing events `End()`. This assumption needs to be tested.
// Some benchmarks exist, but they're not really representative, when there are event buffers being
// written continuously (particularly small events) it's easy to saturate the throughput of the `write`
// call and so the `End()` calls end up spending a lot of time waiting on channel locks. When saturating the
// underlying writer directly writing should always be fastest, but shouldn't be indicative of real-world usage.
//
// Another issue is that benchmarks can easily show the average/amortized cost, but we also need to ensure that
// we don't introduce large latencies to a small number of events. Once potential impact of direct buffered writes
// is that most End()s are just fast copies, but a small number of End()s turn into potentially expensive large write syscalls
// we need a way to test the average and the p99 cases.
func (e *Engine[T]) QueueEvent(evt []byte) {
	e.queue <- evt
}

type bwWrap struct {
	*bufio.Writer
}

func (bw bwWrap) Write(b []byte) (int, error) {
	// Ensure no write of a single event is split into 2 write syscalls
	if bw.Available() < len(b) {
		err := bw.Flush()
		if err != nil {
			return 0, err
		}
	}

	// bufio.Writer will already issue a single write for any incoming byte slices
	// that are larger than the buffer
	return bw.Writer.Write(b)
}

func (e *Engine[T]) ProcessEvents(w io.Writer) {
	// Writing in page size chunks seems to dramatically increases the write speed
	// returns diminish fairly quickly with increasing multiples of the page size
	// In minimal initial testing there doesn't seem to be much discernible difference after 2x
	bw := bwWrap{bufio.NewWriterSize(w, 8*os.Getpagesize())}
	ticker := time.NewTicker(1 * time.Minute)

	defer bw.Flush()

	for {
		select {
		case evt, ok := <-e.queue:
			if !ok {
				return
			}
			bw.Write(evt)
			// TODO: some kind of handler for a write failure or short write

			e.putBuffer(evt)
		case <-ticker.C:
			bw.Flush()
			// TODO: some kind of handler for a flush failure
		}
	}
}

func (e *Engine[T]) Close() {
	close(e.queue)
}

func (e *Engine[T]) getBuffer(len int) []byte {
	r := e.evtPool.Get()
	if r == nil {
		return make([]byte, 0, len)
	}

	ptr := r.(*byte)
	init := ([]byte)(unsafe.Slice(ptr, 4))
	cap := binary.LittleEndian.Uint32(init)
	init = ([]byte)(unsafe.Slice(ptr, cap))
	return init[:0]
}

func (e *Engine[T]) putBuffer(b []byte) {
	if cap(b) < minEventBufferSize || cap(b) > e.maxBuffer {
		return
	}

	binary.LittleEndian.AppendUint32(b[:0], uint32(cap(b)))
	e.evtPool.Put((*byte)(unsafe.SliceData(b)))
}

func NewEngine[T encoder]() *Engine[T] {
	chachaSeed := [32]byte{}
	crand.Read(chachaSeed[:])
	return &Engine[T]{
		evtPool:   sync.Pool{},
		queue:     make(chan []byte, eventQueueSize),
		idsrc:     rand.New(rand.NewChaCha8(chachaSeed)),
		maxBuffer: maxEventBufferSize, // 4 MiB
	}
}
