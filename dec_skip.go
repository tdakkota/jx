package jx

import (
	"io"
	"math/bits"

	"github.com/go-faster/errors"
)

func (d *Decoder) readExact4(b *[4]byte) error {
	if buf := d.buf[d.head:d.tail]; len(buf) >= len(b) {
		d.head += copy(b[:], buf[:4])
		return nil
	}

	n := copy(b[:], d.buf[d.head:d.tail])
	if err := d.readAtLeast(len(b) - n); err != nil {
		return err
	}
	d.head += copy(b[n:], d.buf[d.head:d.tail])
	return nil
}

func findInvalidToken4(buf [4]byte, mask uint32) error {
	c := uint32(buf[0]) | uint32(buf[1])<<8 | uint32(buf[2])<<16 | uint32(buf[3])<<24
	idx := bits.TrailingZeros32(c^mask) / 8
	return badToken(buf[idx])
}

// Null reads a json object as null and
// returns whether it's a null or not.
func (d *Decoder) Null() error {
	var buf [4]byte
	if err := d.readExact4(&buf); err != nil {
		return err
	}

	if string(buf[:]) != "null" {
		const encodedNull = 'n' | 'u'<<8 | 'l'<<16 | 'l'<<24
		return findInvalidToken4(buf, encodedNull)
	}
	return nil
}

// Bool reads a json object as Bool
func (d *Decoder) Bool() (bool, error) {
	var buf [4]byte
	if err := d.readExact4(&buf); err != nil {
		return false, err
	}

	switch string(buf[:]) {
	case "true":
		return true, nil
	case "fals":
		if err := d.consume('e'); err != nil {
			return false, err
		}
		return false, nil
	default:
		switch c := buf[0]; c {
		case 't':
			const encodedTrue = 't' | 'r'<<8 | 'u'<<16 | 'e'<<24
			return false, findInvalidToken4(buf, encodedTrue)
		case 'f':
			const encodedAlse = 'a' | 'l'<<8 | 's'<<16 | 'e'<<24
			return false, findInvalidToken4(buf, encodedAlse)
		default:
			return false, badToken(c)
		}
	}
}

// Skip skips a json object and positions to relatively the next json object.
func (d *Decoder) Skip() error {
	var (
		stack       = make([]byte, 0, 128)
		unwindParam byte
	)

	push := func(c byte) error {
		if err := d.incDepth(); err != nil {
			return err
		}
		stack = append(stack, c)
		return nil
	}
	pop := func(c byte) error {
		if len(stack) < 1 {
			return badToken(c)
		}
		last := stack[len(stack)-1]
		if last != c {
			return badToken(c)
		}
		stack = stack[:len(stack)-1]
		return d.decDepth()
	}

	c, err := d.next()
	if err != nil {
		return err
	}
	switch c {
	case '[':
		goto skipArray
	case '{':
		goto skipObject
	case '"':
		return d.skipStr()
	case 'n':
		d.unread()
		return d.Null()
	case 't', 'f':
		d.unread()
		_, err := d.Bool()
		return err
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		d.unread()
		return d.skipNumber()
	default:
		return badToken(c)
	}

skipObject:
	if err := push('{'); err != nil {
		return err
	}
	c, err = d.more()
	if err != nil {
		return errors.Wrap(err, "next")
	}
	switch c {
	case '}':
		unwindParam = c
		goto unwind
	case '"':
		d.unread()
	default:
		return badToken(c)
	}

readObject:
	if err := d.consume('"'); err != nil {
		return err
	}
	if err := d.skipStr(); err != nil {
		return errors.Wrap(err, "read field name")
	}
	if err := d.consume(':'); err != nil {
		return errors.Wrap(err, "field")
	}
	{
		c, err := d.next()
		if err != nil {
			return err
		}
		switch c {
		case '[':
			goto skipArray
		case '{':
			goto skipObject
		case '"':
			if err := d.skipStr(); err != nil {
				return errors.Wrap(err, "str")
			}
		case 'n':
			d.unread()
			if err := d.Null(); err != nil {
				return errors.Wrap(err, "null")
			}
		case 't', 'f':
			d.unread()
			if _, err := d.Bool(); err != nil {
				return errors.Wrap(err, "bool")
			}
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			d.unread()
			if err := d.skipNumber(); err != nil {
				return errors.Wrap(err, "number")
			}
		default:
			return badToken(c)
		}
	}
unwindObject:
	c, err = d.more()
	if err != nil {
		return errors.Wrap(err, "read comma")
	}
	switch c {
	case ',':
	case '}':
		unwindParam = c
		goto unwind
	default:
		return badToken(c)
	}
	goto readObject

skipArray:
	if err := push('['); err != nil {
		return err
	}
	c, err = d.more()
	if err != nil {
		return errors.Wrap(err, "next")
	}
	if c == ']' {
		unwindParam = c
		goto unwind
	}
	d.unread()
readArray:
	{
		c, err := d.next()
		if err != nil {
			return err
		}
		switch c {
		case '[':
			goto skipArray
		case '{':
			goto skipObject
		case '"':
			if err := d.skipStr(); err != nil {
				return errors.Wrap(err, "str")
			}
		case 'n':
			d.unread()
			if err := d.Null(); err != nil {
				return errors.Wrap(err, "null")
			}
		case 't', 'f':
			d.unread()
			if _, err := d.Bool(); err != nil {
				return errors.Wrap(err, "bool")
			}
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			d.unread()
			if err := d.skipNumber(); err != nil {
				return errors.Wrap(err, "number")
			}
		default:
			return badToken(c)
		}
	}
unwindArray:
	c, err = d.more()
	if err != nil {
		return errors.Wrap(err, "read comma")
	}
	switch c {
	case ',':
	case ']':
		unwindParam = c
		goto unwind
	default:
		return badToken(c)
	}
	goto readArray

unwind:
	if err := pop(unwindParam - 2); err != nil {
		return err
	}
	if len(stack) > 0 {
		switch stack[len(stack)-1] {
		case '{':
			goto unwindObject
		case '[':
			goto unwindArray
		}
	}
	return nil
}

