package jx

import (
	"math"
	"testing"
)

func BenchmarkWriter_UInt32(b *testing.B) {
	w := Writer{Buf: make([]byte, 0, 32)}
	w.UInt32(math.MaxUint32)

	b.SetBytes(int64(len(w.Buf)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w.Buf = w.Buf[:0]
		w.UInt32(math.MaxUint32)
	}
}

func BenchmarkWriter_UInt64(b *testing.B) {
	w := Writer{Buf: make([]byte, 0, 32)}
	w.UInt64(math.MaxUint64)

	b.SetBytes(int64(len(w.Buf)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w.Buf = w.Buf[:0]
		w.UInt64(math.MaxUint64)
	}
}
