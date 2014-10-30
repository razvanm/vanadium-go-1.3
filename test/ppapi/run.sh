#!/bin/sh
$NACL_ROOT/native_client/scons-out/opt-mac-x86-32/staging/sel_ldr -B $NACL_ROOT/native_client/scons-out/nacl_irt-x86-32/staging/irt_core.nexe -f basic_x86_32.nexe "$@"
