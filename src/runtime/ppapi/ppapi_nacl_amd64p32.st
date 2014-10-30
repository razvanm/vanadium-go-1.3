// -*- mode: asm -*-
// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "../zasm_nacl_amd64p32.h"
#include "../../cmd/ld/textflag.h"
#include "../irt_nacl.h"
#include "ppapi_GOOS.h"

// doreturn(ty) expands to the instruction sequence for saving the return
// value into the caller's stack frame.
{% macro doreturn(resultType, off) -%}
	{% set kind = resultType.kind -%}
	{% if kind == 'int32' or kind == '*' -%}
	MOVL	AX, rval+{{off}}(FP)
	{%- elif kind == 'int64' -%}
	MOVQ	AX, rval+{{off}}(FP)
	{%- elif kind == 'float32' -%}
	FMOVLP	F0, rval+{{off}}(FP)
	{%- elif kind == 'float64' -%}
	FMOVDP	F0, rval+{{off}}(FP)
	{%- else -%}
	// No return value.
	{%- endif %}
{%- endmacro %}

// Callbacks are invoked through cgocallback.
{% for func in callbacks -%}
{% set fsize = framesize(func) -%}
TEXT ppapi·{{func.name}}(SB),NOSPLIT,${{align(fsize+48,16)}}
	MOVQ	BX, {{fsize+16}}(SP)
	MOVQ	R12, {{fsize+24}}(SP)
	MOVQ	R13, {{fsize+32}}(SP)
	MOVQ	R14, {{fsize+40}}(SP)
	{% for aoff in range(0, fsize, 4) -%}
	MOVL	arg{{aoff}}+{{aoff}}(FP), AX
	MOVL	AX, {{aoff+12}}(SP)
	{% endfor -%}
	LEAL	·{{func.name}}(SB), AX
	MOVL	AX, 0(SP)
	LEAL	12(SP), AX
	MOVL	AX, 4(SP)
	MOVL	${{fsize+4}}, 8(SP)
	MOVL	$0, {{fsize+12}}(SP)
	CALL	runtime·cgocallback(SB)
	MOVL	{{fsize+12}}(SP), AX
	MOVQ	{{fsize+16}}(SP), BX
	MOVQ	{{fsize+24}}(SP), R12
	MOVQ	{{fsize+32}}(SP), R13
	MOVQ	{{fsize+40}}(SP), R14
	RET

{% endfor %}

// PPAPI calls are invoked using cgocall.
{% for func in functions -%}
{% set fsize = framesize(func) -%}
TEXT ·{{func.name}}(SB),NOSPLIT,$8
	LEAL	ppapi·{{func.name}}(SB), AX
	MOVL	AX, 0(SP)
	LEAL	arg0+0(FP), AX
	MOVL	AX, 4(SP)
	CALL	runtime·cgocall(SB)
	RET

TEXT ppapi·{{func.name}}(SB),NOSPLIT,${{align(max(0, fsize-24), 16)}}
	{% if fsize > 0 -%}
	MOVL	arg0+0(FP), DI
	{% endif -%}
	{% if fsize > 4 -%}
	MOVL	arg1+4(FP), SI
	{% endif -%}
	{% if fsize > 8 -%}
	MOVL	arg2+8(FP), DX
	{% endif -%}
	{% if fsize > 12 -%}
	MOVL	arg3+12(FP), CX
	{% endif -%}
	{% if fsize > 16 -%}
	MOVL	arg4+16(FP), R8
	{% endif -%}
	{% if fsize > 20 -%}
	MOVL	arg5+20(FP), R9
	{% endif -%}
	{% for aoff in range(24, fsize, 4) -%}
	MOVL	arg{{aoff}}+{{aoff}}(FP), AX
	MOVL	AX, {{aoff-24}}(SP)
	{% endfor -%}
	MOVL	ppapi·ppb_interfaces+({{func.interface}}*8+4)(SB), AX
	MOVL	({{func.index}}*4)(AX), AX
	CALL	AX
	{% if func.structReturn -%}
	SUBL	$4, SP	// Adjust SP due to struct return.
	{% endif -%}
	{{ doreturn(func.result, fsize) }}
	RET

{% endfor %}

// ppapi·ppp_initialize_module_handler is called once at initialization
// initialization time.  Called from the C stack.
TEXT ppapi·ppp_initialize_module_handler(SB),NOSPLIT,$36
	MOVQ	BX, 4(SP)
	MOVQ	R12, 12(SP)
	MOVQ	R13, 20(SP)
	MOVQ	R14, 28(SP)
	MOVL	module_id+0(FP), AX
	MOVL	AX, ppapi·module_id(SB)
	MOVL	get_browser_interface+4(FP), AX
	MOVL	AX, 0(SP)
	CALL	ppapi·ppp_initialize_module(SB)
	MOVQ	4(SP), BX
	MOVQ	12(SP), R12
	MOVQ	20(SP), R13
	MOVQ	28(SP), R14
	RET

// ppapi·ppp_shutdown_module_handler my or may not be called when the
// module is closed.  Ignore the callback.  Called from the C stack.
TEXT ppapi·ppp_shutdown_module_handler(SB),NOSPLIT,$0
	RET

// ppapi·ppp_get_interface_handler is called by the browser to get
// callback functions.  Called from the C stack.
TEXT ppapi·ppp_get_interface_handler(SB),NOSPLIT,$36
	MOVQ	BX, 4(SP)
	MOVQ	R12, 12(SP)
	MOVQ	R13, 20(SP)
	MOVQ	R14, 28(SP)
	MOVL	interface_name+0(FP), AX
	MOVL	AX, 0(SP)
	CALL	ppapi·ppp_get_interface(SB)
	MOVQ	4(SP), BX
	MOVQ	12(SP), R12
	MOVQ	20(SP), R13
	MOVQ	28(SP), R14
	RET

// ppapi·start is called to start PPAPI.  Never returns.
TEXT ppapi·start(SB),NOSPLIT,$0
	LEAL	ppapi·pp_start_functions(SB), DI
	MOVL	runtime·nacl_irt_ppapihook_v0_1+IRT_PPAPI_START(SB), AX
	CALL	AX
	RET

TEXT ·ppapi_start(SB),NOSPLIT,$8
	LEAL	ppapi·start(SB), AX
	MOVL	AX, 0(SP)
	MOVL	$0, 4(SP)
	CALL	runtime·cgocall(SB)
	// Not reached
	INT	$3
	RET
