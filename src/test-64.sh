#!/bin/sh

VERSION=37
$NACL_SDK/pepper_${VERSION}/tools/sel_ldr_x86_64 "$@" test-64.nexe
