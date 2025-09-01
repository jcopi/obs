package fnarg

import (
	"encoding/json"
	"strconv"
	"time"
)

type Level int8

const (
	// NoLevel defines an absent level
	NoLevel Level = iota
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel

	// Disabled level should always result in a no-op
	Disabled Level = -1
)

func LevelString(lvl Level) string {
	switch lvl {
	case DebugLevel:
		return "debug"
	case Disabled:
		return "disabled"
	case ErrorLevel:
		return "error"
	case InfoLevel:
		return "info"
	case NoLevel:
		return "none"
	case WarnLevel:
		return "warn"
	default:
		return "unknown"
	}
}

type EvtType uint32

const (
	GenericEvent EvtType = iota
	MethodEvent
)

type Metadata func(dst []byte) []byte

func encodeString(str string, dst []byte) []byte {
	// TODO: borrow the SWAR technique for fast json serialization (escaping)
	//       of small strings from v8.
	// https://source.chromium.org/chromium/chromium/src/+/main:v8/src/json/json-stringifier.cc;drc=1645281bbd1b183a252835d376166bd210135bbe;l=3353

	// The current implementation is a horribly inefficient PoC, it must be changed before real usage
	out, err := json.Marshal(str)
	if err != nil {
		return append(dst, '"', '?', '"')
	}
	return append(dst, out...)
}

func Bool(key string, b bool) Metadata {
	return Metadata(func(dst []byte) []byte {
		dst = encodeString(key, dst)
		dst = append(dst, ':')
		if b {
			dst = append(dst, 't', 'r', 'u', 'e')
		} else {
			dst = append(dst, 'f', 'a', 'l', 's', 'e')
		}

		return dst
	})
}

func Dur(key string, d time.Duration) Metadata {
	// TODO: implement an efficient duration encoder
	//       time.Duration.String is probably efficient,
	//       but the round-trip through a string can be avoided
	return Str(key, d.String())
}

func Err(e error) Metadata {
	return Str("err", e.Error())
}

func Float32(key string, f float32) Metadata {
	return Float64(key, float64(f))
}

func Float64(key string, f float64) Metadata {
	return Metadata(func(dst []byte) []byte {
		dst = encodeString(key, dst)
		dst = append(dst, ':')
		dst = append(dst, strconv.FormatFloat(f, 'g', -1, 64)...)
		return dst
	})
}

// Hex implements Event.
func Hex(key string, b []byte) Metadata {
	panic("unimplemented")
}

// Int implements Event.
func Int(key string, i int) Metadata {
	return Int64(key, int64(i))
}

// Int64 implements Event.
func Int64(key string, i int64) Metadata {
	return Metadata(func(dst []byte) []byte {
		dst = encodeString(key, dst)
		dst = append(dst, ':')
		dst = append(dst, strconv.FormatInt(i, 10)...)
		return dst
	})
}

// Msg implements Event.
func Msg(msg string) Metadata {
	return Str("msg", msg)
}

func Str(key string, s string) Metadata {
	return Metadata(func(dst []byte) []byte {
		dst = encodeString(key, dst)
		dst = append(dst, ':')
		dst = encodeString(s, dst)
		return dst
	})
}

// Strs implements Event.
func Strs(key string, strs []string) Metadata {
	return Metadata(func(dst []byte) []byte {
		dst = encodeString(key, dst)
		dst = append(dst, ':')
		dst = append(dst, '[')
		for i := range strs {
			dst = encodeString(strs[i], dst)
			dst = append(dst, ',')
		}
		dst[len(dst)-1] = ']'
		return dst
	})
}

// Time implements Event.
func Time(key string, t time.Time) Metadata {
	return Str(key, t.Format(time.RFC3339))
}

// Uint implements Event.
func Uint(key string, u uint) Metadata {
	return Uint64(key, uint64(u))
}

// Uint64 implements Event.
func Uint64(key string, u uint64) Metadata {
	return Metadata(func(dst []byte) []byte {
		dst = encodeString(key, dst)
		dst = append(dst, ':')
		dst = append(dst, strconv.FormatUint(u, 10)...)
		return dst
	})
}
