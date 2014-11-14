// -*- mode: asm -*-
// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "../zasm_nacl_amd64p32.h"
#include "../../cmd/ld/textflag.h"
#include "../irt_nacl.h"
#include "../funcdata.h"
#include "ppapi_GOOS.h"

{% if False %}
// amd64return(ty) expands to the instruction sequence for saving the return
// value into the caller's stack frame.
{% endif %}
{% macro amd64return(resultType, off, fp) -%}
	{%- set kind = resultType.kind -%}
	{%- set size = resultType.size -%}
	{%- if size == 0 -%}
	// No return value
	{%- elif kind == 'float32' -%}
	MOVSS	X0, {{off}}({{fp}})
	{%- elif kind == 'float64' -%}
	MOVSD	X0, {{off}}({{fp}})
	{%- elif size == 4 -%}
	MOVL	AX, {{off}}({{fp}})
	{%- elif size == 8 -%}
	MOVQ	AX, {{off}}({{fp}})
	{%- elif size == 12 -%}
	MOVQ	AX, {{off}}({{fp}})  // (sizeof {{size}})
	MOVL	DX, {{off+8}}({{fp}})
	{%- elif size == 16 -%}
	MOVQ	AX, {{off}}({{fp}})  // (sizeof {{size}})
	MOVQ	DX, {{off+8}}({{fp}})
	{%- else -%}
	Bogus return value of size {{size}}
	{%- endif %}
{%- endmacro -%}

{% if False %}
// amd64args sets up the arguments, converting from the Go calling convention,
// where everything is passed on the stack, to the gcc calling convention, where:
//   - integer args are passed in registers DI, SI, DX, CX, R8, R9.
//   - float args are passed in registers X0, ..., X5.
//   - structs smaller than 16 bytes are passed in registers.
//   - everything else is passed on the stack.
//
// AX refers to the caller stack frame.
// SP refers the the callee stack frame.
// Destroys DX.
{% endif %}
{% macro amd64args(args) -%}
  {%- set regs = ['DI', 'SI', 'DX', 'CX', 'R8', 'R9'] -%}
  {%- set fregs = ['X0', 'X1', 'X2', 'X3', 'X4', 'X5'] -%}
  {%- set ain = 0 -%}
  {%- set soff = 0 -%}
  {%- for arg in args -%}
    {%- set ty = arg.type -%}
    {%- set kind = ty.kind -%}
    {%- set size = ty.size -%}
    {%- if ty.align > 0 %}
      {%- set ain = align(ain, ty.align) -%}
    {%- endif -%}
    {%- if kind == 'float32' and fregs.__len__() >= 1 %}
	MOVSS	{{ain}}(AX), {{fregs[0]}}  // {{arg.name}}
	{%- set fregs = fregs[1:] -%}
    {%- elif kind == 'float64' and fregs.__len__() >= 1 %}
	MOVSD	{{ain}}(AX), {{fregs[0]}}  // {{arg.name}}
	{%- set fregs = fregs[1:] -%}
    {%- elif size == 4 and regs.__len__() >= 1 %}
	MOVL	{{ain}}(AX), {{regs[0]}}  // {{arg.name}}
	{%- set regs = regs[1:] -%}
    {%- elif size == 8 and regs.__len__() >= 1 %}
	MOVQ	{{ain}}(AX), {{regs[0]}}  // {{arg.name}}
	{%- set regs = regs[1:] -%}
    {%- elif size == 12 and regs.__len__() >= 2 %}
	MOVQ	{{ain}}(AX), {{regs[0]}}  // {{arg.name}} (sizeof {{size}})
	MOVL	{{ain+8}}(AX), {{regs[1]}}
	{%- set regs = regs[2:] -%}
    {%- elif size == 16 and regs.__len__() >= 2 %}
	MOVQ	{{ain}}(AX), {{regs[0]}}  // {{arg.name}} (sizeof {{size}})
	MOVQ	{{ain+8}}(AX), {{regs[1]}}
	{%- set regs = regs[2:] -%}
    {%- else %}
	// {{arg.name}} (sizeof {{size}})
      {%- for aoff in range(0, size, 4) %}
	MOVL	{{ain+aoff}}(AX), DX
	MOVL	DX, {{soff+aoff}}(SP)
      {%- endfor -%}
      {%- set soff = soff + size -%}
    {%- endif -%}
    {%- set ain = ain + size -%}
  {%- endfor -%}
{%- endmacro -%}

