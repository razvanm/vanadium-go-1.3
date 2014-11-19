#!/bin/bash
  
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "${DIR}"
GOOS=nacl GOARCH=amd64p32 "${DIR}/make.bash"
