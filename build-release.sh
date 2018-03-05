#!/bin/sh

go test . && go build -ldflags "-s" .
