package test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	j "github.com/ogen-go/json"
)

func Test_read_string(t *testing.T) {
	badInputs := []string{
		``,
		`"`,
		`"\"`,
		`"\\\"`,
		"\"\n\"",
		`"\U0001f64f"`,
		`"\uD83D\u00"`,
	}
	for i := 0; i < 32; i++ {
		// control characters are invalid
		badInputs = append(badInputs, string([]byte{'"', byte(i), '"'}))
	}

	for _, input := range badInputs {
		testReadString(t, input, "", true, "json.Unmarshal", json.Unmarshal)
		testReadString(t, input, "", true, "jsoniter.Unmarshal", json.Unmarshal)

		i := j.ConfigCompat.BorrowIterator([]byte(input))
		i.ReadString()
		assert.Error(t, i.Error)
		assert.False(t, j.ConfigCompat.Valid([]byte(input)))

		j.ConfigCompat.ReturnIterator(i)
	}

	goodInputs := []struct {
		input       string
		expectValue string
	}{
		{`""`, ""},
		{`"a"`, "a"},
		{`null`, ""},
		{`"Iñtërnâtiônàlizætiøn,💝🐹🌇⛔"`, "Iñtërnâtiônàlizætiøn,💝🐹🌇⛔"},
		{`"\uD83D"`, string([]byte{239, 191, 189})},
		{`"\uD83D\\"`, string([]byte{239, 191, 189, '\\'})},
		{`"\uD83D\ub000"`, string([]byte{239, 191, 189, 235, 128, 128})},
		{`"\uD83D\ude04"`, "😄"},
		{`"\uDEADBEEF"`, string([]byte{239, 191, 189, 66, 69, 69, 70})},
		{`"hel\"lo"`, `hel"lo`},
		{`"hel\\\/lo"`, `hel\/lo`},
		{`"hel\\blo"`, `hel\blo`},
		{`"hel\\\blo"`, "hel\\\blo"},
		{`"hel\\nlo"`, `hel\nlo`},
		{`"hel\\\nlo"`, "hel\\\nlo"},
		{`"hel\\tlo"`, `hel\tlo`},
		{`"hel\\flo"`, `hel\flo`},
		{`"hel\\\flo"`, "hel\\\flo"},
		{`"hel\\\rlo"`, "hel\\\rlo"},
		{`"hel\\\tlo"`, "hel\\\tlo"},
		{`"\u4e2d\u6587"`, "中文"},
		{`"\ud83d\udc4a"`, "\xf0\x9f\x91\x8a"},
	}

	for _, tc := range goodInputs {
		testReadString(t, tc.input, tc.expectValue, false, "json.Unmarshal", json.Unmarshal)
		testReadString(t, tc.input, tc.expectValue, false, "jsoniter.Unmarshal", json.Unmarshal)

		i := j.ConfigCompat.BorrowIterator([]byte(tc.input))
		s := i.ReadString()
		assert.NoError(t, i.Error)
		assert.Equal(t, tc.expectValue, s)
		j.ConfigCompat.ReturnIterator(i)
	}
}

func testReadString(t *testing.T, input string, expectValue string, expectError bool, marshalerName string, marshaler func([]byte, interface{}) error) {
	var value string
	err := marshaler([]byte(input), &value)
	if expectError != (err != nil) {
		t.Errorf("%q: %s: expected error %v, got %v", input, marshalerName, expectError, err)
		return
	}
	if value != expectValue {
		t.Errorf("%q: %s: expected %q, got %q", input, marshalerName, expectValue, value)
		return
	}
}
