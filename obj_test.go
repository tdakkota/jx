package jx

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_empty_object(t *testing.T) {
	iter := DecodeString(`{}`)
	require.NoError(t, iter.Object(func(iter *Decoder, field string) error {
		t.Error("should not call")
		return nil
	}))
}

func Test_one_field(t *testing.T) {
	should := require.New(t)
	iter := DecodeString(`{"a": "stream"}`)
	should.NoError(iter.Object(func(iter *Decoder, field string) error {
		should.Equal("a", field)
		return iter.Skip()
	}))
}

func Test_write_object(t *testing.T) {
	should := require.New(t)
	buf := &bytes.Buffer{}
	s := NewEncoder(buf, 4096)
	s.SetIdent(2)
	s.ObjStart()
	s.ObjField("hello")
	s.Int(1)
	s.More()
	s.ObjField("world")
	s.Int(2)
	s.ObjEnd()
	should.NoError(s.Flush())
	should.Equal("{\n  \"hello\": 1,\n  \"world\": 2\n}", buf.String())
}
