// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// struct nacl_irt_basic {
//   void (*exit)(int status);
//   int (*gettod)(struct timeval *tv);
//   int (*clock)(clock_t *ticks);
//   int (*nanosleep)(const struct timespec *req, struct timespec *rem);
//   int (*sched_yield)(void);
//   int (*sysconf)(int name, int *value);
// };
#define IRT_BASIC_EXIT          0
#define IRT_BASIC_GETTOD        1
#define IRT_BASIC_CLOCK         2
#define IRT_BASIC_NANOSLEEP     3
#define IRT_BASIC_SCHED_YIELD   4
#define IRT_BASIC_SYSCONF       5
#define IRT_BASIC_SIZE          6

// #define NACL_IRT_MEMORY_v0_3    "nacl-irt-memory-0.3"
// struct nacl_irt_memory {
//   int (*mmap)(void **addr, size_t len, int prot, int flags, int fd, off_t off);
//   int (*munmap)(void *addr, size_t len);
//   int (*mprotect)(void *addr, size_t len, int prot);
// };
#define IRT_MEMORY_MMAP         0
#define IRT_MEMORY_MUNMAP       1
#define IRT_MEMORY_MPROTECT     2
#define IRT_MEMORY_SIZE         3

// #define NACL_IRT_THREAD_v0_1   "nacl-irt-thread-0.1"
// struct nacl_irt_thread {
//   int (*thread_create)(void (*start_func)(void), void *stack, void *thread_ptr);
//   void (*thread_exit)(int32_t *stack_flag);
//   int (*thread_nice)(const int nice);
// };
#define IRT_THREAD_CREATE       0
#define IRT_THREAD_EXIT         1
#define IRT_THREAD_NICE         2
#define IRT_THREAD_SIZE         3

// #define NACL_IRT_FUTEX_v0_1        "nacl-irt-futex-0.1"
// struct nacl_irt_futex {
//   int (*futex_wait_abs)(volatile int *addr, int value,
//                         const struct timespec *abstime);
//   int (*futex_wake)(volatile int *addr, int nwake, int *count);
// };
#define IRT_FUTEX_WAIT          0
#define IRT_FUTEX_WAKE          1
#define IRT_FUTEX_SIZE          2

// #define NACL_IRT_FDIO_v0_1      "nacl-irt-fdio-0.1"
// #define NACL_IRT_DEV_FDIO_v0_1  "nacl-irt-dev-fdio-0.1"
// struct nacl_irt_fdio {
//   int (*close)(int fd);
//   int (*dup)(int fd, int *newfd);
//   int (*dup2)(int fd, int newfd);
//   int (*read)(int fd, void *buf, size_t count, size_t *nread);
//   int (*write)(int fd, const void *buf, size_t count, size_t *nwrote);
//   int (*seek)(int fd, off_t offset, int whence, off_t *new_offset);
//   int (*fstat)(int fd, struct stat *);
//   int (*getdents)(int fd, struct dirent *, size_t count, size_t *nread);
// };
#define IRT_FDIO_CLOSE          0
#define IRT_FDIO_DUP            1
#define IRT_FDIO_DUP2           2
#define IRT_FDIO_READ           3
#define IRT_FDIO_WRITE          4
#define IRT_FDIO_SEEK           5
#define IRT_FDIO_FSTAT          6
#define IRT_FDIO_GETDENTS       7
#define IRT_FDIO_SIZE           8

// #define NACL_IRT_FILENAME_v0_1      "nacl-irt-filename-0.1"
// struct nacl_irt_filename {
//   int (*open)(const char *pathname, int oflag, mode_t cmode, int *newfd);
//   int (*stat)(const char *pathname, struct stat *);
// };
#define IRT_FILENAME_OPEN       0
#define IRT_FILENAME_STAT       1
#define IRT_FILENAME_SIZE       2

