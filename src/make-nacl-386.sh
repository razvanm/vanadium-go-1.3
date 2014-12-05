#!/bin/bash
  
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "${DIR}"
#(cd runtime/ppapi; ./mkzfile.py config.txt ppapi_nacl_386.st > zppapi_nacl_386.s)
#(cd runtime/ppapi; ./mkzfile.py --include=$NACL_SDK/pepper_35/include/ppapi/c config.txt cdecl_nacl.got > zcdecl_nacl.go)
GOOS=nacl GOARCH=386 "${DIR}/make.bash"
