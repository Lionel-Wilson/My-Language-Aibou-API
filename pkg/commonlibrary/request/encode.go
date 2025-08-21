package request

import (
	"bytes"
	"encoding/json"
)

func JsonReader(v any) (*bytes.Reader, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(b), nil
}
