package cmotel

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

// MarshalSpanMap marshals into a string the map of spans so that they can be passed as a header
func MarshalSpanMap(spans map[string]string) (string, error) {
	// Encoding the map
	marshaled, err := json.Marshal(spans)
	if err != nil {
		return "", err
	}

	return string(marshaled), nil
}

// UnmarshalToSpanMap unmarshal the specified string to a map
func UnmarshalToSpanMap(span string) (map[string]string, error) {
	b := new(bytes.Buffer)
	countWrite, errWrite := b.WriteString(span)
	if countWrite != len(span) {
		return map[string]string{}, fmt.Errorf("number of bytes written was %d while length of string provided was %d", countWrite, len(span))
	}

	if errWrite != nil {
		return map[string]string{}, errors.Join(errors.New("could not write the bytes to be converted"), errWrite)
	}

	// Decoding the serialized data
	var decodedMap map[string]string
	err := json.Unmarshal(b.Bytes(), &decodedMap)
	if err != nil {
		return map[string]string{}, errors.Join(errors.New("could not decode to a map"), err)
	}

	return decodedMap, nil
}
