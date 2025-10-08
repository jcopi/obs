package fnchain

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"strconv"
	"time"
	"unsafe"
)

type encoder interface {
	appendStart(dst []byte) []byte
	appendEnd(dst []byte) []byte

	// Fields
	appendBool(dst []byte, key string, b bool) []byte
	appendDur(dst []byte, key string, d time.Duration) []byte
	appendInt(dst []byte, key string, i int) []byte
	appendInt64(dst []byte, key string, i int64) []byte
	appendUint(dst []byte, key string, u uint) []byte
	appendUint64(dst []byte, key string, u uint64) []byte
	appendFloat32(dst []byte, key string, f float32) []byte
	appendFloat64(dst []byte, key string, f float64) []byte
	appendStr(dst []byte, key, s string) []byte
	appendStrs(dst []byte, key string, strs []string) []byte
	appendTime(dst []byte, key string, t time.Time) []byte
	appendHex(dst []byte, key string, b []byte) []byte
	appendB64(dst []byte, key string, b []byte) []byte
	appendID(dst []byte, key string, id uint64) []byte
	appendErr(dst []byte, key string, e error) []byte
}

type jsonEncoder struct{}

// appendB64 implements encoder.
func (j jsonEncoder) appendB64(dst []byte, key string, b []byte) []byte {
	dst = j.appendKeyOfPair(dst, key)
	dst = append(dst, '"')
	dst = base64.StdEncoding.AppendEncode(dst, b)
	return append(dst, '"', ',')
}

// appendID implements encoder.
func (j jsonEncoder) appendID(dst []byte, key string, id uint64) []byte {

	b := [8]byte{}
	return j.appendB64(dst, key, binary.LittleEndian.AppendUint64(b[:0], id))

}

var escapeTableIdx = [256]int8{
	// 0x00 - 0xf
	0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
	// 0x10 - 0x1f
	16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
	// 0x20 - 0x2f
	-1, -1, 32, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	// 0x30 - 0x3f
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	// 0x40 - 0x4f
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	// 0x50 - 0x5f
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, 33, -1, -1, -1,
	// 0x60 - 0x6f
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	// 0x70 - 0x7f
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
}

var escapeTable = [34]string{
	`\u0000`, `\u0001`, `\u0002`, `\u0003`, `\u0004`, `\u0005`, `\u0006`, `\u0007`, `\b`, `\t`,
	`\n`, `\u000b`, `\f`, `\r`, `\u000e`, `\u000f`, `\u0010`, `\u0011`, `\u0012`, `\u0013`,
	`\u0014`, `\u0015`, `\u0016`, `\u0017`, `\u0018`, `\u0019`, `\u001a`, `\u001b`, `\u001c`,
	`\u001d`, `\u001e`, `\u001f`, `\"`, `\\`,
}

func (j jsonEncoder) appendEscapedStringComplex(dst []byte, s string) []byte {
	for i := 0; i < len(s); i++ {
		if ei := escapeTableIdx[uint8(s[i])]; ei < 0 {
			dst = append(dst, s[i])
		} else {
			dst = append(dst, escapeTable[ei]...)
		}
	}
	return dst
}

// Accpets a normal string and returns a quoted string escaped for use in JSON
// This implementation should be optimized heavily to escape strings as fast as possible
// with minimal allocations. It should prefer optimizations that favor short strings.
func (j jsonEncoder) appendEscapedString(dst []byte, s string) []byte {
	// TODO: real implementation, this is a PoC hack
	dst = append(dst, '"')

	x := 8 * (len(s) / 8)
	for i := 0; i < x; i += 8 {
		if needsJSONEscape(*(*uint64)(unsafe.Pointer(unsafe.StringData(s[i:])))) == 0 {
			dst = append(dst, s[i:i+8]...)
		} else {
			dst = j.appendEscapedStringComplex(dst, s[i:i+8])
		}
	}

	dst = j.appendEscapedStringComplex(dst, s[x:])

	dst = append(dst, '"')
	return dst
}

