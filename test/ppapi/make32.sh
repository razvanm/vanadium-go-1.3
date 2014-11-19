#!/bin/sh -x -e

export GOROOT=${HOME}/google/nacl/go-ppapi
GOOS=nacl GOARCH=386 ${GOROOT}/bin/go build -o basic_x86_32.nexe basic_nacl.go
GOOS=nacl GOARCH=386 ${GOROOT}/bin/go build -o mandel_x86_32.nexe mandel_nacl.go
