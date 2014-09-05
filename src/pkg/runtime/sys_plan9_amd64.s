// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "zasm_GOOS_GOARCH.h"
#include "textflag.h"

// setldt(int entry, int address, int limit)
TEXT runtime·setldt(SB),NOSPLIT,$0
	RET

TEXT runtime·open(SB),NOSPLIT,$0
	MOVQ	$14, BP
	SYSCALL
	MOVL	AX, ret+16(FP)
	RET

TEXT runtime·pread(SB),NOSPLIT,$0
	MOVQ	$50, BP
	SYSCALL
	MOVL	AX, ret+32(FP)
	RET

TEXT runtime·pwrite(SB),NOSPLIT,$0
	MOVQ	$51, BP
	SYSCALL
	MOVL	AX, ret+32(FP)
	RET

// int32 _seek(int64*, int32, int64, int32)
TEXT _seek<>(SB),NOSPLIT,$0
	MOVQ	$39, BP
	SYSCALL
	RET

// int64 seek(int32, int64, int32)
// Convenience wrapper around _seek, the actual system call.
TEXT runtime·seek(SB),NOSPLIT,$32
	LEAQ	ret+24(FP), AX
	MOVL	fd+0(FP), BX
	MOVQ	offset+8(FP), CX
	MOVL	whence+16(FP), DX
	MOVQ	AX, 0(SP)
	MOVL	BX, 8(SP)
	MOVQ	CX, 16(SP)
	MOVL	DX, 24(SP)
	CALL	_seek<>(SB)
	CMPL	AX, $0
	JGE	2(PC)
	MOVQ	$-1, ret+24(FP)
	RET

TEXT runtime·close(SB),NOSPLIT,$0
	MOVQ	$4, BP
	SYSCALL
	MOVL	AX, ret+8(FP)
	RET

TEXT runtime·exits(SB),NOSPLIT,$0
	MOVQ	$8, BP
	SYSCALL
	RET

TEXT runtime·brk_(SB),NOSPLIT,$0
	MOVQ	$24, BP
	SYSCALL
	MOVQ	AX, ret+8(FP)
	RET

TEXT runtime·sleep(SB),NOSPLIT,$0
	MOVQ	$17, BP
	SYSCALL
	MOVL	AX, ret+8(FP)
	RET

TEXT runtime·plan9_semacquire(SB),NOSPLIT,$0
	MOVQ	$37, BP
	SYSCALL
	MOVL	AX, ret+16(FP)
	RET

TEXT runtime·plan9_tsemacquire(SB),NOSPLIT,$0
	MOVQ	$52, BP
	SYSCALL
	MOVL	AX, ret+16(FP)
	RET

TEXT runtime·nsec(SB),NOSPLIT,$0
	MOVQ	$53, BP
	SYSCALL
	MOVQ	AX, ret+8(FP)
	RET

TEXT runtime·notify(SB),NOSPLIT,$0
	MOVQ	$28, BP
	SYSCALL
	MOVL	AX, ret+8(FP)
	RET

TEXT runtime·noted(SB),NOSPLIT,$0
	MOVQ	$29, BP
	SYSCALL
	MOVL	AX, ret+8(FP)
	RET
	
TEXT runtime·plan9_semrelease(SB),NOSPLIT,$0
	MOVQ	$38, BP
	SYSCALL
	MOVL	AX, ret+16(FP)
	RET

TEXT runtime·rfork(SB),NOSPLIT,$0
	MOVQ	$19, BP // rfork
	SYSCALL

	// In parent, return.
	CMPQ	AX, $0
	JEQ	3(PC)
	MOVL	AX, ret+40(FP)
	RET

	// In child on forked stack.
	MOVQ	mm+24(SP), BX	// m
	MOVQ	gg+32(SP), DX	// g
	MOVQ	fn+40(SP), SI	// fn

	// set SP to be on the new child stack
	MOVQ	stack+16(SP), CX
	MOVQ	CX, SP

	// Initialize m, g.
	get_tls(AX)
	MOVQ	DX, g(AX)
	MOVQ	BX, g_m(DX)

	// Initialize procid from TOS struct.
	MOVQ	_tos(SB), AX
	MOVQ	64(AX), AX
	MOVQ	AX, m_procid(BX)	// save pid as m->procid
	
	CALL	runtime·stackcheck(SB)	// smashes AX, CX
	
	MOVQ	0(DX), DX	// paranoia; check they are not nil
	MOVQ	0(BX), BX
	
	CALL	SI	// fn()
	CALL	runtime·exit(SB)
	MOVL	AX, ret+40(FP)
	RET

// This is needed by asm_amd64.s
TEXT runtime·settls(SB),NOSPLIT,$0
	RET

// void sigtramp(void *ureg, int8 *note)
TEXT runtime·sigtramp(SB),NOSPLIT,$0
	get_tls(AX)

	// check that g exists
	MOVQ	g(AX), BX
	CMPQ	BX, $0
	JNE	3(PC)
	CALL	runtime·badsignal2(SB) // will exit
	RET

	// save args
	MOVQ	ureg+8(SP), CX
	MOVQ	note+16(SP), DX

	// change stack
	MOVQ	g_m(BX), BX
	MOVQ	m_gsignal(BX), R10
	MOVQ	g_stackbase(R10), BP
	MOVQ	BP, SP

	// make room for args and g
	SUBQ	$40, SP

	// save g
	MOVQ	g(AX), BP
	MOVQ	BP, 32(SP)

	// g = m->gsignal
	MOVQ	R10, g(AX)

	// load args and call sighandler
	MOVQ	CX, 0(SP)
	MOVQ	DX, 8(SP)
	MOVQ	BP, 16(SP)

	CALL	runtime·sighandler(SB)
	MOVL	24(SP), AX

	// restore g
	get_tls(BX)
	MOVQ	32(SP), R10
	MOVQ	R10, g(BX)

	// call noted(AX)
	MOVQ	AX, 0(SP)
	CALL	runtime·noted(SB)
	RET

TEXT runtime·setfpmasks(SB),NOSPLIT,$8
	STMXCSR	0(SP)
	MOVL	0(SP), AX
	ANDL	$~0x3F, AX
	ORL	$(0x3F<<7), AX
	MOVL	AX, 0(SP)
	LDMXCSR	0(SP)
	RET

#define ERRMAX 128	/* from os_plan9.h */

// void errstr(int8 *buf, int32 len)
TEXT errstr<>(SB),NOSPLIT,$0
	MOVQ    $41, BP
	SYSCALL
	RET

// func errstr() string
// Only used by package syscall.
// Grab error string due to a syscall made
// in entersyscall mode, without going
// through the allocator (issue 4994).
// See ../syscall/asm_plan9_amd64.s:/·Syscall/
TEXT runtime·errstr(SB),NOSPLIT,$16-16
	get_tls(AX)
	MOVQ	g(AX), BX
	MOVQ	g_m(BX), BX
	MOVQ	m_errstr(BX), CX
	MOVQ	CX, 0(SP)
	MOVQ	$ERRMAX, 8(SP)
	CALL	errstr<>(SB)
	CALL	runtime·findnull(SB)
	MOVQ	8(SP), AX
	MOVQ	AX, ret_len+8(FP)
	MOVQ	0(SP), AX
	MOVQ	AX, ret_base+0(FP)
	RET
