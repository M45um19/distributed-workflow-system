#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"

if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" ]]; then
    PROTO_BASE_DIR=$(cygpath -m "$PROJECT_ROOT/shared-proto")
    GO_PB_DIR=$(cygpath -m "$PROJECT_ROOT/services/workspace-service/pb")
else
    PROTO_BASE_DIR="$PROJECT_ROOT/shared-proto"
    GO_PB_DIR="$PROJECT_ROOT/services/workspace-service/pb"
fi

echo "Target Directory: $GO_PB_DIR"
mkdir -p "$GO_PB_DIR"

if [ ! -d "$PROTO_BASE_DIR" ]; then
    echo "Error: Shared proto directory not found at $PROTO_BASE_DIR"
    exit 1
fi

echo "Generating Go PB files..."

export MSYS_NO_PATHCONV=1

protoc --proto_path="$PROTO_BASE_DIR" \
       --go_out="$GO_PB_DIR" --go_opt=paths=source_relative \
       --go-grpc_out="$GO_PB_DIR" --go-grpc_opt=paths=source_relative \
       $(find "$PROTO_BASE_DIR" -name "*.proto")

if [ $? -eq 0 ]; then
    echo "Success! Go PB files generated."
else
    echo "Error: Proto generation failed!"
    exit 1
fi