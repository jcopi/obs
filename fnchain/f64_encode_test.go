package fnchain

import (
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
