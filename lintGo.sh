#!/bin/bash
if go run github.com/golangci/golangci-lint/cmd/golangci-lint \
	run ./... \
	--timeout=5m \
	--out-format colored-line-number \
	--skip-dirs-use-default; then
	echo "OK"
else
	echo "FAILED"
fi

