package fnarg

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMetadataEncoders(t *testing.T) {
	cases := []struct {
		expected any
		name     string
		metadata []Metadata
	}{
		{
			name: "happy path",
			metadata: []Metadata{
				Bool("Bool", true),
				Dur("Dur", time.Minute),
				Err(errors.New("error")),
				Float64("Float", 2.01),
				Int64("Int", -42),
				Uint64("Uint", 42),
				Str("Str", "string"),
				Time("Time", time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
			},
			expected: struct {
				Time  time.Time
				Err   string `json:"err"`
				Str   string
				Dur   time.Duration
				Float float64
				Int   int64
				Uint  uint64
				Bool  bool
			}{
				Bool: true, Dur: time.Minute, Err: errors.New("error").Error(), Float: 2.01, Int: -42, Uint: 42, Str: "string", Time: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := Ctx{buf: []byte{'{'}}
			ctx = ctx.With(tc.metadata...)
			ctx.buf = append(ctx.buf, '}', '\n')

			expect, err := json.Marshal(tc.expected)
			require.NoError(t, err)

			require.JSONEq(t, string(expect), string(ctx.buf))
		})
	}
}
