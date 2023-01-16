package jx

import (
	"io"
	"math"
	"strconv"

	"github.com/go-faster/errors"
)

var (
	intDigits   [256]int8
	errOverflow = strconv.ErrRange
)

const (
	uint8SafeToMultiply10  = uint8(0xff)/10 - 1
	uint16SafeToMultiply10 = uint16(0xffff)/10 - 1
	uint32SafeToMultiply10 = uint32(0xffffffff)/10 - 1
	uint64SafeToMultiple10 = uint64(0xffffffffffffffff)/10 - 1
)

func init() {
	for i := 0; i < len(intDigits); i++ {
		intDigits[i] = invalidCharForNumber
	}
	for i := int8('0'); i <= int8('9'); i++ {
		intDigits[i] = i - int8('0')
	}
}

func (d *Decoder) int(size int) (int, error) {
	switch size {
	case 8:
		v, err := d.Int8()
		return int(v), err
	case 16:
		v, err := d.Int16()
		return int(v), err
	case 32:
		v, err := d.Int32()
		return int(v), err
	default:
		v, err := d.Int64()
		return int(v), err
	}
}

// Int reads int.
func (d *Decoder) Int() (int, error) {
	return d.int(strconv.IntSize)
}

func (d *Decoder) uint(size int) (uint, error) {
	switch size {
	case 8:
		v, err := d.UInt8()
		return uint(v), err
	case 16:
		v, err := d.UInt16()
		return uint(v), err
	case 32:
		v, err := d.UInt32()
		return uint(v), err
	default:
		v, err := d.UInt64()
		return uint(v), err
	}
}

// UInt reads uint.
func (d *Decoder) UInt() (uint, error) {
	return d.uint(strconv.IntSize)
}

// Int8 reads int8.
func (d *Decoder) Int8() (int8, error) {
	c, err := d.byte()
	if err != nil {
		return 0, err
	}
	if c == '-' {
		val, err := d.readUInt8()
		if err != nil {
			return 0, err
		}
		if val > math.MaxInt8+1 {
			return 0, errOverflow
		}
		return -int8(val), nil
	}
	d.unread()
	val, err := d.readUInt8()
	if err != nil {
		return 0, err
	}
	if val > math.MaxInt8 {
		return 0, errOverflow
	}
	return int8(val), nil
}

// UInt8 reads uint8.
func (d *Decoder) UInt8() (uint8, error) {
	return d.readUInt8()
}

func (d *Decoder) readUInt8() (uint8, error) {
	c, err := d.byte()
	if err != nil {
		return 0, err
	}
	ind := intDigits[c]
	if ind == 0 {
		return 0, nil
	}
	if ind == invalidCharForNumber {
		return 0, errors.Wrap(err, "bad token")
	}
	value := uint8(ind)
	if d.tail-d.head > 10 {
		i := d.head
		ind2 := intDigits[d.buf[i]]
		if ind2 == invalidCharForNumber {
			d.head = i
			return value, nil
		}
		i++
		ind3 := intDigits[d.buf[i]]
		if ind3 == invalidCharForNumber {
			d.head = i
			return value*10 + uint8(ind2), nil
		}
		i++
		ind4 := intDigits[d.buf[i]]
		value = value*100 + uint8(ind2)*10 + uint8(ind3)
		d.head = i
		if ind4 == invalidCharForNumber {
			return value, nil
		}
	}
	for {
		buf := d.buf[d.head:d.tail]
		for i, c := range buf {
			ind = intDigits[c]
			if ind == invalidCharForNumber {
				d.head += i
				return value, nil
			}
			if value > uint8SafeToMultiply10 {
				value2 := (value << 3) + (value << 1) + uint8(ind)
				if value2 < value {
					return 0, errOverflow
				}
				value = value2
				continue
			}
			value = (value << 3) + (value << 1) + uint8(ind)
		}
		err := d.read()
		if err == io.EOF {
			return value, nil
		}
		if err != nil {
			return 0, err
		}
	}
}

// Int16 reads int16.
func (d *Decoder) Int16() (int16, error) {
	c, err := d.byte()
	if err != nil {
		return 0, err
	}
	if c == '-' {
		val, err := d.readUInt16()
		if err != nil {
			return 0, err
		}
		if val > math.MaxInt16+1 {
			return 0, errOverflow
		}
		return -int16(val), nil
	}
	d.unread()
	val, err := d.readUInt16()
	if err != nil {
		return 0, err
	}
	if val > math.MaxInt16 {
		return 0, errOverflow
	}
	return int16(val), nil
}

// UInt16 reads uint16.
func (d *Decoder) UInt16() (uint16, error) {
	return d.readUInt16()
}

