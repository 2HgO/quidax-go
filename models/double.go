package models

import (
	"encoding/json"
	"strings"
)

type Double float64

func (d *Double) UnmarshalJSON(input []byte) error {
	if d == nil {
		d = new(Double)
	}
	strInput := string(input)
	strInput = strings.Trim(strInput, `"`)
	var buf float64
	err := json.Unmarshal([]byte(strInput), &buf)
	if err == nil {
		*d = Double(buf)
	}
	return err
}

func (d Double) MarshalJSON() ([]byte, error) {
	return json.Marshal(float64(d))
}
