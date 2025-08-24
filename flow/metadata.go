package flow

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
