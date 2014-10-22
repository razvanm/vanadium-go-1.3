// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// NaCl system calls are mediated through an "outer sandbox."  System calls are
// internally implemented by branching to a stub in the system call table
// located at address 0x10000+32*number.
//
// However, the system call interface isn't part of the public NaCl API, and it
// is subject to change.  The officially supported interface is the NaCl IRT
// (for "integrated runtime"), which must be manually linked.  If the IRT is
// available, a lookup function is passed via the Elf aux vector, for fetching
// the function pointers corresponding to each of the IRT functions.
//
// For now, we support both mechanisms.	 If the IRT is available, we use it.
// Otherwise, we invoke the raw system calls directly.	The two interfaces are
// similar; the IRT is really just a thin wrapper around the system calls,
// changing the signature somewhat.  For example, the raw system calls use
// return values to pass back results in some cases.
//
//	 // Raw syscall.
//	 int fd = NACL_SYSCALL(SYS_open)(pathname, oflag, mode);
//	 if (fd < 0) { ...error... }
//
// In contrast, the IRT uses return values only for error codes.  Results are
// returned through result parameters.
//
//	 // IRT syscall.
//	 int fd;
//	 int code = (*irt_filename[IRT_FILENAME_OPEN])(pathname, oflag, mode, &fd);
//	 if (code != 0) { ...error... }
//
// Whether raw or IRT, the system call that is invoked runs in the same address
// space as the caller, using the caller's stack.  Since the stack requirements
// for the syscall aren't known, the wrappers here switch from the goroutine
// stack to the main thread stack, invoke the system call, then switch back to
// the original stack.	The basic form of these wrappers is as follows.
//
//	 switch to main thread stack
//	 if !irt_is_enabled
//	     invoke raw syscall
//	 else
//	     invoke IRT syscall
//	 switch back to original stack
//
// One of the reasons for supporting raw syscalls is for the Go playground.  At
// some point, we may want to consider requiring the IRT for the playground,
// removing support for raw syscalls.

#include "zasm_GOOS_GOARCH.h"
#include "textflag.h"
#include "syscall_nacl.h"
#include "irt_nacl.h"

#define NACL_SYSCALL(code) \
	MOVL $(0x10000 + ((code)<<5)), AX; CALL AX

// Swap stacks so that the caller runs on the main m->g0 stack.
// On return, SP is the main m->g0 stack, BP holds the original stack frame, and
// the current stack contains contains the original SP at 0(SP).
TEXT runtime·nacl_swapstack(SB),NOSPLIT,$0
	LEAL	4(SP), BP

	get_tls(CX)
	CMPL	CX, $0
	JE	nss_return  // Not a Go-managed thread. Do not switch stack.

	MOVL	g(CX), DI
	MOVL	g_m(DI), DI
	MOVL	m_g0(DI), SI
	CMPL	g(CX), SI
	JE	nss_return  // Executing on m->g0 already.

	// Switch to m->g0 stack.
	MOVL	(g_sched+gobuf_sp)(SI), SI
	LEAL	-4(SI), SP

nss_return:
	MOVL	-4(BP), AX
	MOVL	BP, (SP)
	ADDL	$4, BP
	JMP	AX

// Restore the original stack.	Must not modify AX.
TEXT runtime·nacl_restorestack(SB),NOSPLIT,$0
	MOVL	(SP), BX
	MOVL	4(SP), SP
	JMP	BX

// Begin a NaCl IRT call.  Execute call on the main m->g0 stack.
// On return, SP is the main m->g0 stack, BP holds the original stack frame, and
// the current stack contains the original SP at 0(SP).	 Call runtime·entersyscall
// only if the stack is switched and the current P is set.
TEXT runtime·nacl_entersyscall(SB),NOSPLIT,$0-0
	MOVL	$0, AX	// AX is set to true only if runtime·entersyscall was called.
	MOVL	SP, BP

	get_tls(CX)
	TESTL	CX, CX
	JZ	nen_return  // Not a Go-managed thread. Do not switch stack.

	MOVL	g(CX), DI
	MOVL	g_m(DI), DI
	MOVL	m_g0(DI), SI
	CMPL	g(CX), SI
	JE	nen_return  // Executing on m->g0 already.

	MOVL	m_p(DI), DX
	TESTL	DX, DX
	JZ	nen_swapstack  // Not a goroutine.

	CALL	runtime·entersyscall(SB)
	MOVL	$1, AX
	MOVL	SP, BP
	get_tls(CX)
	MOVL	g(CX), DI
	MOVL	g_m(DI), DI
	MOVL	m_g0(DI), SI

