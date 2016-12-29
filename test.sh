#!/usr/bin/env bash

# https://github.com/codecov/example-go/blob/2cc4936b5d2b4e64eb1c6a314a2531808e157fb7/README.md#caveat-multiple-files

set -e
echo "" > coverage.txt

for d in $(go list ./... | grep -v vendor); do
    go test -race -coverprofile=profile.out -covermode=atomic $d
    if [ -f profile.out ]; then
        cat profile.out >> coverage.txt
        rm profile.out
    fi
done
