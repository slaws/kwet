package lib

import (
	"bytes"
	"encoding/gob"
)

func ContainsString(list []string, str string) bool {
	for _, b := range list {
		if b == str {
			return true
		}
	}
	return false
}

func GetBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