// #define NACL_IRT_EXCEPTION_HANDLING_v0_1 \
//   "nacl-irt-exception-handling-0.1"
// typedef void (*NaClExceptionHandler)(struct NaClExceptionContext *context);
// struct nacl_irt_exception_handling {
//   int (*exception_handler)(NaClExceptionHandler handler,
//                            NaClExceptionHandler *old_handler);
//   int (*exception_stack)(void *stack, size_t size);
//   int (*exception_clear_flag)(void);
// };
#define IRT_EXCEPTION_HANDLER   0
#define IRT_EXCEPTION_STACK     1
#define IRT_EXCEPTION_CLEAR     2
#define IRT_EXCEPTION_SIZE      3

// #define NACL_IRT_MUTEX_v0_1        "nacl-irt-mutex-0.1"
// struct nacl_irt_mutex {
//   int (*mutex_create)(int *mutex_handle);
//   int (*mutex_destroy)(int mutex_handle);
//   int (*mutex_lock)(int mutex_handle);
//   int (*mutex_unlock)(int mutex_handle);
//   int (*mutex_trylock)(int mutex_handle);
// };
#define IRT_MUTEX_CREATE        0
#define IRT_MUTEX_DESTROY       1
#define IRT_MUTEX_LOCK          2
#define IRT_MUTEX_UNLOCK        3
#define IRT_MUTEX_TRYLOCK       4
#define IRT_MUTEX_SIZE          5

// #define NACL_IRT_COND_v0_1      "nacl-irt-cond-0.1"
// struct nacl_irt_cond {
//   int (*cond_create)(int *cond_handle);
//   int (*cond_destroy)(int cond_handle);
//   int (*cond_signal)(int cond_handle);
//   int (*cond_broadcast)(int cond_handle);
//   int (*cond_wait)(int cond_handle, int mutex_handle);
//   int (*cond_timed_wait_abs)(int cond_handle, int mutex_handle,
//                              const struct timespec *abstime);
// };
#define IRT_COND_CREATE         0
#define IRT_CONT_DESTROY        1
#define IRT_COND_SIGNAL         2
#define IRT_COND_BROADCAST      3
#define IRT_COND_WAIT           4
#define IRT_COND_TIMED_WAIT_ABS 5
#define IRT_COND_SIZE           6

// #define NACL_IRT_SEM_v0_1       "nacl-irt-sem-0.1"
// struct nacl_irt_sem {
//   int (*sem_create)(int *sem_handle, int32_t value);
//   int (*sem_destroy)(int sem_handle);
//   int (*sem_post)(int sem_handle);
//   int (*sem_wait)(int sem_handle);
// };
#define IRT_SEM_CREATE          0
#define IRT_SEM_DESTROY         1
#define IRT_SEM_POST            2
#define IRT_SEM_WAIT            3
#define IRT_SEM_SIZE            4

// #define NACL_IRT_TLS_v0_1       "nacl-irt-tls-0.1"
// struct nacl_irt_tls {
//   int (*tls_init)(void *thread_ptr);
//   void *(*tls_get)(void);
// };
#define IRT_TLS_INIT            0
#define IRT_TLS_GET             1
#define IRT_TLS_SIZE            2

// #define NACL_IRT_RANDOM_v0_1 "nacl-irt-random-0.1"
// struct nacl_irt_random {
//   int (*get_random_bytes)(void *buf, size_t count, size_t *nread);
// };
#define IRT_RANDOM_BYTES        0
#define IRT_RANDOM_SIZE         1

// #define NACL_IRT_CLOCK_v0_1 "nacl-irt-clock_get-0.1"
// struct nacl_irt_clock {
//   int (*clock_getres)(nacl_irt_clockid_t clock_id, struct timespec *res);
//   int (*clock_gettime)(nacl_irt_clockid_t clock_id, struct timespec *tp);
// };
#define IRT_CLOCK_GETRES        0
#define IRT_CLOCK_GETTIME       1
#define IRT_CLOCK_SIZE          2

// #define NACL_IRT_PPAPIHOOK_v0_1 "nacl-irt-ppapihook-0.1"
// struct nacl_irt_ppapihook {
//   int (*ppapi_start)(const struct PP_StartFunctions *);
//   void (*ppapi_register_thread_creator)(const struct PP_ThreadFunctions *);
// };
#define IRT_PPAPI_START         0
#define IRT_PPAPI_REGISTER_THREAD_CREATOR       1
#define IRT_PPAPIHOOK_SIZE      2
