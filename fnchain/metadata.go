package fnchain

import "time"

type Metadata[T any] interface {
	Bool(key string, b bool) T
	Int(key string, i int) T
	Int64(key string, i int64) T
	Uint(key string, u uint) T
	Uint64(key string, u uint64) T
	Float32(key string, f float32) T
	Float64(key string, f float64) T
	Str(key string, s string) T
	Strs(key string, strs []string) T
	Time(key string, t time.Time) T
	Dur(key string, d time.Duration) T
	Hex(key string, b []byte) T

	// Some well-known fields which have automatic keys
	Err(e error) T
	Msg(msg string) T
}

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

func AppendLevel(dst []byte, lvl Level) []byte {
	// This should be efficient because the strings in LevelString are constants
	// there shouldn't be any additional allocations from this approach
	return append(dst, LevelString(lvl)...)
}

type EvtType uint32

const (
	GenericEvent EvtType = iota
	MethodEvent
)
