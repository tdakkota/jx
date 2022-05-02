package jx

type charClass byte

func (class charClass) is(set charClass) bool {
	return class&set != 0
}

const (
	// Space characters.
	charSpace charClass = 1 << iota
	// Characters which should be escaped by string encoder.
	charEscape
	// Digits.
	charDigit
	// Possible characters after number.
	charNumberEnd
	// JSON number characters (digits, signs, dot, e/E).
	charNumber
)

func mergeSets(sets ...[256]charClass) (r [256]charClass) {
	for _, set := range sets {
		for i, val := range set {
			r[i] |= val
		}
	}
	return r
}

var charset = mergeSets(
	[256]charClass{' ': charSpace, '\n': charSpace, '\t': charSpace, '\r': charSpace},
	// safeSet holds the value true if the ASCII character with the given array
	// position can be represented inside a JSON string without any further
	// escaping.
	//
	// All values are true except for the ASCII control characters (0-31), the
	// double quote ("), and the backslash character ("\").
	[256]charClass{
		// First 31 characters.
		charEscape, charEscape, charEscape, charEscape, charEscape, charEscape, charEscape, charEscape,
		charEscape, charEscape, charEscape, charEscape, charEscape, charEscape, charEscape, charEscape,
		charEscape, charEscape, charEscape, charEscape, charEscape, charEscape, charEscape, charEscape,
		charEscape, charEscape, charEscape, charEscape, charEscape, charEscape, charEscape, charEscape,
		'"':  charEscape,
		'\\': charEscape,
	},
	[256]charClass{
		'+': charNumber,
		'-': charNumber,
		'.': charNumber,
		'e': charNumber,
		'E': charNumber,
		'0': charDigit | charNumber,
		'1': charDigit | charNumber,
		'2': charDigit | charNumber,
		'3': charDigit | charNumber,
		'4': charDigit | charNumber,
		'5': charDigit | charNumber,
		'6': charDigit | charNumber,
		'7': charDigit | charNumber,
		'8': charDigit | charNumber,
		'9': charDigit | charNumber,
	},
	[256]charClass{
		',':  charNumberEnd,
		']':  charNumberEnd,
		'}':  charNumberEnd,
		' ':  charNumberEnd,
		'\t': charNumberEnd,
		'\n': charNumberEnd,
		'\r': charNumberEnd,
	},
)
