package json

import (
	"encoding/json"
	"testing"
	"unicode/utf8"
)

func TestAssertIndex(t *testing.T) {
	for i, idx := range escapeTableIdx {
		if int(idx) >= len(escapeTable) {
			t.Errorf("escape table index at %v (%v) is too large", i, idx)
		}
	}
}

func FuzzJsonEscape(f *testing.F) {
	corpus := []string{
		"0123456789_-+=:'/.,><abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
		"\t\r\n /?_\\\"\000\001\002\003",
		"", " ", "\000",
		"✅🙃🤦👍", "\\\"\000\001\002✅🙃\t\r\n 🤦👍",
		"a", "ab", "abc", "abcd", "abcde", "abcdef", "abcdefg", "abcdefgh", "abcdefghi",
		`\"\\"\'"\\"\'|'''|\\\"\"`, "ndjfnjknfjkndjkvnßtj13op4k5tr893ur=-132plf[350qi0]lgf[n2g\"",
		"  \n  \t  \f  ", "        ", "         ",
	}
	for _, s := range corpus {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		dst := make([]byte, 0, 10)
		dst = AppendEscapedString(dst, input)
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
		b.Run(tc.name, func(b *testing.B) {
			dst := make([]byte, 0, 100)
			for b.Loop() {
				dst = dst[:0]
				dst = AppendEscapedString(dst, tc.input)
			}
		})
	}
}

func BenchmarkAppendStringRemainder(b *testing.B) {
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
		{name: "es_1", input: "\x00"},
		{name: "es_2", input: "\x00b"},
		{name: "es_3", input: "\x00bc"},
		{name: "es_4", input: "\x00bcd"},
		{name: "es_5", input: "\x00bcde"},
		{name: "es_6", input: "\x00bcdef"},
		{name: "es_7", input: "\x00bcdefg"},
		{name: "ee_1", input: "\x00"},
		{name: "ee_2", input: "a\x00"},
		{name: "ee_3", input: "ab\x00"},
		{name: "ee_4", input: "abc\x00"},
		{name: "ee_5", input: "abcd\x00"},
		{name: "ee_6", input: "abcde\x00"},
		{name: "ee_7", input: "abcdef\x00"},
		{name: "e_all", input: "\x00\x19\n\r\t\\\""},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			dst := make([]byte, 0, 32)
			for b.Loop() {
				dst = dst[:0]
				dst = appendStringRemainder(dst, tc.input)
			}
		})
	}
}
