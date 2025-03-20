package store

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"fmt"
)

func Serialize(value any) ([]byte, error) {
	var compressed bytes.Buffer
	writer := zlib.NewWriter(&compressed)

	encoder := json.NewEncoder(writer)

	err := encoder.Encode(value)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		fmt.Println("Error closing zlib writer:", err)
		return nil, err
	}

	return compressed.Bytes(), nil
}

func Deserialize(data []byte, dest any) error {
	// Decompress the JSON data
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer reader.Close()

	return json.NewDecoder(reader).Decode(dest)
}
