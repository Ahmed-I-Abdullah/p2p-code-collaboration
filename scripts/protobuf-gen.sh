#!/bin/bash
cd "$(git rev-parse --show-toplevel)"
protoc -I ./ --go_out=./pb --go-grpc_out=./pb ./pb/*.proto