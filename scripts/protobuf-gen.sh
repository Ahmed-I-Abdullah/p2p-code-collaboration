#!/bin/bash
cd "$(git rev-parse --show-toplevel)"
protoc -I ./internal/api --go_out=./internal/api/pb --go-grpc_out=./internal/api/pb ./internal/api/pb/*.proto