#!/bin/bash

protoc \
  --plugin=protoc-gen-ts=$(which protoc-gen-ts) \
  -I . \
  --js_out=import_style=commonjs,binary:. \
  --ts_out=service=grpc-web:. \
  api.proto

go generate
