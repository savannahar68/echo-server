package core

import (
	"errors"
	"fmt"
)

// DecodeArrayString simple function to convert byte to array of strings (specifically)
// instead of just interface (this is more of a helper method)
func DecodeArrayString(data []byte) ([]string, error) {
	value, err := Decode(data)
	if err != nil {
		return nil, err
	}

	if value == nil {
		return nil, errors.New("no data")
	}

	ts := value.([]interface{})
	tokens := make([]string, len(ts))

	for i := range tokens {
		tokens[i] = ts[i].(string)
	}
	return tokens, nil
}

func Decode(data []byte) (interface{}, error) {
	if len(data) == 0 {
		return nil, errors.New("no data")
	}

	// the 2nd value is delta which indicates the length uptil which encoding has happened
	value, _, err := DecodeOne(data)
	return value, err
}

// DecodeOne decode only the 1st value
func DecodeOne(data []byte) (interface{}, int, error) {
	if len(data) == 0 {
		return nil, 0, errors.New("no data")
	}
	switch data[0] {
	case '+':
		return readSimpleString(data)

	case ':':
		return readInt64(data)

	case '$':
		return readBulkString(data)

	case '*':
		return readArray(data)

	case '-':
		return readError(data)
	}

	return nil, 0, nil
}

// readSimpleString reads the RESP encoded simple string and returns
// string, the delta and the error
// Example of encoded simple string in RESP "+OK\r\n"
func readSimpleString(data []byte) (string, int, error) {
	// pos first value is +
	pos := 1

	for ; data[pos] != '\r'; pos++ {
	}

	return string(data[1:pos]), pos + 2, nil
}

// readError reads the RESP encoded error from data and returns
// the error string, the delta and parsing error if any
// Example of encoded error in RESP "-This is a error\r\n"
func readError(data []byte) (string, int, error) {
	return readSimpleString(data)
}

// readInt64 reads the RESP encoded error from data and returns
// the int64, the delta and parsing error if any
// Example of encoded error in RESP ":10\r\n"
func readInt64(data []byte) (int64, int, error) {
	pos := 1
	var value int64 = 0
	for ; data[pos] != '\r'; pos++ {
		value = value*10 + int64(data[pos]-'0')
	}

	return value, pos + 2, nil
}

// readBulkString reads the RESP encoded error from data and returns
// the string, the delta and parsing error if any
// Example of encoded error in RESP "$4\r\nOkay\r\n"
func readBulkString(data []byte) (string, int, error) {
	// first character $
	pos := 1

	// reading the length and forwarding the pos by
	// the length of the integer + the first special character
	len, delta := readLength(data[pos:])
	pos += delta

	// reading `len` bytes as string
	return string(data[pos:(pos + len)]), pos + len + 2, nil
}

// readArray reads the RESP encoded error from data and returns
// the array of elements, the delta and parsing error if any
// Example of encoded error in RESP "*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n"
func readArray(data []byte) ([]interface{}, int, error) {
	pos := 1
	count, delta := readLength(data[pos:])
	pos += delta

	elements := make([]interface{}, count)
	for i := range elements {
		elem, delta, err := DecodeOne(data[pos:])
		if err != nil {
			return nil, 0, err
		}
		elements[i] = elem
		pos += delta
	}

	return elements, pos, nil
}

// readLength reads the lenght of the string until it hits first non digit string
// returns length of the string and delta (next start point)
func readLength(data []byte) (int, int) {
	pos, length := 0, 0
	for pos = range data {
		b := data[pos]
		if !(b >= '0' && b <= '9') {
			return length, pos + 2
		}
		length = length*10 + int(b-'0')
	}
	return 0, 0
}

func Encode(value interface{}, isSimple bool) []byte {
	switch v := value.(type) {
	case string:
		if isSimple {
			return []byte(fmt.Sprintf("+%s\r\n", v))
		}
		return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v))
	case int64:
		return []byte(fmt.Sprintf(":%d\r\n", v))
	}
	return []byte{}
}
