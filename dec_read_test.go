package jx

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecoder_readAtLeast(t *testing.T) {
	a := require.New(t)
	d := Decode(strings.NewReader("aboba"), 1)
	a.NoError(d.readAtLeast(4))
	a.Equal(d.buf[d.head:d.tail], []byte("abob"))
}

func TestDecoder_consume(t *testing.T) {
	r := errReader{}
	d := Decode(r, 1)
	require.ErrorIs(t, d.consume('"'), r.Err())
}

func BenchmarkDecoder_next(b *testing.B) {
	bench := func(input []byte) func(b *testing.B) {
		return func(b *testing.B) {
			d := DecodeBytes(input)

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				d.ResetBytes(input)
				if c, err := d.next(); err != nil {
					b.Fatal(c, err)
				}
			}
		}
	}

	for _, tc := range []struct {
		name  string
		input string
	}{
		{"NoSpace", `{"x": 0}`},
		{"Space", ` {"x": 0}`},
		{"6Newlines", "\n\n\n\n\n\n" + ` {"x": 0}`},
	} {
		b.Run(tc.name, bench([]byte(tc.input)))
	}
}