func (d *Decoder) readUInt16() (uint16, error) {
	c, err := d.byte()
	if err != nil {
		return 0, err
	}
	ind := intDigits[c]
	if ind == 0 {
		return 0, nil
	}
	if ind == invalidCharForNumber {
		return 0, errors.Wrap(err, "bad token")
	}
	value := uint16(ind)
	if d.tail-d.head > 10 {
		i := d.head
		ind2 := intDigits[d.buf[i]]
		if ind2 == invalidCharForNumber {
			d.head = i
			return value, nil
		}
		i++
		ind3 := intDigits[d.buf[i]]
		if ind3 == invalidCharForNumber {
			d.head = i
			return value*10 + uint16(ind2), nil
		}
		i++
		ind4 := intDigits[d.buf[i]]
		if ind4 == invalidCharForNumber {
			d.head = i
			return value*100 + uint16(ind2)*10 + uint16(ind3), nil
		}
		i++
		ind5 := intDigits[d.buf[i]]
		if ind5 == invalidCharForNumber {
			d.head = i
			return value*1000 + uint16(ind2)*100 + uint16(ind3)*10 + uint16(ind4), nil
		}
		i++
		ind6 := intDigits[d.buf[i]]
		value = value*10000 + uint16(ind2)*1000 + uint16(ind3)*100 + uint16(ind4)*10 + uint16(ind5)
		d.head = i
		if ind6 == invalidCharForNumber {
			return value, nil
		}
	}
	for {
		buf := d.buf[d.head:d.tail]
		for i, c := range buf {
			ind = intDigits[c]
			if ind == invalidCharForNumber {
				d.head += i
				return value, nil
			}
			if value > uint16SafeToMultiply10 {
				value2 := (value << 3) + (value << 1) + uint16(ind)
				if value2 < value {
					return 0, errOverflow
				}
				value = value2
				continue
			}
			value = (value << 3) + (value << 1) + uint16(ind)
		}
		err := d.read()
		if err == io.EOF {
			return value, nil
		}
		if err != nil {
			return 0, err
		}
	}
}

// Int32 reads int32.
func (d *Decoder) Int32() (int32, error) {
	c, err := d.byte()
	if err != nil {
		return 0, err
	}
	if c == '-' {
		val, err := d.readUInt32()
		if err != nil {
			return 0, err
		}
		if val > math.MaxInt32+1 {
			return 0, errOverflow
		}
		return -int32(val), nil
	}
	d.unread()
	val, err := d.readUInt32()
	if err != nil {
		return 0, err
	}
	if val > math.MaxInt32 {
		return 0, errOverflow
	}
	return int32(val), nil
}

// UInt32 reads uint32.
func (d *Decoder) UInt32() (uint32, error) {
	return d.readUInt32()
}

func (d *Decoder) readUInt32() (uint32, error) {
	c, err := d.byte()
	if err != nil {
		return 0, err
	}
	ind := intDigits[c]
	if ind == 0 {
		return 0, nil
	}
	if ind == invalidCharForNumber {
		return 0, errors.Wrap(err, "bad token")
	}
	value := uint32(ind)
	if d.tail-d.head > 10 {
		i := d.head
		ind2 := intDigits[d.buf[i]]
		if ind2 == invalidCharForNumber {
			d.head = i
			return value, nil
		}
		i++
		ind3 := intDigits[d.buf[i]]
		if ind3 == invalidCharForNumber {
			d.head = i
			return value*10 + uint32(ind2), nil
		}
		i++
		ind4 := intDigits[d.buf[i]]
		if ind4 == invalidCharForNumber {
			d.head = i
			return value*100 + uint32(ind2)*10 + uint32(ind3), nil
		}
		i++
		ind5 := intDigits[d.buf[i]]
		if ind5 == invalidCharForNumber {
			d.head = i
			return value*1000 + uint32(ind2)*100 + uint32(ind3)*10 + uint32(ind4), nil
		}
		i++
		ind6 := intDigits[d.buf[i]]
		if ind6 == invalidCharForNumber {
			d.head = i
			return value*10000 + uint32(ind2)*1000 + uint32(ind3)*100 + uint32(ind4)*10 + uint32(ind5), nil
		}
		i++
		ind7 := intDigits[d.buf[i]]
		if ind7 == invalidCharForNumber {
			d.head = i
			return value*100000 + uint32(ind2)*10000 + uint32(ind3)*1000 + uint32(ind4)*100 + uint32(ind5)*10 + uint32(ind6), nil
		}
		i++
		ind8 := intDigits[d.buf[i]]
		if ind8 == invalidCharForNumber {
			d.head = i
			return value*1000000 + uint32(ind2)*100000 + uint32(ind3)*10000 + uint32(ind4)*1000 + uint32(ind5)*100 + uint32(ind6)*10 + uint32(ind7), nil
		}
		i++
		ind9 := intDigits[d.buf[i]]
		value = value*10000000 + uint32(ind2)*1000000 + uint32(ind3)*100000 + uint32(ind4)*10000 + uint32(ind5)*1000 + uint32(ind6)*100 + uint32(ind7)*10 + uint32(ind8)
		d.head = i
		if ind9 == invalidCharForNumber {
			return value, nil
		}
	}
	for {
		buf := d.buf[d.head:d.tail]
		for i, c := range buf {
			ind = intDigits[c]
			if ind == invalidCharForNumber {
				d.head += i
				return value, nil
			}
			if value > uint32SafeToMultiply10 {
				value2 := (value << 3) + (value << 1) + uint32(ind)
				if value2 < value {
					return 0, errOverflow
				}
				value = value2
				continue
			}
			value = (value << 3) + (value << 1) + uint32(ind)
		}
		err := d.read()
		if err == io.EOF {
			return value, nil
		}
		if err != nil {
			return 0, err
		}
	}
}

