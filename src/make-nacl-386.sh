#!/bin/bash
  
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "${DIR}"
GOOS=nacl GOARCH=386 "${DIR}/make-tmp.sh"
