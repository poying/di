#!/usr/bin/env bash

set -e
echo "mode: atomic" > coverage.txt

for d in $(go list ./... | grep -v vendor); do
    go test -race -timeout=5s -coverprofile=profile.out -coverpkg=./... -covermode=atomic $d
    if [ -f profile.out ]; then
        cat profile.out | grep -v '^mode' >> coverage.txt
        rm profile.out
    fi
done