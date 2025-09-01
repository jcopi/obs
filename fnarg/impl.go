package fnarg

import (
	"math/rand"
)

type Ctx struct {
	root   *Engine
	buf    []byte
	parent uint64
	this   uint64
}

func (c Ctx) With(meta ...Metadata) Ctx {
	for i := range meta {
		c.buf = meta[i](c.buf)
		if i != len(meta)-1 {
			c.buf = append(c.buf, ',')
		}
	}

	return c
}

func (c Ctx) End() {
	c.buf = append(c.buf, '}', '\n')

	c.root.QueueEvent(c.buf)
}

type Engine struct {
}

func (e *Engine) NewID() uint64 {
	return rand.Uint64()
}

func (e *Engine) QueueEvent(evt []byte) {

}
