package fnchain

import (
	"encoding/hex"
	"strconv"
	"time"
)

type encoder interface {
	needsSeperator(dst []byte) bool
	appendStart(dst []byte) []byte
	appendSeperator(dst []byte) []byte
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
	appendErr(dst []byte, key string, e error) []byte
}

type jsonEncoder struct{}

// needsSeperator implements encoder.
func (j jsonEncoder) needsSeperator(dst []byte) bool {
	return len(dst) > 0 && dst[len(dst)-1] == '{'
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
	`\u0000`, `\u0001`, `\u0002`, `\u0003`, `\u0004`, `\u0005`, `\u0006`, `\u0007`, `\b`, `\u0009`,
	`\n`, `\u000b`, `\f`, `\r`, `\u000e`, `\u000f`, `\u0010`, `\u0011`, `\u0012`, `\u0013`,
	`\u0014`, `\u0015`, `\u0016`, `\u0017`, `\u0018`, `\u0019`, `\u001a`, `\u001b`, `\u001c`,
	`\u001d`, `\u001e`, `\u001f`, `\"`, `\\`,
}

// Accpets a normal string and returns a quoted string escaped for use in JSON
// This implementation should be optimized heavily to escape strings as fast as possible
// with minimal allocations. It should prefer optimizations that favor short strings.
func (j jsonEncoder) appendEscapedString(dst []byte, s string) []byte {
	// TODO: real implementation, this is a PoC hack
	dst = append(dst, '"')
	for i := 0; i < len(s); i += 8 {
		var b0, b1, b2, b3, b4, b5, b6, b7 byte
		r := len(s) - i
		switch r {
		case 1:
			b0 = s[i]

			if needsJSONEscape(s[i]) {
				dst = append(dst, escapeTable[escapeTableIdx[s[i]]]...)
			} else {
				dst = append(dst, b0)
			}
		case 2:
			b0 = s[i]
			b1 = s[i+1]
			if needsJSONEscape((uint16(b0) << 8) | uint16(b1)) {
				if ei := escapeTableIdx[b0]; ei < 0 {
					dst = append(dst, b0)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
				if ei := escapeTableIdx[b1]; ei < 0 {
					dst = append(dst, b1)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
			} else {
				dst = append(dst, b0, b1)
			}
		case 3:
			b0 = s[i]
			b1 = s[i+1]
			b2 = s[i+2]
			if needsJSONEscape((uint32(b0) << 24) | (uint32(b1) << 16) | (uint32(b2) << 8) | 'A') {
				if ei := escapeTableIdx[b0]; ei < 0 {
					dst = append(dst, b0)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
				if ei := escapeTableIdx[b1]; ei < 0 {
					dst = append(dst, b1)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
				if ei := escapeTableIdx[b2]; ei < 0 {
					dst = append(dst, b2)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
			} else {
				dst = append(dst, b0, b1, b2)
			}
		case 4:
			b0 = s[i]
			b1 = s[i+1]
			b2 = s[i+2]
			b3 = s[i+3]
			if needsJSONEscape((uint32(b0) << 24) | (uint32(b1) << 16) | (uint32(b2) << 8) | uint32(b3)) {
				if ei := escapeTableIdx[b0]; ei < 0 {
					dst = append(dst, b0)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
				if ei := escapeTableIdx[b1]; ei < 0 {
					dst = append(dst, b1)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
				if ei := escapeTableIdx[b2]; ei < 0 {
					dst = append(dst, b2)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
				if ei := escapeTableIdx[b3]; ei < 0 {
					dst = append(dst, b3)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
			} else {
				dst = append(dst, b0, b1, b2, b3)
			}
		case 5:
			b0 = s[i]
			b1 = s[i+1]
			b2 = s[i+2]
			b3 = s[i+3]
			b4 = s[i+4]
			if needsJSONEscape((uint64(b0) << 56) | (uint64(b1) << 48) | (uint64(b2) << 40) | (uint64(b3) << 32) | (uint64(b4) << 24) | (uint64('A') << 16) | (uint64('A') << 8) | uint64('A')) {
				if ei := escapeTableIdx[b0]; ei < 0 {
					dst = append(dst, b0)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
				if ei := escapeTableIdx[b1]; ei < 0 {
					dst = append(dst, b1)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
				if ei := escapeTableIdx[b2]; ei < 0 {
					dst = append(dst, b2)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
				if ei := escapeTableIdx[b3]; ei < 0 {
					dst = append(dst, b3)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
				if ei := escapeTableIdx[b4]; ei < 0 {
					dst = append(dst, b4)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
			} else {
				dst = append(dst, b0, b1, b2, b3, b4)
			}
		case 6:
			b0 = s[i]
			b1 = s[i+1]
			b2 = s[i+2]
			b3 = s[i+3]
			b4 = s[i+4]
			b5 = s[i+5]
			if needsJSONEscape((uint64(b0) << 56) | (uint64(b1) << 48) | (uint64(b2) << 40) | (uint64(b3) << 32) | (uint64(b4) << 24) | (uint64(b5) << 16) | (uint64('A') << 8) | 'A') {
				if ei := escapeTableIdx[b0]; ei < 0 {
					dst = append(dst, b0)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
				if ei := escapeTableIdx[b1]; ei < 0 {
					dst = append(dst, b1)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
				if ei := escapeTableIdx[b2]; ei < 0 {
					dst = append(dst, b2)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
				if ei := escapeTableIdx[b3]; ei < 0 {
					dst = append(dst, b3)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
				if ei := escapeTableIdx[b4]; ei < 0 {
					dst = append(dst, b4)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
				if ei := escapeTableIdx[b5]; ei < 0 {
					dst = append(dst, b5)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
			} else {
				dst = append(dst, b0, b1, b2, b3, b4, b5)
			}
		case 7:
			b0 = s[i]
			b1 = s[i+1]
			b2 = s[i+2]
			b3 = s[i+3]
			b4 = s[i+4]
			b5 = s[i+5]
			b6 = s[i+6]
			if needsJSONEscape((uint64(b0) << 56) | (uint64(b1) << 48) | (uint64(b2) << 40) | (uint64(b3) << 32) | (uint64(b4) << 24) | (uint64(b5) << 16) | (uint64(b6) << 8) | 'A') {
				if ei := escapeTableIdx[b0]; ei < 0 {
					dst = append(dst, b0)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
				if ei := escapeTableIdx[b1]; ei < 0 {
					dst = append(dst, b1)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
				if ei := escapeTableIdx[b2]; ei < 0 {
					dst = append(dst, b2)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
				if ei := escapeTableIdx[b3]; ei < 0 {
					dst = append(dst, b3)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
				if ei := escapeTableIdx[b4]; ei < 0 {
					dst = append(dst, b4)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
				if ei := escapeTableIdx[b5]; ei < 0 {
					dst = append(dst, b5)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
				if ei := escapeTableIdx[b6]; ei < 0 {
					dst = append(dst, b6)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
			} else {
				dst = append(dst, b0, b1, b2, b3, b4, b5, b6)
			}
		default:
			b0 = s[i]
			b1 = s[i+1]
			b2 = s[i+2]
			b3 = s[i+3]
			b4 = s[i+4]
			b5 = s[i+5]
			b6 = s[i+6]
			b7 = s[i+7]
			if needsJSONEscape((uint64(b0) << 56) | (uint64(b1) << 48) | (uint64(b2) << 40) | (uint64(b3) << 32) | (uint64(b4) << 24) | (uint64(b5) << 16) | (uint64(b6) << 8) | uint64(b7)) {
				if ei := escapeTableIdx[b0]; ei < 0 {
					dst = append(dst, b0)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
				if ei := escapeTableIdx[b1]; ei < 0 {
					dst = append(dst, b1)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
				if ei := escapeTableIdx[b2]; ei < 0 {
					dst = append(dst, b2)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
				if ei := escapeTableIdx[b3]; ei < 0 {
					dst = append(dst, b3)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
				if ei := escapeTableIdx[b4]; ei < 0 {
					dst = append(dst, b4)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
				if ei := escapeTableIdx[b5]; ei < 0 {
					dst = append(dst, b5)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
				if ei := escapeTableIdx[b6]; ei < 0 {
					dst = append(dst, b6)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
				if ei := escapeTableIdx[b7]; ei < 0 {
					dst = append(dst, b7)
				} else {
					dst = append(dst, escapeTable[ei]...)
				}
			} else {
				dst = append(dst, b0, b1, b2, b3, b4, b5, b6, b7)
			}
		}
	}
	dst = append(dst, '"')
	return dst
}

func needsJSONEscape[T uint8 | uint16 | uint32 | uint64](u T) bool {
	// SWAR technique pulled from V8 string escaping
	// https://source.chromium.org/chromium/chromium/src/+/main:v8/src/json/json-stringifier.cc;l=522-538;drc=50ade2d8d071e10bc5d53234bb2c0b311c515940
	var mask0x20 uint64 = 0x2020202020202020
	var mask0x22 uint64 = 0x2222222222222222
	var mask0x5c uint64 = 0x5C5C5C5C5C5C5C5C
	var mask0x01 uint64 = 0x0101010101010101
	var maskMSB uint64 = 0x8080808080808080

	// Find control characters (< 0x20)
	hasCtrl := u - T(mask0x20)
	hasDblQuote := (u ^ T(mask0x22)) - T(mask0x01)
	hasBackslash := (u ^ T(mask0x5c)) - T(mask0x01)
	resultMask := ^u & T(maskMSB)
	result := (hasCtrl | hasDblQuote | hasBackslash) & resultMask

	return result != 0
}

func (j jsonEncoder) appendKeyOfPair(dst []byte, key string) []byte {
	dst = j.appendEscapedString(dst, key)
	return append(dst, ':')
}

// appendBool implements encoder.
func (j jsonEncoder) appendBool(dst []byte, key string, b bool) []byte {
	dst = j.appendKeyOfPair(dst, key)
	if b {
		return append(dst, "true"...)
	}
	return append(dst, "false"...)
}

// appendDur implements encoder.
func (j jsonEncoder) appendDur(dst []byte, key string, d time.Duration) []byte {
	return j.appendFloat64(dst, key, d.Seconds())
}

// appendEnd implements encoder.
func (j jsonEncoder) appendEnd(dst []byte) []byte {
	return append(dst, '}')
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
	return strconv.AppendFloat(dst, f, 'g', -1, 64)
}

// appendHex implements encoder.
func (j jsonEncoder) appendHex(dst []byte, key string, b []byte) []byte {
	dst = j.appendKeyOfPair(dst, key)
	dst = append(dst, '"')
	dst = hex.AppendEncode(dst, b)
	return append(dst, '"')
}

// appendInt implements encoder.
func (j jsonEncoder) appendInt(dst []byte, key string, i int) []byte {
	return j.appendInt64(dst, key, int64(i))
}

// appendInt64 implements encoder.
func (j jsonEncoder) appendInt64(dst []byte, key string, i int64) []byte {
	dst = j.appendKeyOfPair(dst, key)
	return strconv.AppendInt(dst, i, 10)
}

// appendSeperator implements encoder.
func (j jsonEncoder) appendSeperator(dst []byte) []byte {
	return append(dst, ',')
}

// appendStart implements encoder.
func (j jsonEncoder) appendStart(dst []byte) []byte {
	return append(dst, '{')
}

// appendStr implements encoder.
func (j jsonEncoder) appendStr(dst []byte, key string, s string) []byte {
	dst = j.appendKeyOfPair(dst, key)
	return j.appendEscapedString(dst, s)
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
	return append(dst, ']')
}

// appendTime implements encoder.
func (j jsonEncoder) appendTime(dst []byte, key string, t time.Time) []byte {
	dst = j.appendKeyOfPair(dst, key)
	dst = append(dst, '"')
	dst = t.AppendFormat(dst, time.RFC3339)
	return append(dst, '"')
}

// appendUint implements encoder.
func (j jsonEncoder) appendUint(dst []byte, key string, u uint) []byte {
	return j.appendUint64(dst, key, uint64(u))
}

// appendUint64 implements encoder.
func (j jsonEncoder) appendUint64(dst []byte, key string, u uint64) []byte {
	dst = j.appendKeyOfPair(dst, key)
	return strconv.AppendUint(dst, u, 10)
}

var _ encoder = jsonEncoder{}