func needsJSONEscape(u uint64) uint64 {
	// SWAR technique pulled from V8 string escaping
	// https://source.chromium.org/chromium/chromium/src/+/main:v8/src/json/json-stringifier.cc;l=522-538;drc=50ade2d8d071e10bc5d53234bts[2]c0ts[3]11c515940
	var mask0x20 uint64 = 0x2020202020202020
	var mask0x22 uint64 = 0x2222222222222222
	var mask0x5c uint64 = 0x5C5C5C5C5C5C5C5C
	var mask0x01 uint64 = 0x0101010101010101
	var maskMSB uint64 = 0x8080808080808080

	// Find control characters (< 0x20)
	hasCtrl := u - mask0x20
	hasDblQuote := (u ^ mask0x22) - mask0x01
	hasBackslash := (u ^ mask0x5c) - mask0x01
	resultMask := ^u & maskMSB
	return (hasCtrl | hasDblQuote | hasBackslash) & resultMask
}

func (j jsonEncoder) appendKeyOfPair(dst []byte, key string) []byte {
	dst = j.appendEscapedString(dst, key)
	return append(dst, ':')
}

// appendBool implements encoder.
func (j jsonEncoder) appendBool(dst []byte, key string, b bool) []byte {
	dst = j.appendKeyOfPair(dst, key)
	if b {
		return append(dst, "true,"...)
	}
	return append(dst, "false,"...)
}

// appendDur implements encoder.
func (j jsonEncoder) appendDur(dst []byte, key string, d time.Duration) []byte {
	return j.appendFloat64(dst, key, d.Seconds())
}

// appendEnd implements encoder.
func (j jsonEncoder) appendEnd(dst []byte) []byte {
	if dst[len(dst)-1] == ',' {
		dst = dst[:len(dst)-1]
	}
	return append(dst, '}', '\n')
}

// appendErr implements encoder.
func (j jsonEncoder) appendErr(dst []byte, key string, e error) []byte {
	return j.appendStr(dst, key, e.Error())
}

// appendFloat32 implements encoder.
func (j jsonEncoder) appendFloat32(dst []byte, key string, f float32) []byte {
	return j.appendFloat64(dst, key, float64(f))
}

// appendFloat64 implements encoder.
func (j jsonEncoder) appendFloat64(dst []byte, key string, f float64) []byte {
	dst = j.appendKeyOfPair(dst, key)
	dst = ftoa(dst, f)
	return append(dst, ',')
}

// appendHex implements encoder.
func (j jsonEncoder) appendHex(dst []byte, key string, b []byte) []byte {
	dst = j.appendKeyOfPair(dst, key)
	dst = append(dst, '"')
	dst = hex.AppendEncode(dst, b)
	return append(dst, '"', ',')
}

// appendInt implements encoder.
func (j jsonEncoder) appendInt(dst []byte, key string, i int) []byte {
	return j.appendInt64(dst, key, int64(i))
}

// appendInt64 implements encoder.
func (j jsonEncoder) appendInt64(dst []byte, key string, i int64) []byte {
	dst = j.appendKeyOfPair(dst, key)
	dst = strconv.AppendInt(dst, i, 10)
	return append(dst, ',')
}

// appendStart implements encoder.
func (j jsonEncoder) appendStart(dst []byte) []byte {
	return append(dst, '{')
}

// appendStr implements encoder.
func (j jsonEncoder) appendStr(dst []byte, key string, s string) []byte {
	dst = j.appendKeyOfPair(dst, key)
	dst = j.appendEscapedString(dst, s)
	return append(dst, ',')
}

// appendStrs implements encoder.
func (j jsonEncoder) appendStrs(dst []byte, key string, strs []string) []byte {
	dst = j.appendKeyOfPair(dst, key)
	dst = append(dst, '[')
	for i, s := range strs {
		dst = j.appendEscapedString(dst, s)
		if i < len(strs)-1 {
			dst = append(dst, ',')
		}
	}
	return append(dst, ']', ',')
}

// appendTime implements encoder.
func (j jsonEncoder) appendTime(dst []byte, key string, t time.Time) []byte {
	dst = j.appendKeyOfPair(dst, key)
	dst = append(dst, '"')
	dst = t.AppendFormat(dst, time.RFC3339)
	return append(dst, '"', ',')
}

// appendUint implements encoder.
func (j jsonEncoder) appendUint(dst []byte, key string, u uint) []byte {
	return j.appendUint64(dst, key, uint64(u))
}

// appendUint64 implements encoder.
func (j jsonEncoder) appendUint64(dst []byte, key string, u uint64) []byte {
	dst = j.appendKeyOfPair(dst, key)
	dst = strconv.AppendUint(dst, u, 10)
	return append(dst, ',')
}

var _ encoder = jsonEncoder{}
