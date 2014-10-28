// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "zasm_GOOS_GOARCH.h"
#include "textflag.h"
#include "syscall_nacl.h"
#include "irt_nacl.h"

#define NACL_SYSCALL(code) \
	MOVL $(0x10000 + ((code)<<5)), AX; CALL AX

// NFP is the "NaCl frame pointer."  We use it to refer to the caller's frame.
// The callee executes on the C stack, which is different from the caller's stack
// if the caller is a goroutine.
#define NFP	R14

TEXT runtime·nacl_swapstack(SB),NOSPLIT,$0
	LEAL	8(SP), NFP

	get_tls(CX)
	TESTL	CX, CX
	JE	nss_return  // Not a Go-managed thread. Do not switch stack.

	MOVL	g(CX), DI
	MOVL	g_m(DI), DI
	MOVL	m_g0(DI), SI
	CMPL	g(CX), SI
	JE	nss_return // executing on m->g0 already

	// Switch to m->g0 stack.
	MOVQ	(g_sched+gobuf_sp)(SI), SI
	LEAL	-8(SI), SP

nss_return:
	MOVL	-8(NFP), AX
	MOVL	NFP, (SP)
	ADDL	$8, NFP
	JMP	AX

// Restore the original stack.	Must not modify AX.
TEXT runtime·nacl_restorestack(SB),NOSPLIT,$0
	MOVL	(SP), BX
	MOVL	8(SP), SP
	JMP	BX

// Begin a NaCl IRT call.  Execute call on the main m->g0 stack.
// On return, SP is the main m->g0 stack, NFP holds the original stack frame, and
// the current stack contains the original SP at 0(SP).	 Call runtime·entersyscall
// only if the stack is switched and the current P is set.
TEXT runtime·nacl_entersyscall(SB),NOSPLIT,$0-0
	MOVL	$0, AX	// AX is set to true only if runtime·entersyscall was called.
	MOVL	SP, NFP

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
	MOVL	SP, NFP
	get_tls(CX)
	MOVL	g(CX), DI
	MOVL	g_m(DI), DI
	MOVL	m_g0(DI), SI

nen_swapstack:
	// Switch to m->g0 stack.
	MOVL	(g_sched+gobuf_sp)(SI), SI
	LEAL	-8(SI), SP

nen_return:
	SUBL	$8, SP
	MOVL	(NFP), CX
	MOVL	NFP, (SP)
	MOVL	AX, 4(SP)
	ADDL	$16, NFP
	JMP	CX

// Finish a NaCl IRT call.  Restores the stack, and calls runtime·exitsyscall,
// but only if runtime·entersyscall was called on entry.  Must not modify AX.
TEXT runtime·nacl_exitsyscall(SB),NOSPLIT,$0-0
	MOVL	12(SP), BX
	TESTL	BX, BX
	JZ	nex_return
	MOVL	(SP), BX
	MOVL	8(SP), SP
	MOVL	BX, (SP)
	PUSHQ	AX
	CALL	runtime·exitsyscall(SB)
	POPQ	AX
	RET
nex_return:
	MOVL	(SP), BX
	MOVL	8(SP), SP
	MOVL	BX, (SP)
	RET

TEXT runtime·settls(SB),NOSPLIT,$0
	MOVL	DI, TLS // really BP
	RET

TEXT runtime·exit(SB),NOSPLIT,$0
	CALL	runtime·nacl_swapstack(SB)
	MOVL	0(NFP), DI
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	exit_irt
	NACL_SYSCALL(SYS_exit)
	CALL	runtime·nacl_restorestack(SB)
	MOVL	$0x13, 0x13  // crash
exit_irt:
	MOVL	runtime·nacl_irt_basic_v0_1+(IRT_BASIC_EXIT*4)(SB), AX
	CALL	AX
	MOVL	$0x14, 0x14  // crash

TEXT runtime·exit1(SB),NOSPLIT,$0
	CALL	runtime·nacl_swapstack(SB)
	MOVL	0(NFP), DI  // exit code
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
	MOVL	0(NFP), DI  // pathname
	MOVL	4(NFP), SI  // oflag
	MOVL	8(NFP), DX  // cmode
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	open_irt
	NACL_SYSCALL(SYS_open)
	JMP	open_done
open_irt:
	SUBL	$4, SP
	MOVL	SP, CX	// result
	MOVL	runtime·nacl_irt_filename_v0_1+(IRT_FILENAME_OPEN*4)(SB), AX
	CALL	AX
	NEGL	AX
	JNZ	open_fail
	MOVL	(SP), AX
