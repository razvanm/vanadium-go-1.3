#!/bin/sh

(cd runtime/ppapi && ./mkzfile.py config.txt ppapi_nacl_amd64p32.st > zppapi_nacl_amd64p32.s)
GOOS=nacl GOARCH=amd64p32 ./make.bash
