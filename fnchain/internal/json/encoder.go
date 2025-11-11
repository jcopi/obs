package json

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"math/bits"
	"strconv"
	"time"
	"unsafe"
)

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

// escapeSingleChar is an inlineable function to append a single character to a json encoded string
// if the character requires escaping it will append the appropriate escape sequence.
// This function is intended to be inlined in the innermost tight loop, it is performance critical
func escapeSingleChar(dst []byte, b byte) []byte {
	if ei := escapeTableIdx[b]; ei < 0 {
		return append(dst, b)
	} else {
		// We know that `ei` here cannot be >= len(escapeTable) because it's a package level constant,
		// technically an unexported variable. The compiler isn't able to track that far and elide the
		// bounds check though. Here we're manually eliding the bounds check using unsafe because the
		// extra check and branch in the hot loop can be a significant performance detriment
		s := unsafe.Add(unsafe.Pointer(&escapeTable[0]), unsafe.Sizeof(escapeTable[0])*uintptr(ei))
		return append(dst, *(*string)(s)...)
	}
}

// appendStringRemainder handles escaping the remainder of an 8-byte section of a string
// this contains an unrolled loop that appends each char from the string (or it's appropriate
// escape sequence) for a substring up to 8 bytes long. This function is used in the heavily in
// escaping string to JSON representations, it is performance critical.
func appendStringRemainder(dst []byte, s string) []byte {
	switch len(s) {
	default:
		return dst
	case 1:
		return escapeSingleChar(dst, s[0])
	case 2:
		dst = escapeSingleChar(dst, s[0])
		return escapeSingleChar(dst, s[1])
	case 3:
		dst = escapeSingleChar(dst, s[0])
		dst = escapeSingleChar(dst, s[1])
		return escapeSingleChar(dst, s[2])
	case 4:
		dst = escapeSingleChar(dst, s[0])
		dst = escapeSingleChar(dst, s[1])
		dst = escapeSingleChar(dst, s[2])
		return escapeSingleChar(dst, s[3])
	case 5:
		dst = escapeSingleChar(dst, s[0])
		dst = escapeSingleChar(dst, s[1])
		dst = escapeSingleChar(dst, s[2])
		dst = escapeSingleChar(dst, s[3])
		return escapeSingleChar(dst, s[4])
	case 6:
		dst = escapeSingleChar(dst, s[0])
		dst = escapeSingleChar(dst, s[1])
		dst = escapeSingleChar(dst, s[2])
		dst = escapeSingleChar(dst, s[3])
		dst = escapeSingleChar(dst, s[4])
		return escapeSingleChar(dst, s[5])
	case 7:
		dst = escapeSingleChar(dst, s[0])
		dst = escapeSingleChar(dst, s[1])
		dst = escapeSingleChar(dst, s[2])
		dst = escapeSingleChar(dst, s[3])
		dst = escapeSingleChar(dst, s[4])
		dst = escapeSingleChar(dst, s[5])
		return escapeSingleChar(dst, s[6])
	case 8:
		dst = escapeSingleChar(dst, s[0])
		dst = escapeSingleChar(dst, s[1])
		dst = escapeSingleChar(dst, s[2])
		dst = escapeSingleChar(dst, s[3])
		dst = escapeSingleChar(dst, s[4])
		dst = escapeSingleChar(dst, s[5])
		dst = escapeSingleChar(dst, s[6])
		return escapeSingleChar(dst, s[7])
	}
}

func escapedWithMask(dst []byte, s string, mask uint64) []byte {
	// Calculate the number of leading bytes that do not require escaping
	leading := bits.TrailingZeros64(mask) / 8
	dst = append(dst, s[:leading]...)

	return appendStringRemainder(dst, s[leading:])
}

func AppendEscapedString(dst []byte, s string) []byte {
	// TODO: real implementation, this is a PoC hack
	dst = append(dst, '"')

	x := (len(s) / 8) * 8
	for i := 0; i < x; i += 8 {
		ts := unsafe.Add(unsafe.Pointer(unsafe.StringData(s)), i)
		if u := needsJSONEscape(*(*uint64)(ts)); u == 0 {
			// TODO: defer all copies of substrings that don't require escaping
			// until a substring does require a copy or the end of the string is reached
			dst = append(dst, unsafe.String((*byte)(ts), 8)...)
		} else {
			dst = escapedWithMask(dst, unsafe.String((*byte)(ts), 8), u)
		}
	}

	dst = appendStringRemainder(dst, s[x:])

	dst = append(dst, '"')
	return dst
}

// This is pulled from chromium (v8), it uses a SWAR technique to test 8 bytes of a string
// (packed as a uint64), and returns a non-zero number if any bytes need escaping
// The returned value also has the property that we can count the number of trailing zeros
// on the uint64 and that can tell us how many bytes at the start of our 8bytes string do not require escaping
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

func AppendKeyOfPair(dst []byte, key string) []byte {
	dst = AppendEscapedString(dst, key)
	return append(dst, ':')
}