nen_swapstack:
	// Switch to m->g0 stack.
	MOVL	(g_sched+gobuf_sp)(SI), SI
	LEAL	-4(SI), SP

nen_return:
	SUBL	$8, SP
	MOVL	(BP), CX
	MOVL	BP, (SP)
	MOVL	AX, 4(SP)
	ADDL	$8, BP
	JMP	CX

// Finish a NaCl IRT call.  Restores the stack, and calls runtime·exitsyscall,
// but only if runtime·entersyscall was called on entry.  Must not modify AX.
TEXT runtime·nacl_exitsyscall(SB),NOSPLIT,$0-0
	MOVL	8(SP), BX
	TESTL	BX, BX
	JZ	nex_return
	MOVL	(SP), BX
	MOVL	4(SP), SP
	MOVL	BX, (SP)
	CALL	runtime·nacl_wrap_exitsyscall(SB)
	RET
nex_return:	
	MOVL	(SP), BX
	MOVL	4(SP), SP
	MOVL	BX, (SP)
	RET

TEXT runtime·nacl_wrap_exitsyscall(SB),NOSPLIT,$4
	MOVL	AX, (SP)
	CALL	runtime·exitsyscall(SB)
	MOVL	(SP), AX
	RET

TEXT runtime·exit(SB),NOSPLIT,$0
	CALL	runtime·nacl_swapstack(SB)
	SUBL	$4, SP
	MOVL	0(BP), AX
	MOVL	AX, 0(SP)
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	exit_irt
	NACL_SYSCALL(SYS_exit)
	MOVL	$0x13, 0x13  // crash
exit_irt:
	MOVL	runtime·nacl_irt_basic_v0_1+(IRT_BASIC_EXIT*4)(SB), AX
	CALL	AX
	MOVL	$0x14, 0x14  // crash

TEXT runtime·exit1(SB),NOSPLIT,$0
	CALL	runtime·nacl_swapstack(SB)
	SUBL	$4, SP
	MOVL	0(BP), AX
	MOVL	AX, 0(SP)
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	exit1_irt
	NACL_SYSCALL(SYS_thread_exit)
	MOVL	$0x15, 0x15 // crash
exit1_irt:
	MOVL	runtime·nacl_irt_thread_v0_1+(IRT_THREAD_EXIT*4)(SB), AX
	CALL	AX
	MOVL	$0x16, 0x16 // crash

