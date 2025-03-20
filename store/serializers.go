package store

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"fmt"
)

func Serialize(value any) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	// Compress the JSON data
	var compressed bytes.Buffer
	writer := zlib.NewWriter(&compressed)
	_, err = writer.Write(data)
	if err != nil {
		fmt.Println("Error compressing data:", err)
		return nil, err
	}
	writer.Close()

	return compressed.Bytes(), nil
}

func Deserialize(data []byte, dest any) error {
	// Decompress the JSON data
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer reader.Close()
	var decompressed bytes.Buffer
	_, err = decompressed.ReadFrom(reader)
	if err != nil {
		return err
	}
	return json.Unmarshal(decompressed.Bytes(), dest)
}