var (
	skipNumberSet = [256]byte{
		'0': 1,
		'1': 1,
		'2': 1,
		'3': 1,
		'4': 1,
		'5': 1,
		'6': 1,
		'7': 1,
		'8': 1,
		'9': 1,

		',':  2,
		']':  2,
		'}':  2,
		' ':  2,
		'\t': 2,
		'\n': 2,
		'\r': 2,
	}
)

// skipNumber reads one JSON number.
//
// Assumes d.buf is not empty.
func (d *Decoder) skipNumber() error {
	const (
		digitTag  byte = 1
		closerTag byte = 2
	)
	c := d.buf[d.head]
	d.head++
	switch c {
	case '-':
		c, err := d.byte()
		if err != nil {
			return err
		}
		// Character after '-' must be a digit.
		if skipNumberSet[c] != digitTag {
			return badToken(c)
		}
		if c != '0' {
			break
		}
		fallthrough
	case '0':
		// If buffer is empty, try to read more.
		if d.head == d.tail {
			err := d.read()
			if err != nil {
				// There is no data anymore.
				if err == io.EOF {
					return nil
				}
				return err
			}
		}

		c = d.buf[d.head]
		if skipNumberSet[c] == closerTag {
			return nil
		}
		switch c {
		case '.':
			goto stateDot
		case 'e', 'E':
			goto stateExp
		default:
			return badToken(c)
		}
	}
	for {
		for i, c := range d.buf[d.head:d.tail] {
			switch skipNumberSet[c] {
			case closerTag:
				d.head += i
				return nil
			case digitTag:
				continue
			}

			switch c {
			case '.':
				d.head += i
				goto stateDot
			case 'e', 'E':
				d.head += i
				goto stateExp
			default:
				return badToken(c)
			}
		}

		if err := d.read(); err != nil {
			// There is no data anymore.
			if err == io.EOF {
				d.head = d.tail
				return nil
			}
			return err
		}
	}

stateDot:
	d.head++
	{
		var last byte = '.'
		for {
			for i, c := range d.buf[d.head:d.tail] {
				switch skipNumberSet[c] {
				case closerTag:
					d.head += i
					// Check that dot is not last character.
					if last == '.' {
						return io.ErrUnexpectedEOF
					}
					return nil
				case digitTag:
					last = c
					continue
				}

				switch c {
				case 'e', 'E':
					if last == '.' {
						return badToken(c)
					}
					d.head += i
					goto stateExp
				default:
					return badToken(c)
				}
			}

			if err := d.read(); err != nil {
				// There is no data anymore.
				if err == io.EOF {
					d.head = d.tail
					// Check that dot is not last character.
					if last == '.' {
						return io.ErrUnexpectedEOF
					}
					return nil
				}
				return err
			}
		}
	}
stateExp:
	d.head++
	// There must be a number or sign after e.
	{
		numOrSign, err := d.byte()
		if err != nil {
			return err
		}
		if skipNumberSet[numOrSign] != digitTag { // If next character is not a digit, check for sign.
			if numOrSign == '-' || numOrSign == '+' {
				num, err := d.byte()
				if err != nil {
					return err
				}
				// There must be a number after sign.
				if skipNumberSet[num] != digitTag {
					return badToken(num)
				}
			} else {
				return badToken(numOrSign)
			}
		}
	}
	for {
		for i, c := range d.buf[d.head:d.tail] {
			if skipNumberSet[c] == closerTag {
				d.head += i
				return nil
			}
			if skipNumberSet[c] == 0 {
				return badToken(c)
			}
		}

		if err := d.read(); err != nil {
			// There is no data anymore.
			if err == io.EOF {
				d.head = d.tail
				return nil
			}
			return err
		}
	}
}

