#!/usr/bin/env bash

# Check if golangci-lint is installed
if ! command -v golangci-lint &> /dev/null && ! command -v $(go env GOPATH)/bin/golangci-lint &> /dev/null; then
    echo "golangci-lint not found. Installing..."
    # Install latest version of golangci-lint
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
fi

# Run golangci-lint
echo "Running golangci-lint..."
$(go env GOPATH)/bin/golangci-lint run --timeout=5m 