// Callbacks are invoked through cgocallback.
{% for func in callbacks -%}
{% set fsize = align(framesize(func),8) -%}
TEXT ppapi·{{func.name}}(SB),NOSPLIT,${{align(fsize+56,16)}}
	LEAL	0(BP), AX
	MOVL	AX, {{fsize+20}}(SP)
	MOVQ	BX, {{fsize+24}}(SP)
	MOVQ	R12, {{fsize+32}}(SP)
	MOVQ	R13, {{fsize+40}}(SP)
	MOVQ	R14, {{fsize+48}}(SP)
	{% if fsize > 0 -%}
	MOVL	DI, 16(SP)
	{% endif -%}
	{% if fsize > 4 -%}
	MOVL	SI, 20(SP)
	{% endif -%}
	{% if fsize > 8 -%}
	MOVL	DX, 24(SP)
	{% endif -%}
	{% if fsize > 12 -%}
	MOVL	CX, 28(SP)
	{% endif -%}
	{% if fsize > 16 -%}
	MOVL	R8, 32(SP)
	{% endif -%}
	{% if fsize > 20 -%}
	MOVL	R9, 36(SP)
	{% endif -%}
	{% for aoff in range(24, fsize, 4) -%}
	MOVL	arg{{aoff/4}}+{{aoff-24}}(FP), AX
	MOVL	AX, {{aoff+12}}(SP)
	{% endfor -%}
	MOVL	$0, {{fsize+16}}(SP)
	LEAL	·{{func.name}}(SB), AX
	MOVL	AX, 0(SP)
	LEAL	16(SP), AX
	MOVL	AX, 4(SP)
	MOVL	${{fsize+4}}, 8(SP)
	CALL	runtime·cgocallback(SB)
	MOVL	{{fsize+16}}(SP), AX
	MOVL	{{fsize+20}}(SP), BX
	LEAL	0(BX), BP
	MOVQ	{{fsize+24}}(SP), BX
	MOVQ	{{fsize+32}}(SP), R12
	MOVQ	{{fsize+40}}(SP), R13
	MOVQ	{{fsize+48}}(SP), R14
	RET

{% endfor %}

// PPAPI calls are invoked using cgocall.
{% for func in functions -%}
{% set fsize = framesize(func) -%}
{% set asize = align(fsize, 16)+8 -%}
TEXT ·{{func.name}}(SB),NOSPLIT,$8
	GO_ARGS
	LEAL	ppapi·{{func.name}}(SB), AX
	MOVL	AX, 0(SP)
	LEAL	arg0+0(FP), AX
	MOVL	AX, 4(SP)
	CALL	runtime·cgocall(SB)
	RET

// Called on the C stack.
{% if func.structReturn and func.result.size <= 16 -%}
TEXT ppapi·{{func.name}}(SB),NOSPLIT,${{asize}}
	MOVL	DI, AX
	MOVL	0(AX), DX  // *rval
	MOVL	DX, {{asize-4}}(SP)
	ADDL	$4, AX
	{{- amd64args(func.args[1:]) }}
	MOVL	ppapi·ppb_interfaces+({{func.interface}}*8+4)(SB), AX
	MOVL	({{func.index}}*4)(AX), AX
	CALL	AX
	MOVL	{{asize-4}}(SP), DI
	{{ amd64return(func.result, 0, 'DI') }}
	RET
{% else -%}
TEXT ppapi·{{func.name}}(SB),NOSPLIT,${{asize}}
	MOVL	DI, AX
	MOVL	AX, {{asize-4}}(SP)
	{{- amd64args(func.args) }}
	MOVL	ppapi·ppb_interfaces+({{func.interface}}*8+4)(SB), AX
	MOVL	({{func.index}}*4)(AX), AX
	CALL	AX
	{% if func.structReturn -%}
	// Struct returned as *return_struct (sizeof {{func.result.size}}).
	{%- else -%}
	MOVL	{{asize-4}}(SP), DI
	{{ amd64return(func.result, align(fsize, 8), 'DI') }}
	{%- endif %}
	RET
{% endif %}
{% endfor %}

// ppapi·ppp_initialize_module_handler is called once at initialization
// initialization time.  Called on the C stack.
TEXT ppapi·ppp_initialize_module_handler(SB),NOSPLIT,$24
	MOVQ	R13, 0(SP)
	MOVQ	R14, 8(SP)
	MOVL	DI, ppapi·module_id(SB)  // module_id
	MOVL	SI, R14  // get_interface
	LEAL	ppapi·ppb_interfaces(SB), R13
initialize_module_loop:
	MOVL	0(R13), DI  // name
	TESTL	DI, DI
	JZ	initialize_module_done
	CALL	R14
	MOVL	AX, 4(R13)  // ppb
	ADDL	$8, R13
	JMP	initialize_module_loop
initialize_module_done:
	MOVQ	0(SP), R13
	MOVQ	8(SP), R14
	XORL	AX, AX
	RET

// ppapi·ppp_shutdown_module_handler may or may not be called when the
// module is closed.  Ignore the callback.  Called on the C stack.
TEXT ppapi·ppp_shutdown_module_handler(SB),NOSPLIT,$0
	RET

// ppapi·ppp_get_interface_handler is called by the browser to get
// callback functions.  Called on the C stack.
TEXT ppapi·ppp_get_interface_handler(SB),NOSPLIT,$40
	MOVQ	BX, 4(SP)
	MOVQ	R12, 12(SP)
	MOVQ	R13, 20(SP)
	MOVQ	R14, 28(SP)
	MOVL	DI, 0(SP)  // interface_name
	CALL	ppapi·ppp_get_interface(SB)
	MOVQ	4(SP), BX
	MOVQ	12(SP), R12
	MOVQ	20(SP), R13
	MOVQ	28(SP), R14
	RET

// ppapi·start is called to start PPAPI.  Never returns.
// Called on the C stack.
TEXT ppapi·start(SB),NOSPLIT,$8
	LEAL	0(SP), BP
	LEAL	ppapi·pp_start_functions(SB), DI
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

TEXT ·ppb_graphics2d_createx(SB),NOSPLIT,$8
	GO_ARGS
	MOVL	$22, rval+24(FP)
	RET

	
