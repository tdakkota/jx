package jx

import (
	"math"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkDecoder_Int(b *testing.B) {
	data := []byte(`69315063`)
	d := GetDecoder()
	for i := 0; i < b.N; i++ {
		d.ResetBytes(data)
		if _, err := d.Int(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecoder_UInt(b *testing.B) {
	data := []byte(`69315063`)
	d := GetDecoder()
	for i := 0; i < b.N; i++ {
		d.ResetBytes(data)
		if _, err := d.UInt(); err != nil {
			b.Fatal(err)
		}
	}
}

func TestDecoderInt(t *testing.T) {
	type testCase struct {
		input string
		value int
	}
	inputs := []testCase{
		{`0`, 0},
		{`69315063`, 69315063},
	}
	for i := 1; i < math.MaxInt32; i *= 10 {
		inputs = append(inputs, testCase{
			input: strconv.Itoa(i),
			value: i,
		})
	}

	d := GetDecoder()
	for _, tt := range inputs {
		t.Run(tt.input, func(t *testing.T) {
			for _, size := range []int{32, 64} {
				d.ResetBytes([]byte(tt.input))
				v, err := d.uint(size)
				require.NoError(t, err)
				require.Equal(t, uint(tt.value), v)

				d.ResetBytes([]byte(tt.input))
				v2, err := d.int(size)
				require.NoError(t, err)
				require.Equal(t, int(tt.value), v2)
			}
		})
	}
}

func TestDecoder_Int(t *testing.T) {
	r := errReader{}
	get := func() *Decoder {
		return &Decoder{
			buf:    []byte{'1', '2'},
			tail:   2,
			reader: errReader{},
		}
	}
	t.Run("32", func(t *testing.T) {
		d := get()
		_, err := d.Int32()
		require.ErrorIs(t, err, r.Err())
	})
	t.Run("64", func(t *testing.T) {
		d := get()
		_, err := d.Int64()
		require.ErrorIs(t, err, r.Err())
	})
}
