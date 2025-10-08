package fnchain

import (
	"bufio"
	"encoding/json"
	"errors"
	"math"
	"math/rand/v2"
	"os"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestEscape(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{name: "happy path", input: "0123456789_-+=:'/.,><abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"},
		{name: "happy path", input: "\t\r\n /?_\\\"\000\001\002\003"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dst := make([]byte, 0, 10)
			dst = (jsonEncoder{}).appendEscapedString(dst, tc.input)
			var out string
			err := json.Unmarshal(dst, &out)
			require.NoError(t, err)

			require.Equal(t, tc.input, out)
		})
	}
}

func FuzzJsonEscape(f *testing.F) {
	corpus := []string{
		"0123456789_-+=:'/.,><abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
		"\t\r\n /?_\\\"\000\001\002\003",
		"", " ", "\000",
		"✅🙃🤦👍", "\\\"\000\001\002✅🙃\t\r\n 🤦👍",
		"a", "ab", "abc", "abcd", "abcde", "abcdef", "abcdefg", "abcdefgh", "abcdefghi",
	}
	for _, s := range corpus {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		dst := make([]byte, 0, 10)
		dst = (jsonEncoder{}).appendEscapedString(dst, input)
		var out string
		err := json.Unmarshal(dst, &out)
		if err != nil {
			t.Errorf("failed to unmarshal result (%s) of marshalling %s", dst, input)
		}
		if input != out && utf8.ValidString(input) {
			t.Errorf("unmarshalled string (%x) does match original %x from encoded %x", out, input, dst)
		}
	})
}

