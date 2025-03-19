package store

import (
	"encoding/json"
	"errors"

	"github.com/johannessarpola/poor-cache-go/pb"
)

func Serialize(value any) (*pb.Value, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	return &pb.Value{
		Data: data,
	}, nil
}

func Deserialize(pbValue *pb.Value, dest any) error {
	if pbValue == nil || len(pbValue.Data) == 0 {
		return errors.New("no data to deserialize")
	}
	return json.Unmarshal(pbValue.Data, dest)
}
