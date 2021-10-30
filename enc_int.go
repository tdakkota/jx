package jx

var digits []uint32

func init() {
	digits = make([]uint32, 1000)
	for i := uint32(0); i < 1000; i++ {
		digits[i] = (((i / 100) + '0') << 16) + ((((i / 10) % 10) + '0') << 8) + i%10 + '0'
		if i < 10 {
			digits[i] += 2 << 24
		} else if i < 100 {
			digits[i] += 1 << 24
		}
	}
}

func writeFirstBuf(space []byte, v uint32) []byte {
	start := v >> 24
	if start == 0 {
		space = append(space, byte(v>>16), byte(v>>8))
	} else if start == 1 {
		space = append(space, byte(v>>8))
	}
	space = append(space, byte(v))
	return space
}

func writeBuf(buf []byte, v uint32) []byte {
	return append(buf, byte(v>>16), byte(v>>8), byte(v))
}

// Uint32 writes uint32 to stream.
func (e *Encoder) Uint32(val uint32) {
	q1 := val / 1000
	if q1 == 0 {
		e.buf = writeFirstBuf(e.buf, digits[val])
		return
	}
	r1 := val - q1*1000
	q2 := q1 / 1000
	if q2 == 0 {
		e.buf = writeFirstBuf(e.buf, digits[q1])
		e.buf = writeBuf(e.buf, digits[r1])
		return
	}
	r2 := q1 - q2*1000
	q3 := q2 / 1000
	if q3 == 0 {
		e.buf = writeFirstBuf(e.buf, digits[q2])
	} else {
		r3 := q2 - q3*1000
		e.buf = append(e.buf, byte(q3+'0'))
		e.buf = writeBuf(e.buf, digits[r3])
	}
	e.buf = writeBuf(e.buf, digits[r2])
	e.buf = writeBuf(e.buf, digits[r1])
}

// Int32 writes int32 to stream.
func (e *Encoder) Int32(nval int32) {
	var val uint32
	if nval < 0 {
		val = uint32(-nval)
		e.buf = append(e.buf, '-')
	} else {
		val = uint32(nval)
	}
	e.Uint32(val)
}

// Uint64 writes uint64 to stream.
func (e *Encoder) Uint64(val uint64) {
	q1 := val / 1000
	if q1 == 0 {
		e.buf = writeFirstBuf(e.buf, digits[val])
		return
	}
	r1 := val - q1*1000
	q2 := q1 / 1000
	if q2 == 0 {
		e.buf = writeFirstBuf(e.buf, digits[q1])
		e.buf = writeBuf(e.buf, digits[r1])
		return
	}
	r2 := q1 - q2*1000
	q3 := q2 / 1000
	if q3 == 0 {
		e.buf = writeFirstBuf(e.buf, digits[q2])
		e.buf = writeBuf(e.buf, digits[r2])
		e.buf = writeBuf(e.buf, digits[r1])
		return
	}
	r3 := q2 - q3*1000
	q4 := q3 / 1000
	if q4 == 0 {
		e.buf = writeFirstBuf(e.buf, digits[q3])
		e.buf = writeBuf(e.buf, digits[r3])
		e.buf = writeBuf(e.buf, digits[r2])
		e.buf = writeBuf(e.buf, digits[r1])
		return
	}
	r4 := q3 - q4*1000
	q5 := q4 / 1000
	if q5 == 0 {
		e.buf = writeFirstBuf(e.buf, digits[q4])
		e.buf = writeBuf(e.buf, digits[r4])
		e.buf = writeBuf(e.buf, digits[r3])
		e.buf = writeBuf(e.buf, digits[r2])
		e.buf = writeBuf(e.buf, digits[r1])
		return
	}
	r5 := q4 - q5*1000
	q6 := q5 / 1000
	if q6 == 0 {
		e.buf = writeFirstBuf(e.buf, digits[q5])
	} else {
		e.buf = writeFirstBuf(e.buf, digits[q6])
		r6 := q5 - q6*1000
		e.buf = writeBuf(e.buf, digits[r6])
	}
	e.buf = writeBuf(e.buf, digits[r5])
	e.buf = writeBuf(e.buf, digits[r4])
	e.buf = writeBuf(e.buf, digits[r3])
	e.buf = writeBuf(e.buf, digits[r2])
	e.buf = writeBuf(e.buf, digits[r1])
}

// Int64 writes int64 to stream
func (e *Encoder) Int64(nval int64) {
	var val uint64
	if nval < 0 {
		val = uint64(-nval)
		e.buf = append(e.buf, '-')
	} else {
		val = uint64(nval)
	}
	e.Uint64(val)
}

// Int writes int to stream.
func (e *Encoder) Int(val int) {
	e.Int64(int64(val))
}

// Uint writes uint to stream.
func (e *Encoder) Uint(val uint) {
	e.Uint64(uint64(val))
}