#!/bin/sh

GOARCH=amd64p32 GOOS=nacl ../bin/go build -o test-64.nexe ../test/helloworld.go
