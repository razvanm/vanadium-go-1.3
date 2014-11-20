#!/bin/sh -x -e

export GOROOT=${HOME}/google/nacl/go-ppapi-veyron
GOOS=nacl GOARCH=386 ${GOROOT}/bin/go build -o basic_x86_32.nexe basic_nacl.go
GOOS=nacl GOARCH=386 ${GOROOT}/bin/go build -o mandel_x86_32.nexe mandel_nacl.go
GOOS=nacl GOARCH=386 ${GOROOT}/bin/go build -o file_io_x86_32.nexe file_io_nacl.go
GOOS=nacl GOARCH=386 ${GOROOT}/bin/go build -o network_x86_32.nexe network_nacl.go
GOOS=nacl GOARCH=386 ${GOROOT}/bin/go build -o message_x86_32.nexe message_nacl.go
