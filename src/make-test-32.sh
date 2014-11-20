#!/bin/sh

GOARCH=386 GOOS=nacl ../bin/go build -o test-32.nexe ../test/"$1".go
