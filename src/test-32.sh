#!/bin/sh

VERSION=37
$NACL_SDK/pepper_${VERSION}/tools/sel_ldr_x86_32 "$@" test-32.nexe