// appendBool implements encoder.
func AppendBool(dst []byte, key string, b bool) []byte {
	dst = AppendKeyOfPair(dst, key)
	if b {
		return append(dst, "true,"...)
	}
	return append(dst, "false,"...)
}

// AppendB64 appends the base64 standard encoding of `b` as a json string
func AppendB64(dst []byte, key string, b []byte) []byte {
	dst = AppendKeyOfPair(dst, key)
	dst = append(dst, '"')
	dst = base64.StdEncoding.AppendEncode(dst, b)
	return append(dst, '"', ',')
}

// appendDur implements encoder.
func AppendDur(dst []byte, key string, d time.Duration) []byte {
	return AppendFloat64(dst, key, d.Seconds())
}

// appendEnd implements encoder.
func AppendEnd(dst []byte) []byte {
	if dst[len(dst)-1] == ',' {
		dst = dst[:len(dst)-1]
	}
	return append(dst, '}', '\n')
}

// appendErr implements encoder.
func AppendErr(dst []byte, key string, e error) []byte {
	return AppendStr(dst, key, e.Error())
}

// appendFloat32 implements encoder.
func AppendFloat32(dst []byte, key string, f float32) []byte {
	return AppendFloat64(dst, key, float64(f))
}

// appendFloat64 implements encoder.
func AppendFloat64(dst []byte, key string, f float64) []byte {
	dst = AppendKeyOfPair(dst, key)
	dst = ftoa(dst, f)
	return append(dst, ',')
}

// appendHex implements encoder.
func AppendHex(dst []byte, key string, b []byte) []byte {
	dst = AppendKeyOfPair(dst, key)
	dst = append(dst, '"')
	dst = hex.AppendEncode(dst, b)
	return append(dst, '"', ',')
}

// appendInt implements encoder.
func AppendInt(dst []byte, key string, i int) []byte {
	return AppendInt64(dst, key, int64(i))
}

// appendInt64 implements encoder.
func AppendInt64(dst []byte, key string, i int64) []byte {
	dst = AppendKeyOfPair(dst, key)
	dst = strconv.AppendInt(dst, i, 10)
	return append(dst, ',')
}

// appendStart implements encoder.
func AppendStart(dst []byte) []byte {
	return append(dst, '{')
}

// appendStr implements encoder.
func AppendStr(dst []byte, key string, s string) []byte {
	dst = AppendKeyOfPair(dst, key)
	dst = AppendEscapedString(dst, s)
	return append(dst, ',')
}

// appendStrs implements encoder.
func AppendStrs(dst []byte, key string, strs []string) []byte {
	dst = AppendKeyOfPair(dst, key)
	dst = append(dst, '[')
	for i, s := range strs {
		dst = AppendEscapedString(dst, s)
		if i < len(strs)-1 {
			dst = append(dst, ',')
		}
	}
	return append(dst, ']', ',')
}

// appendTime implements encoder.
func AppendTime(dst []byte, key string, t time.Time) []byte {
	dst = AppendKeyOfPair(dst, key)
	dst = append(dst, '"')
	dst = t.AppendFormat(dst, time.RFC3339)
	return append(dst, '"', ',')
}

// appendUint implements encoder.
func AppendUint(dst []byte, key string, u uint) []byte {
	return AppendUint64(dst, key, uint64(u))
}

// appendUint64 implements encoder.
func AppendUint64(dst []byte, key string, u uint64) []byte {
	dst = AppendKeyOfPair(dst, key)
	dst = strconv.AppendUint(dst, u, 10)
	return append(dst, ',')
}

// For fields that will have a constant string key and are in the extreme hot path
// (used in every event write) have some keys defined that skip unnecessary string
// escaping. The strings in these fields are known to be valid JSON strings w/o escaping

// append a const key that is known to not require any escaping
func AppendKnownKeyOfPair(dst []byte, knownKey string) []byte {
	dst = append(dst, '"')
	dst = append(dst, knownKey...)
	return append(dst, '"', ':')
}

func AppendKnownKeyID(dst []byte, knownKey string, id uint64) []byte {
	// will be common to all events, it makes sense to define as efficient of
	// an encoding as possible. Currently the encoding is not maximally efficient
	b := [8]byte{}
	binary.LittleEndian.AppendUint64(b[:0], id)
	dst = AppendKnownKeyOfPair(dst, knownKey)
	dst = append(dst, '"')
	dst = hex.AppendEncode(dst, b[:])
	return append(dst, '"', ',')
}

func AppendKnownKeyStr(dst []byte, knownKey string, s string) []byte {
	dst = AppendKnownKeyOfPair(dst, knownKey)
	dst = AppendEscapedString(dst, s)
	return append(dst, ',')
}

func AppendKnownKeyError(dst []byte, knownKey string, e error) []byte {
	return AppendKnownKeyStr(dst, knownKey, e.Error())
}

type Value interface {
	Append(dst []byte) []byte
}

func AppendKnownKeyValue[T Value](dst []byte, knownKey string, v T) []byte {
	dst = AppendKnownKeyOfPair(dst, knownKey)
	dst = v.Append(dst)
	return append(dst, ',')
}
