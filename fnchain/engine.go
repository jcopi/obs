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
const eventQueueSize = 196

type Engine[T encoder] interface {
	RootEvent(lvl Level, typ EvtType) Ctx[T]
	QueueEvent(evt []byte)
	ProcessEvents()
	Close()

	initEvent(minRequiredBuffer int) (uint64, []byte)
}

type bufferPool struct {
	sync.Pool
}

func (bp *bufferPool) getBuffer(len int) []byte {
	r := bp.Pool.Get()
	if r == nil {
		return make([]byte, 0, len)
	}

	ptr := r.(*byte)
	init := ([]byte)(unsafe.Slice(ptr, 4))
	cap := binary.LittleEndian.Uint32(init)
	init = ([]byte)(unsafe.Slice(ptr, cap))
	return init[:0]
}

func (bp *bufferPool) putBuffer(b []byte, maxSize int) {
	if cap(b) < minEventBufferSize || cap(b) > maxSize {
		return
	}

	binary.LittleEndian.AppendUint32(b[:0], uint32(cap(b)))
	bp.Pool.Put((*byte)(unsafe.SliceData(b)))
}

type engine[T encoder] struct {
	w io.Writer
	bufferPool
	queue     chan []byte
	idsrc     *rand.Rand
	maxBuffer int
}

func (e *engine[T]) RootEvent(lvl Level, typ EvtType) Ctx[T] {
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

func (e *engine[T]) initEvent(knownBufferSize int) (uint64, []byte) {
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
//
// Benchmarks show that this method produces a large amount of lock contention overhead
func (e *engine[T]) QueueEvent(evt []byte) {
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

func (e *engine[T]) ProcessEvents() {
	// Writing in page size chunks seems to dramatically increases the write speed
	// returns diminish fairly quickly with increasing multiples of the page size
	// In minimal initial testing there doesn't seem to be much discernible difference after 2x
	bw := bwWrap{bufio.NewWriterSize(e.w, 2*os.Getpagesize())}
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

			e.putBuffer(evt, e.maxBuffer)
		case <-ticker.C:
			bw.Flush()
			// TODO: some kind of handler for a flush failure
		}
	}
}

func (e *engine[T]) Close() {
	close(e.queue)
}

func NewEngine[T encoder](w io.Writer) Engine[T] {
	chachaSeed := [32]byte{}
	crand.Read(chachaSeed[:])
	return &engine[T]{
		w:          w,
		bufferPool: bufferPool{},
		queue:      make(chan []byte, eventQueueSize),
		idsrc:      rand.New(rand.NewChaCha8(chachaSeed)),
		maxBuffer:  maxEventBufferSize, // 4 MiB
	}
}
