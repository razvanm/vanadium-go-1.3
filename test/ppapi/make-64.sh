#!/bin/sh -x -e

export GOROOT=${HOME}/google/nacl/go-ppapi-file
GOOS=nacl GOARCH=amd64p32 ${GOROOT}/bin/go build -o basic_x86_64.nexe basic_nacl.go
GOOS=nacl GOARCH=amd64p32 ${GOROOT}/bin/go build -o mandel_x86_64.nexe mandel_nacl.go
GOOS=nacl GOARCH=amd64p32 ${GOROOT}/bin/go build -o file_io_x86_64.nexe file_io_nacl.go