TEXT runtime·open(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	open_irt
	SUBL	$12, SP
	MOVL	0(BP), AX  // pathname
	MOVL	AX, 0(SP)
	MOVL	4(BP), AX  // oflag
	MOVL	AX, 4(SP)
	MOVL	8(BP), AX  // cmode
	MOVL	AX, 8(SP)
	NACL_SYSCALL(SYS_open)
	ADDL	$12, SP
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+12(FP)
	RET
open_irt:
	SUBL	$20, SP
	MOVL	0(BP), AX  // pathname
	MOVL	AX, 0(SP)
	MOVL	4(BP), AX  // oflag
	MOVL	AX, 4(SP)
	MOVL	8(BP), AX  // cmode
	MOVL	AX, 8(SP)
	LEAL	16(SP), AX  // &fd
	MOVL	AX, 12(SP)
	MOVL	runtime·nacl_irt_filename_v0_1+(IRT_FILENAME_OPEN*4)(SB), AX
	CALL	AX
	NEGL	AX
	JNZ	open_done
	MOVL	16(SP), AX
open_done:
	ADDL	$20, SP
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+12(FP)
	RET

TEXT runtime·close(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	SUBL	$4, SP
	MOVL	0(BP), AX
	MOVL	AX, 0(SP)
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	close_irt
	NACL_SYSCALL(SYS_close)
	JMP	close_done
close_irt:
	MOVL	runtime·nacl_irt_fdio_v0_1+(IRT_FDIO_CLOSE*4)(SB), AX
	CALL	AX
	NEGL	AX
close_done:
	ADDL	$4, SP
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+4(FP)
	RET

TEXT runtime·read(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	read_irt
	SUBL	$12, SP
	MOVL	0(BP), AX  // fd
	MOVL	AX, 0(SP)
	MOVL	4(BP), AX  // buf
	MOVL	AX, 4(SP)
	MOVL	8(BP), AX  // count
	MOVL	AX, 8(SP)
	NACL_SYSCALL(SYS_read)
	ADDL	$12, SP
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+12(FP)
	RET
read_irt:
	SUBL	$20, SP
	MOVL	0(BP), AX  // fd
	MOVL	AX, 0(SP)
	MOVL	4(BP), AX  // buf
	MOVL	AX, 4(SP)
	MOVL	8(BP), AX  // count
	MOVL	AX, 8(SP)
	LEAL	16(SP), AX  // nread
	MOVL	AX, 12(SP)
	MOVL	runtime·nacl_irt_fdio_v0_1+(IRT_FDIO_READ*4)(SB), AX
	CALL	AX
	NEGL	AX
	JNZ	read_done
	MOVL	16(SP), AX
read_done:
	ADDL	$20, SP
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+12(FP)
	RET

TEXT syscall·naclWrite(SB), NOSPLIT, $16-16
	MOVL arg1+0(FP), DI
	MOVL arg2+4(FP), SI
	MOVL arg3+8(FP), DX
	MOVL DI, 0(SP)
	MOVL SI, 4(SP)
	MOVL DX, 8(SP)
	CALL runtime·write(SB)
	MOVL AX, ret+16(FP)
	RET

TEXT runtime·write(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	write_irt
	SUBL	$12, SP
	MOVL	0(BP), AX  // fd
	MOVL	AX, 0(SP)
	MOVL	4(BP), AX  // buf
	MOVL	AX, 4(SP)
	MOVL	8(BP), AX  // count
	MOVL	AX, 8(SP)
	NACL_SYSCALL(SYS_write)
	ADDL	$12, SP
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+12(FP)
	RET
write_irt:
	SUBL	$20, SP
	MOVL	0(BP), AX  // fd
	MOVL	AX, 0(SP)
	MOVL	4(BP), AX  // buf
	MOVL	AX, 4(SP)
	MOVL	8(BP), AX  // count
	MOVL	AX, 8(SP)
	LEAL	16(SP), AX  // nwrite
	MOVL	AX, 12(SP)
	MOVL	runtime·nacl_irt_fdio_v0_1+(IRT_FDIO_WRITE*4)(SB), AX
	CALL	AX
	NEGL	AX
	JNZ	write_done
	MOVL	16(SP), AX
write_done:
	ADDL	$20, SP
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+12(FP)
	RET

TEXT runtime·nacl_exception_stack(SB),NOSPLIT,$0
	CALL	runtime·nacl_swapstack(SB)
	SUBL	$8, SP
	MOVL	0(BP), AX  // stack
	MOVL	AX, 0(SP)
	MOVL	4(BP), AX  // size
	MOVL	AX, 4(SP)
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	nacl_exception_stack_irt
	NACL_SYSCALL(SYS_exception_stack)
	JMP	nacl_exception_stack_done
nacl_exception_stack_irt:
	MOVL	runtime·nacl_irt_exception_handling_v0_1+(IRT_EXCEPTION_STACK*4)(SB), AX
	CALL	AX
	NEGL	AX
nacl_exception_stack_done:
	ADDL	$8, SP
	CALL	runtime·nacl_restorestack(SB)
	MOVL	AX, ret+8(FP)
	RET

TEXT runtime·nacl_exception_handler(SB),NOSPLIT,$0
	CALL	runtime·nacl_swapstack(SB)
	SUBL	$8, SP
	MOVL	0(BP), AX  // handler
	MOVL	AX, 0(SP)
	MOVL	4(BP), AX  // *old_handler
	MOVL	AX, 4(SP)
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	nacl_exception_handler_irt
	NACL_SYSCALL(SYS_exception_handler)
	JMP	nacl_exception_handler_done
nacl_exception_handler_irt:
	MOVL	runtime·nacl_irt_exception_handling_v0_1+(IRT_EXCEPTION_HANDLER*4)(SB), AX
	CALL	AX
	NEGL	AX
nacl_exception_handler_done:
	ADDL	$8, SP
	CALL	runtime·nacl_restorestack(SB)
	MOVL	AX, ret+8(FP)
	RET

TEXT runtime·nacl_sem_create(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	sem_create_irt
	SUBL	$4, SP
	MOVL	0(BP), AX  // value
	MOVL	AX, 0(SP)
	NACL_SYSCALL(SYS_sem_create)
	ADDL	$4, SP
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+4(FP)
	RET
sem_create_irt:
	SUBL	$12, SP
	LEAL	8(SP), AX  // *sem_handle
	MOVL	AX, 0(SP)
	MOVL	4(BP), AX  // value
	MOVL	AX, 4(SP)
	MOVL	runtime·nacl_irt_sem_v0_1+(IRT_SEM_CREATE*4)(SB), AX
	CALL	AX
	NEGL	AX
	JNZ	sem_create_done
	MOVL	8(SP), AX
sem_create_done:
	ADDL	$12, SP
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+4(FP)
	RET

TEXT runtime·nacl_sem_wait(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	SUBL	$4, SP
	MOVL	0(BP), AX  // sem_handle
	MOVL	AX, 0(SP)
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	sem_wait_irt
	NACL_SYSCALL(SYS_sem_wait)
	JMP	sem_wait_done
sem_wait_irt:
	MOVL	runtime·nacl_irt_sem_v0_1+(IRT_SEM_WAIT*4)(SB), AX
	CALL	AX
	NEGL	AX
sem_wait_done:
	ADDL	$4, SP
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+4(FP)
	RET

TEXT runtime·nacl_sem_post(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	SUBL	$4, SP
	MOVL	0(BP), AX  // sem_handle
	MOVL	AX, 0(SP)
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	sem_post_irt
	NACL_SYSCALL(SYS_sem_post)
	JMP	sem_post_done
sem_post_irt:
	MOVL	runtime·nacl_irt_sem_v0_1+(IRT_SEM_POST*4)(SB), AX
	CALL	AX
	NEGL	AX
sem_post_done:
	ADDL	$4, SP
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+4(FP)
	RET

TEXT runtime·nacl_mutex_create(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	mutex_create_irt
	SUBL	$4, SP
	MOVL	0(BP), AX  // flag
	MOVL	AX, (SP)
	NACL_SYSCALL(SYS_mutex_create)
	ADDL	$4, SP
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+4(FP)
	RET
mutex_create_irt:
	SUBL	$8, SP
	LEAL	4(SP), AX  // *mutex_handle
	MOVL	AX, 0(SP)
	MOVL	runtime·nacl_irt_mutex_v0_1+(IRT_MUTEX_CREATE*4)(SB), AX
	CALL	AX
	NEGL	AX
	JNZ	mutex_create_done
	MOVL	4(SP), AX
mutex_create_done:
	ADDL	$8, SP
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+4(FP)
	RET

TEXT runtime·nacl_mutex_lock(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	SUBL	$4, SP
	MOVL	0(BP), AX  // mutex_handle
	MOVL	AX, 0(SP)
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	mutex_lock_irt
	NACL_SYSCALL(SYS_mutex_lock)
	JMP	mutex_lock_done
mutex_lock_irt:
	MOVL	runtime·nacl_irt_mutex_v0_1+(IRT_MUTEX_LOCK*4)(SB), AX
	CALL	AX
	NEGL	AX
mutex_lock_done:
	ADDL	$4, SP
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+4(FP)
	RET

TEXT runtime·nacl_mutex_trylock(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	SUBL	$4, SP
	MOVL	0(BP), AX  // mutex_handle
	MOVL	AX, 0(SP)
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	mutex_trylock_irt
	NACL_SYSCALL(SYS_mutex_trylock)
	JMP	mutex_trylock_done
mutex_trylock_irt:
	MOVL	runtime·nacl_irt_mutex_v0_1+(IRT_MUTEX_TRYLOCK*4)(SB), AX
	CALL	AX
	NEGL	AX
mutex_trylock_done:
	ADDL	$4, SP
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+4(FP)
	RET

TEXT runtime·nacl_mutex_unlock(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	SUBL	$4, SP
	MOVL	0(BP), AX  // mutex_handle
	MOVL	AX, 0(SP)
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	mutex_unlock_irt
	NACL_SYSCALL(SYS_mutex_unlock)
	JMP	mutex_unlock_done
mutex_unlock_irt:
	MOVL	runtime·nacl_irt_mutex_v0_1+(IRT_MUTEX_UNLOCK*4)(SB), AX
	CALL	AX
	NEGL	AX
mutex_unlock_done:
	ADDL	$4, SP
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+4(FP)
	RET

TEXT runtime·nacl_cond_create(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	cond_create_irt
	SUBL	$4, SP
	MOVL	0(BP), AX  // flag
	MOVL	AX, (SP)
	NACL_SYSCALL(SYS_cond_create)
	ADDL	$4, SP
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+4(FP)
	RET
cond_create_irt:
	SUBL	$8, SP
	LEAL	4(SP), AX  // *cond_handle
	MOVL	AX, 0(SP)
	MOVL	runtime·nacl_irt_cond_v0_1+(IRT_COND_CREATE*4)(SB), AX
	CALL	AX
	NEGL	AX
	JNZ	cond_create_done
	MOVL	4(SP), AX
cond_create_done:
	ADDL	$8, SP
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+4(FP)
	RET

TEXT runtime·nacl_cond_wait(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	SUBL	$8, SP
	MOVL	0(BP), AX  // cond_handle
	MOVL	AX, 0(SP)
	MOVL	4(BP), AX  // mutex_handle
	MOVL	AX, 4(SP)
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	cond_wait_irt
	NACL_SYSCALL(SYS_cond_wait)
	JMP	cond_wait_done
cond_wait_irt:
	MOVL	runtime·nacl_irt_cond_v0_1+(IRT_COND_WAIT*4)(SB), AX
	CALL	AX
	NEGL	AX
cond_wait_done:
	ADDL	$8, SP
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+8(FP)
	RET

TEXT runtime·nacl_cond_signal(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	SUBL	$4, SP
	MOVL	0(BP), AX  // cond_handle
	MOVL	AX, 0(SP)
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	cond_signal_irt
	NACL_SYSCALL(SYS_cond_signal)
	JMP	cond_signal_done
cond_signal_irt:
	MOVL	runtime·nacl_irt_cond_v0_1+(IRT_COND_SIGNAL*4)(SB), AX
	CALL	AX
	NEGL	AX
cond_signal_done:
	ADDL	$4, SP
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+4(FP)
	RET

TEXT runtime·nacl_cond_broadcast(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	SUBL	$4, SP
	MOVL	0(BP), AX  // cond_handle
	MOVL	AX, 0(SP)
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	cond_broadcast_irt
	NACL_SYSCALL(SYS_cond_broadcast)
	JMP	cond_broadcast_done
cond_broadcast_irt:
	MOVL	runtime·nacl_irt_cond_v0_1+(IRT_COND_BROADCAST*4)(SB), AX
	CALL	AX
	NEGL	AX
cond_broadcast_done:
	ADDL	$4, SP
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+4(FP)
	RET

TEXT runtime·nacl_cond_timed_wait_abs(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	SUBL	$12, SP
	MOVL	0(BP), AX  // cond_handle
	MOVL	AX, 0(SP)
	MOVL	4(BP), AX  // mutex_handle
	MOVL	AX, 4(SP)
	MOVL	8(BP), AX  // abstime
	MOVL	AX, 8(SP)
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	cond_timed_wait_abs_irt
	NACL_SYSCALL(SYS_cond_timed_wait_abs)
	JMP	cond_timed_wait_abs_done
cond_timed_wait_abs_irt:
	MOVL	runtime·nacl_irt_cond_v0_1+(IRT_COND_TIMED_WAIT_ABS*4)(SB), AX
	CALL	AX
	NEGL	AX
cond_timed_wait_abs_done:
	ADDL	$12, SP
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+12(FP)
	RET

TEXT runtime·nacl_thread_create(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	thread_create_irt
	SUBL	$16, SP
	MOVL	0(BP), AX  // start_func
	MOVL	AX, 0(SP)
	MOVL	4(BP), AX  // stack
	MOVL	AX, 4(SP)
	MOVL	8(BP), AX  // tls
	MOVL	AX, 8(SP)
	MOVL	12(BP), AX  // tp
	MOVL	AX, 12(SP)
	NACL_SYSCALL(SYS_thread_create)
	ADDL	$16, SP
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+16(FP)
	RET
thread_create_irt:
	SUBL	$12, SP
	MOVL	0(BP), AX  // start_func
	MOVL	AX, 0(SP)
	MOVL	4(BP), AX  // stack
	MOVL	AX, 4(SP)
	MOVL	8(BP), AX  // thread_ptr
	MOVL	AX, 8(SP)
	MOVL	runtime·nacl_irt_thread_v0_1+(IRT_THREAD_CREATE*4)(SB), AX
	CALL	AX
	NEGL	AX
	ADDL	$12, SP
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+16(FP)
	RET

TEXT runtime·mstart_nacl(SB),NOSPLIT,$0
	JMP runtime·mstart(SB)

TEXT runtime·nacl_nanosleep(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	SUBL	$8, SP
	MOVL	0(BP), AX  // req
	MOVL	AX, 0(SP)
	MOVL	4(BP), AX  // rem
	MOVL	AX, 4(SP)
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	nanosleep_irt
	NACL_SYSCALL(SYS_nanosleep)
	JMP	nanosleep_done
nanosleep_irt:
	MOVL	runtime·nacl_irt_basic_v0_1+(IRT_BASIC_NANOSLEEP*4)(SB), AX
	CALL	AX
	NEGL	AX
nanosleep_done:
	ADDL	$8, SP
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+8(FP)
	RET

TEXT runtime·osyield(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	osyield_irt
	NACL_SYSCALL(SYS_sched_yield)
	JMP	osyield_done
osyield_irt:
	MOVL	runtime·nacl_irt_basic_v0_1+(IRT_BASIC_SCHED_YIELD*4)(SB), AX
	CALL	AX
osyield_done:
	CALL	runtime·nacl_exitsyscall(SB)
	RET

TEXT runtime·mmap(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	SUBL	$32, SP
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	mmap_irt
	MOVL	0(BP), AX  // addr
	MOVL	AX, 0(SP)
	MOVL	4(BP), AX  // len
	MOVL	AX, 4(SP)
	MOVL	8(BP), AX  // prot
	MOVL	AX, 8(SP)
	MOVL	12(BP), AX  // flags
	MOVL	AX, 12(SP)
	MOVL	16(BP), AX  // fd
	MOVL	AX, 16(SP)
	MOVL	20(BP), AX  // off
	MOVL	AX, 24(SP)
	MOVL	$0, 28(SP)  // sign-extend
	LEAL	24(SP), AX  // &off
	MOVL	AX, 20(SP)
	NACL_SYSCALL(SYS_mmap)
	JMP	mmap_done
mmap_irt:
	MOVL	0(BP), AX  // addr
	MOVL	AX, 28(SP)
	LEAL	28(SP), AX  // &addr
	MOVL	AX, 0(SP)
	MOVL	4(BP), AX  // len
	MOVL	AX, 4(SP)
	MOVL	8(BP), AX  // prot
	MOVL	AX, 8(SP)
	MOVL	12(BP), AX  // flags
	MOVL	AX, 12(SP)
	MOVL	16(BP), AX  // fd
	MOVL	AX, 16(SP)
	MOVL	20(BP), AX  // off
	MOVL	AX, 20(SP)
	MOVL	$0, 24(SP)  // sign-extend
	MOVL	runtime·nacl_irt_memory_v0_3+(IRT_MEMORY_MMAP*4)(SB), AX
	CALL	AX
	NEGL	AX
	JNZ	mmap_done
	MOVL	28(SP), AX
mmap_done:
	ADDL	$32, SP
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+24(FP)
	RET

TEXT time·now(SB),NOSPLIT,$0
	CALL	runtime·nacl_swapstack(SB)
	SUBL	$24, SP
	MOVL	BP, 20(SP)
	MOVL	$0, 0(SP) // real time clock
	LEAL	8(SP), AX
	MOVL	AX, 4(SP) // timespec
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	now_irt
	NACL_SYSCALL(SYS_clock_gettime)
	JMP	now_done
now_irt:
	MOVL	runtime·nacl_irt_clock_v0_1+(IRT_CLOCK_GETTIME*4)(SB), AX
	CALL	AX
now_done:
	MOVL	8(SP), AX // low 32 sec
	MOVL	12(SP), CX // high 32 sec
	MOVL	16(SP), BX // nsec

	// sec is in AX, nsec in BX
	MOVL	20(SP), BP
	MOVL	AX, 0(BP)
	MOVL	CX, 4(BP)
	MOVL	BX, 8(BP)
	ADDL	$24, SP
	CALL	runtime·nacl_restorestack(SB)
	RET

TEXT syscall·now(SB),NOSPLIT,$0
	JMP time·now(SB)

TEXT runtime·nacl_clock_gettime(SB),NOSPLIT,$0
	CALL	runtime·nacl_swapstack(SB)
	SUBL	$8, SP
	MOVL	0(BP), AX  // clk_id
	MOVL	AX, 0(SP)
	MOVL	4(BP), AX  // *tp
	MOVL	AX, 4(SP)
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	clock_gettime_irt
	NACL_SYSCALL(SYS_clock_gettime)
	JMP	clock_gettime_done
clock_gettime_irt:
	MOVL	runtime·nacl_irt_clock_v0_1+(IRT_CLOCK_GETTIME*4)(SB), AX
	CALL	AX
	NEGL	AX
clock_gettime_done:
	ADDL	$8, SP
	CALL	runtime·nacl_restorestack(SB)
	MOVL	AX, ret+8(FP)
	RET

TEXT runtime·nanotime(SB),NOSPLIT,$0
	CALL	runtime·nacl_swapstack(SB)
	SUBL	$20, SP
	MOVL	$0, 0(SP) // real time clock
	LEAL	8(SP), AX
	MOVL	AX, 4(SP) // timespec
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	nanotime_irt
	NACL_SYSCALL(SYS_clock_gettime)
	JMP	nanotime_done
nanotime_irt:
	MOVL	runtime·nacl_irt_clock_v0_1+(IRT_CLOCK_GETTIME*4)(SB), AX
	CALL	AX
nanotime_done:
	MOVL	8(SP), AX // low 32 sec
	MOVL	16(SP), BX // nsec

	// sec is in AX, nsec in BX
	// convert to DX:AX nsec
	MOVL	$1000000000, CX
	MULL	CX
	ADDL	BX, AX
	ADCL	$0, DX
	ADDL	$20, SP
	CALL	runtime·nacl_restorestack(SB)
	MOVL	AX, 4(SP)
	MOVL	DX, 8(SP)
	RET

TEXT runtime·setldt(SB),NOSPLIT,$8
	MOVL	addr+4(FP), BX // aka base
	ADDL	$0x8, BX
	MOVL	BX, 0(SP)
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	setldt_irt
	NACL_SYSCALL(SYS_tls_init)
	RET
setldt_irt:
	MOVL	runtime·nacl_irt_tls_v0_1+(IRT_TLS_INIT*4)(SB), AX
	CALL	AX
	NEGL	AX
	RET

TEXT runtime·sigtramp(SB),NOSPLIT,$0
	get_tls(CX)

	// check that g exists
	MOVL	g(CX), DI
	CMPL	DI, $0
	JNE	6(PC)
	MOVL	$11, BX
	MOVL	$0, 0(SP)
	MOVL	$runtime·badsignal(SB), AX
	CALL	AX
	JMP	sigtramp_ret

	// save g
	MOVL	DI, 20(SP)

	// g = m->gsignal
	MOVL	g_m(DI), BX
	MOVL	m_gsignal(BX), BX
	MOVL	BX, g(CX)

	// copy arguments for sighandler
	MOVL	$11, 0(SP) // signal
	MOVL	$0, 4(SP) // siginfo
	LEAL	ctxt+4(FP), AX
	MOVL	AX, 8(SP) // context
	MOVL	DI, 12(SP) // g

	CALL	runtime·sighandler(SB)

	// restore g
	get_tls(CX)
	MOVL	20(SP), BX
	MOVL	BX, g(CX)

sigtramp_ret:
	// Enable exceptions again.
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	sigtramp_irt
	NACL_SYSCALL(SYS_exception_clear_flag)
	JMP	sigtramp_ret2
sigtramp_irt:
	MOVL	runtime·nacl_irt_exception_handling_v0_1+(IRT_EXCEPTION_STACK*4)(SB), AX
	CALL	AX
sigtramp_ret2:

	// NaCl has abdicated its traditional operating system responsibility
	// and declined to implement 'sigreturn'. Instead the only way to return
	// to the execution of our program is to restore the registers ourselves.
	// Unfortunately, that is impossible to do with strict fidelity, because
	// there is no way to do the final update of PC that ends the sequence
	// without either (1) jumping to a register, in which case the register ends
	// holding the PC value instead of its intended value or (2) storing the PC
	// on the stack and using RET, which imposes the requirement that SP is
	// valid and that is okay to smash the word below it. The second would
	// normally be the lesser of the two evils, except that on NaCl, the linker
	// must rewrite RET into "POP reg; AND $~31, reg; JMP reg", so either way
	// we are going to lose a register as a result of the incoming signal.
	// Similarly, there is no way to restore EFLAGS; the usual way is to use
	// POPFL, but NaCl rejects that instruction. We could inspect the bits and
	// execute a sequence of instructions designed to recreate those flag
	// settings, but that's a lot of work.
	//
	// Thankfully, Go's signal handlers never try to return directly to the
	// executing code, so all the registers and EFLAGS are dead and can be
	// smashed. The only registers that matter are the ones that are setting
	// up for the simulated call that the signal handler has created.
	// Today those registers are just PC and SP, but in case additional registers
	// are relevant in the future (for example DX is the Go func context register)
	// we restore as many registers as possible.
	//
	// We smash BP, because that's what the linker smashes during RET.
	//
	LEAL	ctxt+4(FP), BP
	ADDL	$64, BP
	MOVL	0(BP), AX
	MOVL	4(BP), CX
	MOVL	8(BP), DX
	MOVL	12(BP), BX
	MOVL	16(BP), SP
	// 20(BP) is saved BP, never to be seen again
	MOVL	24(BP), SI
	MOVL	28(BP), DI
	// 36(BP) is saved EFLAGS, never to be seen again
	MOVL	32(BP), BP // saved PC
	JMP	BP

#define AT_SYSINFO	32
TEXT runtime·nacl_sysinfo(SB),NOSPLIT,$16
	// nacl_irt_query is passed via Elf aux vector, which starts at
	// argv[argc + envc + 2];
	//
	// typedef struct {
	//   int32 a_type;	/* Entry type */
	//   union {
	//    int32 a_val;	/* Integer value */
	//   } a_un;
	// } Elf32_auxv_t;
	//
	MOVL	di+0(FP), DI
	LEAL	12(DI), BX	// argv
	MOVL	8(DI), AX	// argc
	ADDL	4(DI), AX	// envc
	ADDL	$2, AX
	LEAL	(BX)(AX*4), BX	// BX = &argv[argc + envc + 2]
	MOVL	BX, runtime·nacl_irt_query(SB)
auxloop:
	MOVL	0(BX), DX	// DX = BX->a_type
	CMPL	DX, $0
	JE	no_irt
	CMPL	DX, $AT_SYSINFO
	JEQ	auxfound
	ADDL	$8, BX
	JMP	auxloop
auxfound:
	MOVL	4(BX), BX
	MOVL	BX, runtime·nacl_irt_query(SB)
	LEAL	runtime·nacl_irt_entries(SB), BX
queryloop:
	MOVL	0(BX), AX	// name
	TESTL	AX, AX
	JE	irt_done
	MOVL	AX, 0(SP)
	MOVL	4(BX), AX	// funtab
	MOVL	AX, 4(SP)
	MOVL	8(BX), AX	// size
	MOVL	AX, 8(SP)
	ADDL	$16, BX
	MOVL	runtime·nacl_irt_query(SB), AX
	CALL	AX
	TESTL	AX, AX
	JNE	queryloop
	CMPL	-4(BX), $0	// is_required
	JE	queryloop
	CALL	runtime·crash(SB)
no_irt:
	MOVL	$0, runtime·nacl_irt_is_enabled(SB)
	RET
irt_done:
	MOVL	$1, runtime·nacl_irt_is_enabled(SB)
	RET