func BenchmarkEncodeString(b *testing.B) {
	cases := []struct {
		name  string
		input string
	}{
		{name: "ne_long", input: "0123456789_-+=:'/.,><abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"},
		{name: "esc_med", input: "\t\r\n /?_\\\"\000\001\002\003'"},
		{name: "ne_short1", input: "a"},
		{name: "ne_short2", input: "ab"},
		{name: "ne_short3", input: "abc"},
		{name: "ne_short4", input: "abcd"},
		{name: "ne_short5", input: "abcde"},
		{name: "ne_short6", input: "abcdef"},
		{name: "ne_short7", input: "abcdefg"},
		{name: "ne_short8", input: "abcdefgh"},
		{name: "ne_short9", input: "abcdefghi"},
		{name: "emoji", input: "✅🙃🤦👍"},
		{name: "esc_emoji", input: "\\\"\000\001\002✅🙃\t\r\n 🤦👍"},
		{name: "esc_last", input: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\""},
		{name: "esc_mid", input: "aaaaaaaaaaaaaaaaaaaaaaaaa\"aaaaaaaaaaaaaaaaaaaaaaaa"},
		{name: "esc_first", input: "\"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
	}

	for _, tc := range cases {
		b.Run("local_"+tc.name, func(b *testing.B) {
			dst := make([]byte, 0, 100)
			for b.Loop() {
				dst = dst[:0]
				dst = jsonEncoder{}.appendEscapedString(dst, tc.input)
			}
		})
		// b.Run("local2_"+tc.name, func(b *testing.B) {
		// 	dst := make([]byte, 0, 100)
		// 	for b.Loop() {
		// 		dst = dst[:0]
		// 		dst = jsonEncoder{}.appendEscapedString2(dst, tc.input)
		// 	}
		// })
	}
}

func BenchmarkEncodeStringInternal(b *testing.B) {
	cases := []struct {
		name  string
		input string
	}{
		{name: "ne_1", input: "a"},
		{name: "ne_2", input: "ab"},
		{name: "ne_3", input: "abc"},
		{name: "ne_4", input: "abcd"},
		{name: "ne_5", input: "abcde"},
		{name: "ne_6", input: "abcdef"},
		{name: "ne_7", input: "abcdefg"},
		{name: "ne_8", input: "abcdefgh"},
		{name: "es_1", input: "\x00"},
		{name: "es_2", input: "\x00b"},
		{name: "es_3", input: "\x00bc"},
		{name: "es_4", input: "\x00bcd"},
		{name: "es_5", input: "\x00bcde"},
		{name: "es_6", input: "\x00bcdef"},
		{name: "es_7", input: "\x00bcdefg"},
		{name: "es_8", input: "\x00bcdefgh"},
		{name: "ee_1", input: "\x00"},
		{name: "ee_2", input: "a\x00"},
		{name: "ee_3", input: "ab\x00"},
		{name: "ee_4", input: "abc\x00"},
		{name: "ee_5", input: "abcd\x00"},
		{name: "ee_6", input: "abcde\x00"},
		{name: "ee_7", input: "abcdef\x00"},
		{name: "ee_8", input: "abcdefg\x00"},
	}

	for _, tc := range cases {
		b.Run("local_"+tc.name, func(b *testing.B) {
			dst := make([]byte, 0, 32)
			for b.Loop() {
				dst = dst[:0]
				dst = jsonEncoder{}.appendEscapedStringComplex(dst, tc.input)
			}
		})
		// b.Run("local2_"+tc.name, func(b *testing.B) {
		// 	dst := make([]byte, 0, 32)
		// 	for b.Loop() {
		// 		dst = dst[:0]
		// 		dst = jsonEncoder{}.appendEscapedStringComplex2(dst, tc.input)
		// 	}
		// })
	}
}

func BenchmarkEncodeID(b *testing.B) {
	id0 := uint64(0)
	idRand := rand.Uint64()
	idMax := uint64(math.MaxUint64)

	b.Run("b64_0", func(b *testing.B) {
		buf := [24]byte{}
		bs := buf[:0]

		for b.Loop() {
			bs = jsonEncoder{}.appendID(bs[:0], "id", id0)
		}
	})
	b.Run("b64_rand", func(b *testing.B) {
		buf := [24]byte{}
		bs := buf[:0]

		for b.Loop() {
			bs = jsonEncoder{}.appendID(bs[:0], "id", idRand)
		}
	})
	b.Run("b64_max", func(b *testing.B) {
		buf := [24]byte{}
		bs := buf[:0]

		for b.Loop() {
			bs = jsonEncoder{}.appendID(bs[:0], "id", idMax)
		}
	})
	b.Run("b10_0", func(b *testing.B) {
		buf := [24]byte{}
		bs := buf[:0]

		for b.Loop() {
			bs = jsonEncoder{}.appendUint64(bs[:0], "id", id0)
		}
	})
	b.Run("b10_rand", func(b *testing.B) {
		buf := [24]byte{}
		bs := buf[:0]

		for b.Loop() {
			bs = jsonEncoder{}.appendUint64(bs[:0], "id", idRand)
		}
	})
	b.Run("b10_max", func(b *testing.B) {
		buf := [24]byte{}
		bs := buf[:0]

		for b.Loop() {
			bs = jsonEncoder{}.appendUint64(bs[:0], "id", idMax)
		}
	})
}

func BenchmarkMarshalling(b *testing.B) {
	dir := b.TempDir()

	b.Run("local_all_types", func(b *testing.B) {
		f, err := os.CreateTemp(dir, "local_all_types.log")
		if err != nil {
			b.Error(err)
		}
		defer f.Close()

		engine := NewEngine[jsonEncoder]()
		go engine.ProcessEvents(f)
		defer engine.Close()

		err = errors.New("error")
		tm := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

		root := engine.RootEvent(NoLevel, GenericEvent)

		for b.Loop() {
			ctx := root.Event(NoLevel, GenericEvent)
			ctx.With().Bool("bool", true).Dur("dur", time.Minute).
				Int("int", -42).Uint("uint", 42).Float64("float", 0.1).
				Str("str", "hello world").Time("time", tm).
				Msg("message").Err(err).Evt().End()
		}
	})
	b.Run("zerolog_all_types", func(b *testing.B) {
		f, err := os.CreateTemp(dir, "zerolog_all_types.log")
		if err != nil {
			b.Error(err)
		}
		w := bufio.NewWriterSize(f, 2*os.Getpagesize())
		defer w.Flush()
		defer f.Close()
		logger := zerolog.New(bwWrap{w})

		err = errors.New("error")
		tm := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

		for b.Loop() {
			logger.Info().Bool("bool", true).Dur("dur", time.Minute).
				Int("int", -42).Uint("uint", 42).Float64("float", 0.1).Uint("type", 0).
				Str("str", "hello world").Time("time", tm).Err(err).Msg("message")
		}
	})
	b.Run("local_strings", func(b *testing.B) {
		f, err := os.CreateTemp(dir, "local_strings.log")
		if err != nil {
			b.Error(err)
		}
		defer f.Close()

		engine := NewEngine[jsonEncoder]()
		go engine.ProcessEvents(f)
		defer engine.Close()

		root := engine.RootEvent(NoLevel, GenericEvent)

		for b.Loop() {

			ctx := root.Event(NoLevel, GenericEvent)
			ctx.With().
				Str("str", "hello world").
				Str("str_esc", "\t\r\n /?_\\\"\000\001\002\003").
				Msg("message").
				Evt().
				End()
		}
	})
	b.Run("zerolog_strings", func(b *testing.B) {
		f, err := os.CreateTemp(dir, "local_strings.log")
		if err != nil {
			b.Error(err)
		}
		w := bufio.NewWriterSize(f, 2*os.Getpagesize())
		defer w.Flush()
		defer f.Close()

		logger := zerolog.New(bwWrap{w})

		for b.Loop() {
			logger.Info().
				Str("str", "hello world").
				Str("str_esc", "\t\r\n /?_\\\"\000\001\002\003").
				Uint("type", 0). // Equivalent of the event's type which is always included by End
				Msg("message")
		}
	})
	b.Run("local_real_log", func(b *testing.B) {
		f, err := os.CreateTemp(dir, "local_strings.log")
		if err != nil {
			b.Error(err)
		}
		defer f.Close()

		engine := NewEngine[jsonEncoder]()
		go engine.ProcessEvents(f)
		defer engine.Close()

		root := engine.RootEvent(NoLevel, GenericEvent)

		for b.Loop() {
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
				Time("time", time.Date(2025, time.September, 15, 11, 39, 13, 00, time.UTC)).
				Str("vars.digest", "sha256:198d5096c399a7cab47a38d8178534daf9f131d53aae4d07f6c6685e289df732").
				Str("vars.name", "image-namespace/image-path/image-name").
				Str("version", "3.0.0").
				Evt().
				End()
		}
	})
	b.Run("zerolog_real_log", func(b *testing.B) {
		f, err := os.CreateTemp(dir, "local_strings.log")
		if err != nil {
			b.Error(err)
		}
		w := bufio.NewWriterSize(f, 2*os.Getpagesize())
		defer w.Flush()
		defer f.Close()

		logger := zerolog.New(bwWrap{w})

		for b.Loop() {
			logger.Info().
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
				Time("time", time.Date(2025, time.September, 15, 11, 39, 13, 00, time.UTC)).
				Str("vars.digest", "sha256:198d5096c399a7cab47a38d8178534daf9f131d53aae4d07f6c6685e289df732").
				Str("vars.name", "image-namespace/image-path/image-name").
				Str("version", "3.0.0").
				Uint("type", 0). // Equivalent of the event's type which is always included by End
				Msg("response completed")
		}
	})
}
