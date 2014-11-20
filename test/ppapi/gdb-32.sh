#!/bin/sh
VERSION=37
$NACL_SDK/pepper_${VERSION}/toolchain/mac_x86_newlib/bin/i686-nacl-gdb "$@"
