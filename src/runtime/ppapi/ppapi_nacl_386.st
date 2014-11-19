// -*- mode: asm -*-
// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// TODO(jyh): Figure out why GOOS/GOARCH don't work.
// #include "../zasm_GOOS_GOARCH.h"
#include "../zasm_nacl_386.h"
#include "../../cmd/ld/textflag.h"
#include "../funcdata.h"
#include "../irt_nacl.h"
#include "ppapi_GOOS.h"

// doreturn(ty) expands to the instruction sequence for saving the return
// value into the caller's stack frame.
{% macro doreturn(resultType, off) -%}
	{% set kind = resultType.kind -%}
	{% if kind == 'int32' or kind == '*' -%}
	MOVL	AX, {{off}}(BP)
	{%- elif kind == 'int64' -%}
	MOVL	AX, {{off}}(BP)
	MOVL	DX, ({{off}}+4)(BP)
	{%- elif kind == 'float32' -%}
	FMOVLP	F0, {{off}}(BP)
	{%- elif kind == 'float64' -%}
	FMOVDP	F0, {{off}}(BP)
	{%- else -%}
	// No return value.
	{%- endif %}
{%- endmacro %}

// Callbacks are invoked through cgocallback.
{% for func in callbacks -%}
{% set fsize = framesize(func) -%}
TEXT ppapi·{{func.name}}(SB),NOSPLIT,${{fsize+32}}
	MOVL	BP, {{fsize+16}}(SP)
	MOVL	BX, {{fsize+20}}(SP)
	MOVL	SI, {{fsize+24}}(SP)
	MOVL	DI, {{fsize+28}}(SP)
	{% for aoff in range(0, fsize, 4) -%}
	MOVL	arg{{aoff}}+{{aoff}}(FP), AX
	MOVL	AX, {{aoff+12}}(SP)
	{% endfor -%}
	MOVL	$·{{func.name}}(SB), AX
	MOVL	AX, 0(SP)
	LEAL	12(SP), AX
	MOVL	AX, 4(SP)
	MOVL	${{fsize+4}}, 8(SP)
	MOVL	$0, {{fsize+12}}(SP)
	CALL	runtime·cgocallback(SB)
	MOVL	{{fsize+12}}(SP), AX
	MOVL	{{fsize+16}}(SP), BP
	MOVL	{{fsize+20}}(SP), BX
	MOVL	{{fsize+24}}(SP), SI
	MOVL	{{fsize+28}}(SP), DI
	RET

{% endfor %}

// PPAPI calls are invoked using cgocall.
{% for func in functions -%}
{% set fsize = framesize(func) -%}
TEXT ·{{func.name}}(SB),NOSPLIT,$8
	MOVL	$ppapi·{{func.name}}(SB), 0(SP)
	LEAL	arg0+0(FP), AX
	MOVL	AX, 4(SP)
	CALL	runtime·cgocall(SB)
	RET

TEXT ppapi·{{func.name}}(SB),NOSPLIT,${{fsize}}
	MOVL	args+0(FP), BP
	{% for aoff in range(0, fsize, 4) -%}
	MOVL	{{aoff}}(BP), AX
	MOVL	AX, {{aoff}}(SP)
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
TEXT ppapi·ppp_initialize_module_handler(SB),NOSPLIT,$12
	MOVL	SI, 4(SP)
	MOVL	DI, 8(SP)
	MOVL	module_id+0(FP), AX
	MOVL	AX, ppapi·module_id(SB)
	MOVL	get_browser_interface+4(FP), DI
	LEAL	ppapi·ppb_interfaces(SB), SI
initialize_module_loop:
	MOVL	0(SI), AX  // name
	TESTL	AX, AX
	JZ	initialize_module_done
	MOVL	AX, 0(SP)
	CALL	DI
	MOVL	AX, 4(SI)  // ppb
	ADDL	$8, SI
	JMP	initialize_module_loop
initialize_module_done:
	MOVL	4(SP), SI
	MOVL	8(SP), DI
	XORL	AX, AX
	RET

// ppapi·ppp_shutdown_module_handler my or may not be called when the
// module is closed.  Ignore the callback.  Called from the C stack.
TEXT ppapi·ppp_shutdown_module_handler(SB),NOSPLIT,$0
	RET

// ppapi·ppp_get_interface_handler is called by the browser to get
// callback functions.  Called from the C stack.
TEXT ppapi·ppp_get_interface_handler(SB),NOSPLIT,$24
	MOVL	BP, 8(SP)
	MOVL	BX, 12(SP)
	MOVL	SI, 16(SP)
	MOVL	DI, 20(SP)
	MOVL	interface_name+0(FP), AX
	MOVL	AX, 0(SP)
	CALL	ppapi·ppp_get_interface(SB)
	MOVL	4(SP), AX
	MOVL	8(SP), BP
	MOVL	12(SP), BX
	MOVL	16(SP), SI
	MOVL	20(SP), DI
	RET

// ppapi·start is called to start PPAPI.  Never returns.
TEXT ppapi·start(SB),NOSPLIT,$4
	LEAL	ppapi·pp_start_functions(SB), DI
	MOVL	DI, 0(SP)
	MOVL	runtime·nacl_irt_ppapihook_v0_1+IRT_PPAPI_START(SB), AX
	CALL	AX
	RET

TEXT ·ppapi_start(SB),NOSPLIT,$8-0
	NO_LOCAL_POINTERS
	LEAL	ppapi·start(SB), AX
	MOVL	AX, 0(SP)
	MOVL	$0, 4(SP)
	CALL	runtime·cgocall(SB)
	// Not reached
	INT	$3
	RET

// Tunnel some functions from runtime.
TEXT ·gostring(SB),NOSPLIT,$0
	JMP	runtime·gostring(SB)

TEXT ·gostringn(SB),NOSPLIT,$0
	JMP	runtime·gostringn(SB)

TEXT ·free(SB),NOSPLIT,$0
	JMP	runtime·cfree(SB)
