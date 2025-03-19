#/!/bin/bash
protoc --go_out=pb --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative value.proto