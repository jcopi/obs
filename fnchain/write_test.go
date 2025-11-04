package fnchain

import (
	"bufio"
	"errors"
	"io"
	"os"
	"testing"
	"time"
)

type nopEngine struct {
	*engine
}

func (d nopEngine) QueueEvent(evt []byte) {
	d.putBuffer(evt, d.maxBuffer)
}

func (d nopEngine) ProcessEvents() {}

func (d nopEngine) RootEvent(lvl Level, typ EvtType) *Ctx {
	ctx := d.engine.RootEvent(lvl, typ)
	ctx.engine = d
	return ctx
}

func newNop() Engine {
	return nopEngine{NewEngine(io.Discard).(*engine)}
}

type directEngine struct {
	*engine
}

func (d directEngine) QueueEvent(evt []byte) {
	d.w.Write(evt)
	d.putBuffer(evt, d.maxBuffer)
}

func (d directEngine) ProcessEvents() {}

func (d directEngine) RootEvent(lvl Level, typ EvtType) *Ctx {
	ctx := d.engine.RootEvent(lvl, typ)
	ctx.engine = d
	return ctx
}

func newDirect(w io.Writer) Engine {
	return directEngine{NewEngine(w).(*engine)}
}

func noError(t testing.TB, err error) {
	if err != nil {
		t.Error(err)
	}
}

func BenchmarkWrite(b *testing.B) {
	//constants to use in the BenchmarkWrite
	constts := time.Date(2025, time.September, 15, 11, 39, 13, 00, time.UTC)
	consterr := errors.New("example error with a generic string")
	devnull, err := os.OpenFile(os.DevNull, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	noError(b, err)
	b.Cleanup(func() { devnull.Close() })
	f, err := os.CreateTemp(b.TempDir(), "*.log")
	noError(b, err)
	b.Cleanup(func() { f.Close() })

	// engine and write lists for building a
	// benchmarking matrix
	engines := []struct {
		name string
		eng  Engine
	}{
		{
			name: "___nop",
			eng:  newNop(),
		},
		{
			name: "d_null",
			eng:  newDirect(devnull),
		},
		{
			name: "d_file",
			eng:  newDirect(f),
		},
		{
			name: "b_null",
			eng:  newDirect(bufio.NewWriterSize(devnull, os.Getpagesize())),
		},
		{
			name: "b_file",
			eng:  newDirect(bufio.NewWriterSize(f, os.Getpagesize())),
		},
		{
			name: "i_null",
			eng: func() Engine {
				eng := NewEngine(devnull)
				b.Cleanup(eng.Close)
				go eng.ProcessEvents()
				return eng
			}(),
		},
		{
			name: "i_file",
			eng: func() Engine {
				eng := NewEngine(f)
				b.Cleanup(eng.Close)
				go eng.ProcessEvents()
				return eng
			}(),
		},
	}

	writes := []struct {
		name string
		fn   func(root *Ctx)
	}{
		{
			name: "minimal",
			fn: func(root *Ctx) {
				ctx := root.Event(NoLevel, GenericEvent)
				ctx.End()
			},
		},
		{
			name: "small__",
			fn: func(root *Ctx) {
				ctx := root.Event(NoLevel, GenericEvent)
				ctx.With().Bool("flag", false).Int("count", 42).Str("test", "test_string").Evt().End()
			},
		},
		{
			name: "medium_",
			fn: func(root *Ctx) {
				ctx := root.Event(NoLevel, GenericEvent)
				ctx.With().Bool("bool", true).Dur("dur", time.Minute).
					Int("int", -42).Uint("uint", 42).Float64("float", 0.1).
					Str("str", "hello world").Time("time", constts).
					Msg("message").Err(consterr).Evt().End()
			},
		},
		{
			name: "large__",
			fn: func(root *Ctx) {
				ctx := root.Event(NoLevel, GenericEvent)
				ctx.With().
					Str("go.version", "go1.23.7").
					Str("http.request.host", "something-registry.region.subdomain.example.invalid").
					Str("http.request.id", "bdbb3a1b-e3e4-495f-9ea2-7a9a7cfc6214").
					Str("http.request.method", "GET").
					Str("http.request.remoteaddr", "172.1.42.0").
					Str("http.request.uri", "/v2/image-namespace/image-path/image-name/blobs/sha256:198d5096c399a7cab47a38d8178534daf9f131d53aae4d07f6c6685e289df732").
					Str("http.request.useragent", "Buildah/1.40.1").
					Str("http.response.contenttype", "application/octet-stream").
					Dur("http.response.duration", time.Duration(2.117177559*float64(time.Second))).
					Int("http.response.status", 200).
					Int("http.response.written", 81745310).
					Str("instance.id", "432f9512-9efb-4b64-ad86-afaabcfe3306").
					Msg("response completed").
					Time("time", constts).
					Str("vars.digest", "sha256:198d5096c399a7cab47a38d8178534daf9f131d53aae4d07f6c6685e289df732").
					Str("vars.name", "image-namespace/image-path/image-name").
					Str("version", "3.0.0").
					Evt().
					End()
			},
		},
	}

	for _, write := range writes {
		for _, eng := range engines {
			b.Run(write.name+"_"+eng.name, func(b *testing.B) {
				root := eng.eng.RootEvent(NoLevel, GenericEvent)
				for b.Loop() {
					write.fn(root)
				}
			})
		}
	}
}