open_fail:
	ADDL	$4, SP
open_done:
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+16(FP)
	RET

TEXT runtime·close(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	MOVL	0(NFP), DI
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	close_irt
	NACL_SYSCALL(SYS_close)
	JMP	close_done
close_irt:
	MOVL	runtime·nacl_irt_fdio_v0_1+(IRT_FDIO_CLOSE*4)(SB), AX
	CALL	AX
	NEGL	AX
close_done:
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+8(FP)
	RET

TEXT runtime·read(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	MOVL	0(NFP), DI  // fd
	MOVL	4(NFP), SI  // p
	MOVL	8(NFP), DX  // n
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	read_irt
	NACL_SYSCALL(SYS_read)
	JMP	read_done
read_irt:
	SUBL	$4, SP
	MOVL	SP, CX	// result
	MOVL	runtime·nacl_irt_fdio_v0_1+(IRT_FDIO_READ*4)(SB), AX
	CALL	AX
	NEGL	AX
	JNZ	read_fail
	MOVL	(SP), AX
read_fail:
	ADDL	$4, SP
read_done:
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+16(FP)
	RET

TEXT syscall·naclWrite(SB), NOSPLIT, $24-20
	MOVL arg1+0(FP), DI
	MOVL arg2+4(FP), SI
	MOVL arg3+8(FP), DX
	MOVL DI, 0(SP)
	MOVL SI, 4(SP)
	MOVL DX, 8(SP)
	CALL runtime·write(SB)
	MOVL 16(SP), AX
	MOVL AX, ret+16(FP)
	RET

TEXT runtime·write(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	write_irt
	// If using fake time and writing to stdout or stderr,
	// emit playback header before actual data.
	MOVQ runtime·timens(SB), AX
	CMPQ AX, $0
	JEQ write
	MOVL 0(NFP), DI	 // fd
	CMPL DI, $1
	JEQ playback
	CMPL DI, $2
	JEQ playback

write:
	// Ordinary write.
	MOVL 0(NFP), DI	 // fd
	MOVL 4(NFP), SI	 // p
	MOVL 8(NFP), DX	 // n
	NACL_SYSCALL(SYS_write)
	JMP write_done

write_irt:
	// Write with IRT enabled.  We don't handle the playback header.
	SUBL	$4, SP
	MOVL	0(NFP), DI  // fd
	MOVL	4(NFP), SI  // p
	MOVL	8(NFP), DX  // n
	MOVL	SP, CX	// result
	MOVL	runtime·nacl_irt_fdio_v0_1+(IRT_FDIO_WRITE*4)(SB), AX
	CALL	AX
	NEGL	AX
	JNZ	write_irt_fail
	MOVL	(SP), AX
write_irt_fail:
	ADDL	$4, SP
	JMP	write_done

	// Write with playback header.
	// First, lock to avoid interleaving writes.
playback:
	SUBL $16, SP
spinlock:
	MOVL $1, BX
	XCHGL	runtime·writelock(SB), BX
	CMPL BX, $0
	JNE spinlock

	// Playback header: 0 0 P B <8-byte time> <4-byte data length>
	MOVL $(('B'<<24) | ('P'<<16)), 0(SP)
	BSWAPQ AX
	MOVQ AX, 4(SP)
	MOVL 8(NFP), DX	 // n
	BSWAPL DX
	MOVL DX, 12(SP)
	MOVL $1, DI // standard output
	MOVL SP, SI
	MOVL $16, DX
	NACL_SYSCALL(SYS_write)

	// Write actual data.
	MOVL $1, DI // standard output
	MOVL 4(NFP), SI	 // p
	MOVL 8(NFP), DX	 // n
	NACL_SYSCALL(SYS_write)
	ADDL $16, SP

	// Unlock.
	MOVL	$0, runtime·writelock(SB)

write_done:
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+16(FP)
	RET

TEXT runtime·nacl_exception_stack(SB),NOSPLIT,$0
	CALL	runtime·nacl_swapstack(SB)
	MOVL	0(NFP), DI  // p
	MOVL	4(NFP), SI  // size
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	nacl_exception_stack_irt
	NACL_SYSCALL(SYS_exception_stack)
	JMP	nacl_exception_stack_done
nacl_exception_stack_irt:
	MOVL	runtime·nacl_irt_exception_handling_v0_1+(IRT_EXCEPTION_STACK*4)(SB), AX
	CALL	AX
	NEGL	AX
nacl_exception_stack_done:
	CALL	runtime·nacl_restorestack(SB)
	MOVL	AX, ret+8(FP)
	RET

TEXT runtime·nacl_exception_handler(SB),NOSPLIT,$0
	CALL	runtime·nacl_swapstack(SB)
	MOVL	0(NFP), DI  // fn
	MOVL	4(NFP), SI  // arg
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	nacl_exception_handler_irt
	NACL_SYSCALL(SYS_exception_handler)
	JMP	nacl_exception_handler_done
nacl_exception_handler_irt:
	MOVL	runtime·nacl_irt_exception_handling_v0_1+(IRT_EXCEPTION_HANDLER*4)(SB), AX
	CALL	AX
	NEGL	AX
nacl_exception_handler_done:
	CALL	runtime·nacl_restorestack(SB)
	MOVL	AX, ret+8(FP)
	RET

TEXT runtime·nacl_sem_create(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	sem_create_irt
	MOVL	0(NFP), DI  // flag
	NACL_SYSCALL(SYS_sem_create)
	JMP	sem_create_done
sem_create_irt:
	SUBL	$4, SP
	MOVL	SP, DI	// *sem_handle
	MOVL	0(NFP), SI  // flag
	MOVL	runtime·nacl_irt_sem_v0_1+(IRT_SEM_CREATE*4)(SB), AX
	CALL	AX
	NEGL	AX
	JNZ	sem_create_irt_done
	MOVL	(SP), AX
sem_create_irt_done:
	ADDL	$4, SP
sem_create_done:
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+8(FP)
	RET

TEXT runtime·nacl_sem_wait(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	MOVL	0(NFP), DI  // sem
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	sem_wait_irt
	NACL_SYSCALL(SYS_sem_wait)
	JMP	sem_wait_done
sem_wait_irt:
	MOVL	runtime·nacl_irt_sem_v0_1+(IRT_SEM_WAIT*4)(SB), AX
	CALL	AX
	NEGL	AX
sem_wait_done:
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+8(FP)

TEXT runtime·nacl_sem_post(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	MOVL	0(NFP), DI  // sem
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	sem_post_irt
	NACL_SYSCALL(SYS_sem_post)
	JMP	sem_post_done
sem_post_irt:
	MOVL	runtime·nacl_irt_sem_v0_1+(IRT_SEM_POST*4)(SB), AX
	CALL	AX
	NEGL	AX
sem_post_done:
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+8(FP)
	RET

TEXT runtime·nacl_mutex_create(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	mutex_create_irt
	MOVL	0(NFP), DI  // flag
	NACL_SYSCALL(SYS_mutex_create)
	JMP	mutex_create_done
mutex_create_irt:
	SUBL	$4, SP
	MOVL	SP, DI	// *mutex_handle
	MOVL	0(NFP), SI  // flag
	MOVL	runtime·nacl_irt_mutex_v0_1+(IRT_MUTEX_CREATE*4)(SB), AX
	CALL	AX
	NEGL	AX
	JNZ	mutex_create_irt_done
	MOVL	4(SP), AX
mutex_create_irt_done:
	ADDL	$4, SP
mutex_create_done:
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+8(FP)
	RET

TEXT runtime·nacl_mutex_lock(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	MOVL	0(NFP), DI  // mutex
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	mutex_lock_irt
	NACL_SYSCALL(SYS_mutex_lock)
	JMP	mutex_lock_done
mutex_lock_irt:
	MOVL	runtime·nacl_irt_mutex_v0_1+(IRT_MUTEX_LOCK*4)(SB), AX
	CALL	AX
	NEGL	AX
mutex_lock_done:
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+8(FP)
	RET

TEXT runtime·nacl_mutex_trylock(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	MOVL	0(NFP), DI  // mutex
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	mutex_trylock_irt
	NACL_SYSCALL(SYS_mutex_trylock)
	JMP	mutex_trylock_done
mutex_trylock_irt:
	MOVL	runtime·nacl_irt_mutex_v0_1+(IRT_MUTEX_TRYLOCK*4)(SB), AX
	CALL	AX
	NEGL	AX
mutex_trylock_done:
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+8(FP)
	RET

TEXT runtime·nacl_mutex_unlock(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	MOVL	0(NFP), DI  // mutex
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	mutex_unlock_irt
	NACL_SYSCALL(SYS_mutex_unlock)
	JMP	mutex_unlock_done
mutex_unlock_irt:
	MOVL	runtime·nacl_irt_mutex_v0_1+(IRT_MUTEX_UNLOCK*4)(SB), AX
	CALL	AX
	NEGL	AX
mutex_unlock_done:
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+8(FP)
	RET

TEXT runtime·nacl_cond_create(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	cond_create_irt
	MOVL	0(NFP), DI  // flag
	NACL_SYSCALL(SYS_cond_create)
	JMP	cond_create_done
cond_create_irt:
	SUBL	$4, SP
	MOVL	SP, DI	// *cond_handle
	MOVL	0(NFP), SI  // flag
	MOVL	runtime·nacl_irt_cond_v0_1+(IRT_COND_CREATE*4)(SB), AX
	CALL	AX
	NEGL	AX
	JNZ	cond_create_irt_done
	MOVL	4(SP), AX
cond_create_irt_done:
	ADDL	$4, SP
cond_create_done:
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+8(FP)
	RET

TEXT runtime·nacl_cond_wait(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	MOVL	0(NFP), DI  // cond
	MOVL	4(NFP), SI  // n
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	cond_wait_irt
	NACL_SYSCALL(SYS_cond_wait)
	JMP	cond_wait_done
cond_wait_irt:
	MOVL	runtime·nacl_irt_cond_v0_1+(IRT_COND_WAIT*4)(SB), AX
	CALL	AX
	NEGL	AX
cond_wait_done:
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+8(FP)
	RET

TEXT runtime·nacl_cond_signal(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	MOVL	0(NFP), DI  // cond
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	cond_signal_irt
	NACL_SYSCALL(SYS_cond_signal)
	JMP	cond_signal_done
cond_signal_irt:
	MOVL	runtime·nacl_irt_cond_v0_1+(IRT_COND_SIGNAL*4)(SB), AX
	CALL	AX
	NEGL	AX
cond_signal_done:
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+8(FP)
	RET

TEXT runtime·nacl_cond_broadcast(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	MOVL	0(NFP), DI  // cond
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	cond_signal_irt
	NACL_SYSCALL(SYS_cond_broadcast)
	JMP	cond_broadcast_done
cond_broadcast_irt:
	MOVL	runtime·nacl_irt_cond_v0_1+(IRT_COND_BROADCAST*4)(SB), AX
	CALL	AX
	NEGL	AX
cond_broadcast_done:
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+8(FP)
	RET

TEXT runtime·nacl_cond_timed_wait_abs(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	MOVL	0(NFP), DI  // cond
	MOVL	4(NFP), SI  // lock
	MOVL	8(NFP), DX  // ts
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	cond_timed_wait_abs_irt
	NACL_SYSCALL(SYS_cond_timed_wait_abs)
	JMP	cond_timed_wait_abs_done
cond_timed_wait_abs_irt:
	MOVL	runtime·nacl_irt_cond_v0_1+(IRT_COND_TIMED_WAIT_ABS*4)(SB), AX
	CALL	AX
	NEGL	AX
cond_timed_wait_abs_done:
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+16(FP)
	RET

TEXT runtime·nacl_thread_create(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	MOVL	0(NFP), DI  // fn
	MOVL	4(NFP), SI  // stk
	MOVL	8(NFP), DX  // tls
	MOVL	12(NFP), CX  // xx
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	thread_create_irt
	NACL_SYSCALL(SYS_thread_create)
	JMP	thread_create_done
thread_create_irt:
	MOVL	runtime·nacl_irt_thread_v0_1+(IRT_THREAD_CREATE*4)(SB), AX
	CALL	AX
	NEGL	AX
thread_create_done:
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+16(FP)
	RET

TEXT runtime·mstart_nacl(SB),NOSPLIT,$0
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	mstart_irt
	NACL_SYSCALL(SYS_tls_get)
	JMP	mstart_done
mstart_irt:
	MOVL	runtime·nacl_irt_tls_v0_1+(IRT_TLS_GET*4)(SB), AX
	CALL	AX
mstart_done:
	SUBL	$8, AX
	MOVL	AX, TLS
	JMP	runtime·mstart(SB)

TEXT runtime·nacl_nanosleep(SB),NOSPLIT,$0
	CALL	runtime·nacl_entersyscall(SB)
	MOVL	0(NFP), DI  // ts
	MOVL	4(NFP), SI  // extra
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	nanosleep_irt
	NACL_SYSCALL(SYS_nanosleep)
	JMP	nanosleep_done
nanosleep_irt:
	MOVL	runtime·nacl_irt_basic_v0_1+(IRT_BASIC_NANOSLEEP*4)(SB), AX
	CALL	AX
	NEGL	AX
nanosleep_done:
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
	SUBL	$16, SP
	MOVL	0(NFP), DI  // addr
	MOVL	4(NFP), SI  // n
	MOVL	8(NFP), DX  // prot
	MOVL	12(NFP), CX  // flags
	MOVL	16(NFP), R8  // fd
	MOVL	20(NFP), R9  // off
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	mmap_irt
	MOVQ	R9, 0(SP)
	LEAL	SP, R9	// &off
	NACL_SYSCALL(SYS_mmap)
	CMPL	AX, $-4095
	JNA	mmap_done
	NEGL	AX
	JMP	mmap_done
mmap_irt:
	MOVQ	DI, 8(SP)  // &addr
	LEAQ	8(SP), DI
	MOVL	runtime·nacl_irt_memory_v0_3+(IRT_MEMORY_MMAP*4)(SB), AX
	CALL	AX
	NEGL	AX
	TESTL	AX, AX
	JNE	mmap_done
	MOVL	8(SP), AX
mmap_done:
	ADDL	$16, SP
	CALL	runtime·nacl_exitsyscall(SB)
	MOVL	AX, ret+24(FP)
	RET

TEXT time·now(SB),NOSPLIT,$0
	MOVQ runtime·timens(SB), AX
	CMPQ AX, $0
	JEQ realtime
	MOVQ $0, DX
	MOVQ $1000000000, CX
	DIVQ CX
	MOVQ AX, sec+0(FP)
	MOVL DX, nsec+8(FP)
	RET
realtime:
	CALL	runtime·nacl_swapstack(SB)
	SUBL	$16, SP
	MOVL	$0, DI // real time clock
	MOVL	SP, SI // timespec
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	now_irt
	NACL_SYSCALL(SYS_clock_gettime)
	JMP	now_done
now_irt:
	MOVL	runtime·nacl_irt_clock_v0_1+(IRT_CLOCK_GETTIME*4)(SB), AX
	CALL	AX
now_done:
	MOVL	0(SP), AX // low 32 sec
	MOVL	4(SP), CX // high 32 sec
	MOVL	8(SP), BX // nsec
	// sec is in AX, nsec in BX
	MOVL	AX, 0(NFP)
	MOVL	CX, 4(NFP)
	MOVL	BX, 8(NFP)
	ADDL	$16, SP
	CALL	runtime·nacl_restorestack(SB)
	RET

TEXT syscall·now(SB),NOSPLIT,$0
	JMP time·now(SB)

TEXT runtime·nacl_clock_gettime(SB),NOSPLIT,$0
	CALL	runtime·nacl_swapstack(SB)
	MOVL	0(NFP), DI  // clk_id
	MOVL	4(NFP), SI  // *tp
	CMPL	runtime·nacl_irt_is_enabled(SB), $0
	JNE	clock_gettime_irt
	NACL_SYSCALL(SYS_clock_gettime)
	JMP	clock_gettime_done
clock_gettime_irt:
	MOVL	runtime·nacl_irt_clock_v0_1+(IRT_CLOCK_GETTIME*4)(SB), AX
	CALL	AX
	NEGL	AX
clock_gettime_done:
	CALL	runtime·nacl_restorestack(SB)
	MOVL	AX, ret+8(FP)
	RET

TEXT runtime·nanotime(SB),NOSPLIT,$0
	MOVQ	runtime·timens(SB), AX
	CMPQ	AX, $0
	JEQ	3(PC)
	MOVQ	AX, ret+0(FP)
	RET
	CALL	runtime·nacl_swapstack(SB)
	SUBL	$16, SP
	MOVL	$0, DI // real time clock
	LEAL	0(SP), AX
	MOVL	AX, SI // &timespec
	NACL_SYSCALL(SYS_clock_gettime)
	JMP	nanotime_done
nanotime_irt:
	MOVL	runtime·nacl_irt_clock_v0_1+(IRT_CLOCK_GETTIME*4)(SB), AX
	CALL	AX
nanotime_done:
	MOVQ	0(SP), AX // sec
	MOVL	8(SP), DX // nsec
	ADDL	$16, SP

	// sec is in AX, nsec in DX
	// return nsec in AX
	IMULQ	$1000000000, AX
	ADDQ	DX, AX
	CALL	runtime·nacl_restorestack(SB)
	MOVQ	AX, ret+0(FP)
	RET

TEXT runtime·sigtramp(SB),NOSPLIT,$80
	// restore TLS register at time of execution,
	// in case it's been smashed.
	// the TLS register is really BP, but for consistency
	// with non-NaCl systems it is referred to here as TLS.
	// NOTE: Cannot use SYS_tls_get here (like we do in mstart_nacl),
	// because the main thread never calls tls_set.
	LEAL ctxt+0(FP), AX
	MOVL (16*4+5*8)(AX), AX
	MOVL	AX, TLS

	// check that g exists
	get_tls(CX)
	MOVL	g(CX), DI

	CMPL	DI, $0
	JEQ	nog

	// save g
	MOVL	DI, 20(SP)

	// g = m->gsignal
	MOVL	g_m(DI), BX
	MOVL	m_gsignal(BX), BX
	MOVL	BX, g(CX)

//JMP debughandler

	// copy arguments for sighandler
	MOVL	$11, 0(SP) // signal
	MOVL	$0, 4(SP) // siginfo
	LEAL	ctxt+0(FP), AX
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

	// Restore registers as best we can. Impossible to do perfectly.
	// See comment in sys_nacl_386.s for extended rationale.
	LEAL	ctxt+0(FP), SI
	ADDL	$64, SI
	MOVQ	0(SI), AX
	MOVQ	8(SI), CX
	MOVQ	16(SI), DX
	MOVQ	24(SI), BX
	MOVL	32(SI), SP	// MOVL for SP sandboxing
	// 40(SI) is saved BP aka TLS, already restored above
	// 48(SI) is saved SI, never to be seen again
	MOVQ	56(SI), DI
	MOVQ	64(SI), R8
	MOVQ	72(SI), R9
	MOVQ	80(SI), R10
	MOVQ	88(SI), R11
	MOVQ	96(SI), R12
	MOVQ	104(SI), R13
	MOVQ	112(SI), R14
	// 120(SI) is R15, which is owned by Native Client and must not be modified
	MOVQ	128(SI), SI // saved PC
	// 136(SI) is saved EFLAGS, never to be seen again
	JMP	SI

debughandler:
	// print basic information
	LEAL	ctxt+0(FP), DI
	MOVL	$runtime·sigtrampf(SB), AX
	MOVL	AX, 0(SP)
	MOVQ	(16*4+16*8)(DI), BX // rip
	MOVQ	BX, 8(SP)
	MOVQ	(16*4+0*8)(DI), BX // rax
	MOVQ	BX, 16(SP)
	MOVQ	(16*4+1*8)(DI), BX // rcx
	MOVQ	BX, 24(SP)
	MOVQ	(16*4+2*8)(DI), BX // rdx
	MOVQ	BX, 32(SP)
	MOVQ	(16*4+3*8)(DI), BX // rbx
	MOVQ	BX, 40(SP)
	MOVQ	(16*4+7*8)(DI), BX // rdi
	MOVQ	BX, 48(SP)
	MOVQ	(16*4+15*8)(DI), BX // r15
	MOVQ	BX, 56(SP)
	MOVQ	(16*4+4*8)(DI), BX // rsp
	MOVQ	0(BX), BX
	MOVQ	BX, 64(SP)
	CALL	runtime·printf(SB)

	LEAL	ctxt+0(FP), DI
	MOVQ	(16*4+16*8)(DI), BX // rip
	MOVL	BX, 0(SP)
	MOVQ	(16*4+4*8)(DI), BX // rsp
	MOVL	BX, 4(SP)
	MOVL	$0, 8(SP)	// lr
	get_tls(CX)
	MOVL	g(CX), BX
	MOVL	BX, 12(SP)	// gp
	CALL	runtime·traceback(SB)

notls:
	MOVL	0, AX
	RET

nog:
	MOVL	0, AX
	RET

// cannot do real signal handling yet, because gsignal has not been allocated.
MOVL $1, DI; NACL_SYSCALL(SYS_exit)

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
	MOVL	0(BX), DI	// name
	TESTL	DI, DI
	JE	irt_done
	MOVL	4(BX), SI	// funtab
	MOVL	8(BX), DX	// size
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
