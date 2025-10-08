package fnchain

import "math"

type decimalSlice struct {
	d      []byte
	nd, dp int
}

const mantBits uint = 52
const expBits uint = 11
const bias int = -1023

func ftoa(dst []byte, f float64) []byte {
	bits := math.Float64bits(f)

	neg := bits>>(expBits+mantBits) != 0
	exp := int(bits>>mantBits) & (1<<expBits - 1)
	mant := bits & (uint64(1)<<mantBits - 1)
	denorm := false

	switch exp {
	case 1<<expBits - 1:
		// Per ECMA JSON definition for numbers (https://ecma-international.org/wp-content/uploads/ECMA-404.pdf)
		// Numeric values that cannot be represented as sequences of digits
		// (such as Infinity and NaN) are not permitted.
		//
		// In this case because we cannot represent +/- Infinity or NaN in JSON we will use null
		return append(dst, "null"...)
	case 0:
		// denormalized
		denorm = true
		exp++
	default:
		mant |= uint64(1) << mantBits
	}
	exp += bias

	digs := decimalSlice{}
	buf := [32]byte{}
	digs.d = buf[:]

	dragonboxFtoa64(&digs, mant, exp-int(mantBits), denorm)

	if neg {
		dst = append(dst, '-')
	}

	if digs.dp < -3 || digs.dp >= 7 {
		return fmtE(dst, digs)
	}

	return fmtF(dst, digs, max(digs.nd-digs.dp, 0))
}

func fmtE(dst []byte, digs decimalSlice) []byte {
	ch := byte('0')
	if digs.nd != 0 {
		ch = digs.d[0]
	}
	dst = append(dst, ch)

	if digs.nd-1 > 0 {
		dst = append(dst, '.')
		i := 1
		if i < digs.nd {
			dst = append(dst, digs.d[i:digs.nd]...)
			i = digs.nd
		}
	}

	dst = append(dst, 'e')
	expe := digs.dp - 1
	if digs.nd == 0 {
		expe = 0
	}
	if expe < 0 {
		ch = '-'
		expe = -expe
	} else {
		ch = '+'
	}
	dst = append(dst, ch)

	switch {
	case expe < 10:
		dst = append(dst, '0', byte(expe)+'0')
	case expe < 100:
		dst = append(dst, byte(expe/10)+'0', byte(expe%10)+'0')
	default:
		dst = append(dst, byte(expe/100)+'0', byte(expe/10)%10+'0', byte(expe%10)+'0')
	}

	return dst
}

func fmtF(dst []byte, d decimalSlice, prec int) []byte {
	// integer, padded with zeros as needed.
	if d.dp > 0 {
		m := min(d.nd, d.dp)
		dst = append(dst, d.d[:m]...)
		for ; m < d.dp; m++ {
			dst = append(dst, '0')
		}
	} else {
		dst = append(dst, '0')
	}

	// fraction
	if prec > 0 {
		dst = append(dst, '.')
		for i := range prec {
			ch := byte('0')
			if j := d.dp + i; 0 <= j && j < d.nd {
				ch = d.d[j]
			}
			dst = append(dst, ch)
		}
	}

	return dst
}