var (
	escapedStrSet = [256]byte{
		'"': 1, '\\': 1, '/': 1, 'b': 1, 'f': 1, 'n': 1, 'r': 1, 't': 1,
		'u': 2,
	}
	hexSet = [256]byte{
		'0': 1, '1': 1, '2': 1, '3': 1,
		'4': 1, '5': 1, '6': 1, '7': 1,
		'8': 1, '9': 1,

		'A': 1, 'B': 1, 'C': 1, 'D': 1,
		'E': 1, 'F': 1,

		'a': 1, 'b': 1, 'c': 1, 'd': 1,
		'e': 1, 'f': 1,
	}
)

// skipStr reads one JSON string.
//
// Assumes first quote was consumed.
func (d *Decoder) skipStr() error {
readStr:
	for {
		for i, c := range d.buf[d.head:d.tail] {
			switch {
			case c == '"':
				d.head += i + 1
				return nil
			case c == '\\':
				d.head += i + 1
				goto readEscaped
			case c < ' ':
				return badToken(c)
			}
		}

		if err := d.read(); err != nil {
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			return err
		}
	}

readEscaped:
	v, err := d.byte()
	if err != nil {
		return err
	}
	switch escapedStrSet[v] {
	case 1:
	case 2:
		for i := 0; i < 4; i++ {
			h, err := d.byte()
			if err != nil {
				return err
			}
			if hexSet[h] == 0 {
				return badToken(h)
			}
		}
	default:
		return badToken(v)
	}
	goto readStr
}

// skipObj reads JSON object.
//
// Assumes first bracket was consumed.
func (d *Decoder) skipObj() error {
	if err := d.incDepth(); err != nil {
		return errors.Wrap(err, "inc")
	}

	c, err := d.more()
	if err != nil {
		return errors.Wrap(err, "next")
	}
	switch c {
	case '}':
		return d.decDepth()
	case '"':
		d.unread()
	default:
		return badToken(c)
	}

	for {
		if err := d.consume('"'); err != nil {
			return err
		}
		if err := d.skipStr(); err != nil {
			return errors.Wrap(err, "read field name")
		}
		if err := d.consume(':'); err != nil {
			return errors.Wrap(err, "field")
		}
		if err := d.Skip(); err != nil {
			return err
		}
		c, err := d.more()
		if err != nil {
			return errors.Wrap(err, "read comma")
		}
		switch c {
		case ',':
			continue
		case '}':
			return d.decDepth()
		default:
			return badToken(c)
		}
	}
}

// skipArr reads JSON array.
//
// Assumes first bracket was consumed.
func (d *Decoder) skipArr() error {
	if err := d.incDepth(); err != nil {
		return errors.Wrap(err, "inc")
	}

	c, err := d.more()
	if err != nil {
		return errors.Wrap(err, "next")
	}
	if c == ']' {
		return d.decDepth()
	}
	d.unread()

	for {
		if err := d.Skip(); err != nil {
			return err
		}
		c, err := d.more()
		if err != nil {
			return errors.Wrap(err, "read comma")
		}
		switch c {
		case ',':
			continue
		case ']':
			return d.decDepth()
		default:
			return badToken(c)
		}
	}
}
