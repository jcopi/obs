package fnchain

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"strconv"
	"time"
	"unsafe"
)

// The jsonEncoder struct is defined here only as a convenient way to organize the encoding methods
// we could get similar organization by moving these to an internal package and exporting all of the functions
type jsonEncoder struct{}

var escapeTableIdx = [256]int8{
	0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
	16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
	-1, -1, 32, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, 33, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
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

	x := (len(s) / 8) * 8
	for i := 0; i < x; i += 8 {
		if needsJSONEscape(*(*uint64)(unsafe.Pointer(unsafe.StringData(s[i:])))) == 0 {
			// TODO: defer all copies of substrings that don't require escaping
			// until a substring does require a copy or the end of the string is reached
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
	const mask0x20 uint64 = 0x2020202020202020
	const mask0x22 uint64 = 0x2222222222222222
	const mask0x5c uint64 = 0x5C5C5C5C5C5C5C5C
	const mask0x01 uint64 = 0x0101010101010101
	const maskMSB uint64 = 0x8080808080808080

	// Find control characters (< 0x20)
	hasCtrl := u - mask0x20
	hasDblQuote := (u ^ mask0x22) - mask0x01
	hasBackslash := (u ^ mask0x5c) - mask0x01
	resultMask := ^u & maskMSB
	return (hasCtrl | hasDblQuote | hasBackslash) & resultMask
}

// func needsJSONEscape32(u uint32) uint32 {
// 	// SWAR technique pulled from V8 string escaping
// 	// https://source.chromium.org/chromium/chromium/src/+/main:v8/src/json/json-stringifier.cc;l=522-538;drc=50ade2d8d071e10bc5d53234bts[2]c0ts[3]11c515940
// 	const mask0x20 uint32 = 0x20202020
// 	const mask0x22 uint32 = 0x22222222
// 	const mask0x5c uint32 = 0x5C5C5C5C
// 	const mask0x01 uint32 = 0x0101010
// 	const maskMSB uint32 = 0x80808080

// 	// Find control characters (< 0x20)
// 	hasCtrl := u - mask0x20
// 	hasDblQuote := (u ^ mask0x22) - mask0x01
// 	hasBackslash := (u ^ mask0x5c) - mask0x01
// 	resultMask := ^u & maskMSB
// 	return (hasCtrl | hasDblQuote | hasBackslash) & resultMask
// }

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

// appendB64 implements encoder.
func (j jsonEncoder) appendB64(dst []byte, key string, b []byte) []byte {
	dst = j.appendKeyOfPair(dst, key)
	dst = append(dst, '"')
	dst = base64.StdEncoding.AppendEncode(dst, b)
	return append(dst, '"', ',')
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

// For fields that will have a constant string key and are in the extreme hot path
// (used in every event write) have some keys defined that skip unnecessary string
// escaping. The strings in these fields are known to be valid JSON strings w/o escaping

// append a const key that is known to not require any escaping
func (j jsonEncoder) appendKnownKeyOfPair(dst []byte, knownKey string) []byte {
	dst = append(dst, '"')
	dst = append(dst, knownKey...)
	return append(dst, '"', ':')
}

func (j jsonEncoder) appendKnownKeyID(dst []byte, knownKey string, id uint64) []byte {
	// will be common to all events, it makes sense to define as efficient of
	// an encoding as possible. Currently the encoding is not maximally efficient
	b := [8]byte{}
	binary.LittleEndian.AppendUint64(b[:0], id)
	dst = j.appendKnownKeyOfPair(dst, knownKey)
	dst = append(dst, '"')
	dst = hex.AppendEncode(dst, b[:])
	return append(dst, '"', ',')
}

func (j jsonEncoder) appendKnownKeyType(dst []byte, knownKey string, typ EvtType) []byte {
	// will be common to all events, it makes sense to define as efficient of
	// an encoding as possible. Currently the encoding is not maximally efficient
	dst = j.appendKnownKeyOfPair(dst, knownKey)
	dst = strconv.AppendUint(dst, uint64(typ), 16)
	return append(dst, ',')
}

func (j jsonEncoder) appendKnownKeyLevel(dst []byte, knownKey string, lvl Level) []byte {
	// will be common to all events, it makes sense to define as efficient of
	// an encoding as possible. Currently the encoding is not maximally efficient
	dst = j.appendKnownKeyOfPair(dst, knownKey)
	dst = append(dst, '"')
	dst = AppendLevel(dst, lvl)
	return append(dst, '"', ',')
}
