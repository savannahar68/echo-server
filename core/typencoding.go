package core

import "errors"

func GetType(te uint8) uint8 {
	return te & 0b11110000
}

func GetEncoding(te uint8) uint8 {
	return te & 0b00001111
}

func AssertType(te uint8, t uint8) error {
	if GetType(te) != t {
		return errors.New("the operation is not permitted on this type")
	}
	return nil
}

func AssertEncoding(te uint8, e uint8) error {
	if GetEncoding(te) != e {
		return errors.New("the operation is not permitted on this encoding")
	}
	return nil
}
