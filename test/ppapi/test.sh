#!/bin/sh

VERSION=37
$NACL_SDK/pepper_${VERSION}/tools/sel_ldr_x86_64 -B $NACL_SDK/pepper_${VERSION}/tools/irt_core_x86_64.nexe "$@" basic_x86_64.nexe
