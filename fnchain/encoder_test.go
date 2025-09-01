package fnchain

import (
	"encoding/json"
	"errors"
	"io"
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
			dst := make([]byte, 0, 15)
			for b.Loop() {
				dst = dst[:0]
				dst = jsonEncoder{}.appendEscapedString(dst, tc.input)
			}
		})
		b.Run("std_"+tc.name, func(b *testing.B) {
			for b.Loop() {
				json.Marshal(tc.input)
			}
		})
	}
}

func BenchmarkMarshalling(b *testing.B) {
	b.Run("local_all_types", func(b *testing.B) {
		engine := NewEngine()
		go engine.ProcessEvents(io.Discard)
		defer engine.Close()

		err := errors.New("error")
		tm := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

		for b.Loop() {
			id, buf := engine.NewEvent()
			ctx := Ctx{
				enc:         jsonEncoder{},
				engine:      engine,
				buf:         buf,
				eventParent: 0,
				event:       id,
				typ:         0,
				lvl:         0,
			}
			ctx.With().Bool("bool", true).Dur("dur", time.Minute).
				Int("int", -42).Uint("uint", 42).Float64("float", 0.1).
				Str("str", "hello world").Time("time", tm).
				Msg("message").Err(err).Evt().End()
		}
	})
	b.Run("zerolog_all_types", func(b *testing.B) {
		logger := zerolog.New(io.Discard)

		err := errors.New("error")
		tm := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

		for b.Loop() {
			logger.Info().Bool("bool", true).Dur("dur", time.Minute).
				Int("int", -42).Uint("uint", 42).Float64("float", 0.1).Uint("type", 0).
				Str("str", "hello world").Time("time", tm).Err(err).Msg("message")
		}
	})
}
