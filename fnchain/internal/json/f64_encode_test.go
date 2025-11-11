package json

import (
	"math"
	"math/rand/v2"
	"strconv"
	"testing"
)

func FuzzF64Encode(f *testing.F) {
	f.Add(float64(0))
	f.Add(float64(1))
	f.Add(3.14159265358979323)
	f.Add(float64(-1))
	f.Add(-1.1)

	f.Fuzz(func(t *testing.T, a float64) {
		bufa := [32]byte{}
		bufb := [32]byte{}

		outdb := ftoa(bufa[:0], a)
		outstd := strconv.AppendFloat(bufb[:0], a, 'g', -1, 64)

		dbstr := string(outdb)
		stdstr := string(outstd)

		switch stdstr {
		case "NaN", "+Inf", "-Inf":
			if dbstr != "null" {
				t.Errorf("expected null for float %s, got %s", stdstr, dbstr)
			}
		default:
			if dbstr != stdstr {
				t.Errorf("encoded floats do not match dragonbox: %s, strconv: %s", dbstr, stdstr)
			}
		}
	})
}

func benchmarkDragonbox(b *testing.B, f float64) {
	buf := make([]byte, 0, 24)

	b.ResetTimer()
	for b.Loop() {
		buf = ftoa(buf[:0], f)
	}
}

func benchmarkStdlib(b *testing.B, f float64) {
	buf := make([]byte, 0, 24)

	b.ResetTimer()
	for b.Loop() {
		buf = strconv.AppendFloat(buf[:0], f, 'g', -1, 64)
	}
}

func BenchmarkF64Encode(b *testing.B) {
	cases := map[string]float64{
		"0":   float64(0),
		"0.5": float64(0.5),
		"1":   float64(1),
		"2":   float64(2),
	}

	for range 10 {
		f := math.Float64frombits(rand.Uint64())
		cases[strconv.FormatFloat(f, 'g', -1, 64)] = f
	}

	for k, f := range cases {
		b.Run("dbx_"+k, func(b *testing.B) {
			benchmarkDragonbox(b, f)
		})
		b.Run("std_"+k, func(b *testing.B) {
			benchmarkStdlib(b, f)
		})
	}
}