// Int64 reads int64.
func (d *Decoder) Int64() (int64, error) {
	c, err := d.byte()
	if err != nil {
		return 0, err
	}
	if c == '-' {
		c, err := d.next()
		if err != nil {
			return 0, err
		}
		val, err := d.readUInt64(c)
		if err != nil {
			return 0, err
		}
		if val > math.MaxInt64+1 {
			return 0, errors.Errorf("%d overflows", val)
		}
		return -int64(val), nil
	}
	val, err := d.readUInt64(c)
	if err != nil {
		return 0, err
	}
	if val > math.MaxInt64 {
		return 0, errors.Errorf("%d overflows", val)
	}
	return int64(val), nil
}

// UInt64 reads uint64.
func (d *Decoder) UInt64() (uint64, error) {
	c, err := d.byte()
	if err != nil {
		return 0, err
	}
	return d.readUInt64(c)
}

func (d *Decoder) readUInt64(c byte) (uint64, error) {
	ind := intDigits[c]
	if ind == 0 {
		return 0, nil // single zero
	}
	if ind == invalidCharForNumber {
		err := badToken(c, d.offset()-1)
		return 0, errors.Wrap(err, "invalid number")
	}
	value := uint64(ind)
	if d.tail-d.head > 10 {
		i := d.head
		ind2 := intDigits[d.buf[i]]
		if ind2 == invalidCharForNumber {
			d.head = i
			return value, nil
		}
		i++
		ind3 := intDigits[d.buf[i]]
		if ind3 == invalidCharForNumber {
			d.head = i
			return value*10 + uint64(ind2), nil
		}
		i++
		ind4 := intDigits[d.buf[i]]
		if ind4 == invalidCharForNumber {
			d.head = i
			return value*100 + uint64(ind2)*10 + uint64(ind3), nil
		}
		i++
		ind5 := intDigits[d.buf[i]]
		if ind5 == invalidCharForNumber {
			d.head = i
			return value*1000 + uint64(ind2)*100 + uint64(ind3)*10 + uint64(ind4), nil
		}
		i++
		ind6 := intDigits[d.buf[i]]
		if ind6 == invalidCharForNumber {
			d.head = i
			return value*10000 + uint64(ind2)*1000 + uint64(ind3)*100 + uint64(ind4)*10 + uint64(ind5), nil
		}
		i++
		ind7 := intDigits[d.buf[i]]
		if ind7 == invalidCharForNumber {
			d.head = i
			return value*100000 + uint64(ind2)*10000 + uint64(ind3)*1000 + uint64(ind4)*100 + uint64(ind5)*10 + uint64(ind6), nil
		}
		i++
		ind8 := intDigits[d.buf[i]]
		if ind8 == invalidCharForNumber {
			d.head = i
			return value*1000000 + uint64(ind2)*100000 + uint64(ind3)*10000 + uint64(ind4)*1000 + uint64(ind5)*100 + uint64(ind6)*10 + uint64(ind7), nil
		}
		i++
		ind9 := intDigits[d.buf[i]]
		value = value*10000000 + uint64(ind2)*1000000 + uint64(ind3)*100000 + uint64(ind4)*10000 + uint64(ind5)*1000 + uint64(ind6)*100 + uint64(ind7)*10 + uint64(ind8)
		d.head = i
		if ind9 == invalidCharForNumber {
			return value, nil
		}
	}
	for {
		buf := d.buf[d.head:d.tail]
		for i, c := range buf {
			ind = intDigits[c]
			if ind == invalidCharForNumber {
				d.head += i
				return value, nil
			}
			if value > uint64SafeToMultiple10 {
				value2 := (value << 3) + (value << 1) + uint64(ind)
				if value2 < value {
					return 0, errOverflow
				}
				value = value2
				continue
			}
			value = (value << 3) + (value << 1) + uint64(ind)
		}
		switch err := d.read(); err {
		case io.EOF:
			return value, nil
		case nil:
			continue
		default:
			return 0, err
		}
	}
}
