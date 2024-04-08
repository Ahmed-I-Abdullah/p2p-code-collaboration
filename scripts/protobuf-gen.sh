#!/bin/bash

# This script compiles Protocol Buffers (.proto) files found in the current Git repository into Go code using protoc
cd "$(git rev-parse --show-toplevel)"
protoc -I ./ --go_out=./pb --go-grpc_out=./pb ./pb/*.proto