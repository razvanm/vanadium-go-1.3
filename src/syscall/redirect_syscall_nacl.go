package syscall

import (
	"runtime"
)

// runtime_ppapi_InitPPAPI calls into a method defined in the runtime/ppapi package
// to initialize PPAPI. It is not fully initialized until a callback is received (triggering
// a call to SetImplementation).
func runtime_ppapi_InitPPAPI()

func init() {
	waitToBeSet = make(chan bool, 1)
	go func() {
		runtime.LockOSThread() // This is the main PPAPI thread and must stay that way.
		runtime_ppapi_InitPPAPI()
	}()
	<-waitToBeSet
}

var waitToBeSet chan bool
var currentImplementation SyscallImplementation

// SetImplementation sets the current system call implementation.
func SetImplementation(impl SyscallImplementation) {
	currentImplementation = impl
	waitToBeSet <- true
}

func Accept(fd int) (nfd int, sa Sockaddr, err error) {
	return currentImplementation.Accept(fd)
}
func Accept4(fd int, flags int) {
	currentImplementation.Accept4(fd, flags)
}
func Access(path string, mode uint32) (err error) {
	return currentImplementation.Access(path, mode)
}
func Acct(path string) (err error) {
	return currentImplementation.Acct(path)
}
func Adjtimex(buf *Timex) (state int, err error) {
	return currentImplementation.Adjtimex(buf)
}
func AttachLsf(fd int, i []SockFilter) error {
	return currentImplementation.AttachLsf(fd, i)
}
func Bind(fd int, sa Sockaddr) (err error) {
	return currentImplementation.Bind(fd, sa)
}
func BindToDevice(fd int, device string) (err error) {
	return currentImplementation.BindToDevice(fd, device)
}
func Chdir(path string) (err error) {
	return currentImplementation.Chdir(path)
}
func Chmod(path string, mode uint32) (err error) {
	return currentImplementation.Chmod(path, mode)
}
func Chown(path string, uid int, gid int) (err error) {
	return currentImplementation.Chown(path, uid, gid)
}
func Chroot(path string) (err error) {
	return currentImplementation.Chroot(path)
}
func Clearenv() {
	currentImplementation.Clearenv()
}
func Close(fd int) (err error) {
	return currentImplementation.Close(fd)
}
func CloseOnExec(fd int) {
	currentImplementation.CloseOnExec(fd)
}
func CmsgLen(datalen int) int {
	return currentImplementation.CmsgLen(datalen)
}
func CmsgSpace(datalen int) int {
	return currentImplementation.CmsgSpace(datalen)
}
func Connect(fd int, sa Sockaddr) (err error) {
	return currentImplementation.Connect(fd, sa)
}
func Creat(path string, mode uint32) (fd int, err error) {
	return currentImplementation.Creat(path, mode)
}
func DetachLsf(fd int) error {
	return currentImplementation.DetachLsf(fd)
}
func Dup(oldfd int) (fd int, err error) {
	return currentImplementation.Dup(oldfd)
}
func Dup2(oldfd int, newfd int) {
	currentImplementation.Dup2(oldfd, newfd)
}
func Dup3(oldfd int, newfd int, flags int) {
	currentImplementation.Dup3(oldfd, newfd, flags)
}
func Environ() []string {
	return currentImplementation.Environ()
}
func EpollCreate(size int) (fd int, err error) {
	return currentImplementation.EpollCreate(size)
}
func EpollCreate1(flag int) {
	currentImplementation.EpollCreate1(flag)
}
func EpollCtl(epfd int, op int, fd int, event *EpollEvent) (err error) {
	return currentImplementation.EpollCtl(epfd, op, fd, event)
}
func EpollWait(epfd int, events []EpollEvent, msec int) (n int, err error) {
	return currentImplementation.EpollWait(epfd, events, msec)
}
func Exec(argv0 string, argv []string, envv []string) (err error) {
	return currentImplementation.Exec(argv0, argv, envv)
}
func Exit(code int) {
	currentImplementation.Exit(code)
}
func Faccessat(dirfd int, path string, mode uint32, flags int) (err error) {
	return currentImplementation.Faccessat(dirfd, path, mode, flags)
}
func Fallocate(fd int, mode uint32, off int64, len int64) (err error) {
	return currentImplementation.Fallocate(fd, mode, off, len)
}
func Fchdir(fd int) (err error) {
	return currentImplementation.Fchdir(fd)
}
func Fchmod(fd int, mode uint32) (err error) {
	return currentImplementation.Fchmod(fd, mode)
}
func Fchmodat(dirfd int, path string, mode uint32, flags int) (err error) {
	return currentImplementation.Fchmodat(dirfd, path, mode, flags)
}
func Fchown(fd int, uid int, gid int) (err error) {
	return currentImplementation.Fchown(fd, uid, gid)
}
func Fchownat(dirfd int, path string, uid int, gid int, flags int) (err error) {
	return currentImplementation.Fchownat(dirfd, path, uid, gid, flags)
}
func FcntlFlock(fd uintptr, cmd int, lk *Flock_t) error {
	return currentImplementation.FcntlFlock(fd, cmd, lk)
}
func Fdatasync(fd int) (err error) {
	return currentImplementation.Fdatasync(fd)
}
func Flock(fd int, how int) (err error) {
	return currentImplementation.Flock(fd, how)
}
func ForkExec(argv0 string, argv []string, attr *ProcAttr) (pid int, err error) {
	return currentImplementation.ForkExec(argv0, argv, attr)
}
func Fstat(fd int, stat *Stat_t) (err error) {
	return currentImplementation.Fstat(fd, stat)
}
func Fstatfs(fd int, buf *Statfs_t) (err error) {
	return currentImplementation.Fstatfs(fd, buf)
}
func Fsync(fd int) (err error) {
	return currentImplementation.Fsync(fd)
}
func Ftruncate(fd int, length int64) (err error) {
	return currentImplementation.Ftruncate(fd, length)
}
func Futimes(fd int, tv []Timeval) (err error) {
	return currentImplementation.Futimes(fd, tv)
}
func Futimesat(dirfd int, path string, tv []Timeval) (err error) {
	return currentImplementation.Futimesat(dirfd, path, tv)
}
func Getcwd(buf []byte) (n int, err error) {
	return currentImplementation.Getcwd(buf)
}
func Getdents(fd int, buf []byte) (n int, err error) {
	return currentImplementation.Getdents(fd, buf)
}
func Getegid() (egid int) {
	return currentImplementation.Getegid()
}
func Getenv(key string) (value string, found bool) {
	return currentImplementation.Getenv(key)
}
func Geteuid() (euid int) {
	return currentImplementation.Geteuid()
}
func Getgid() (gid int) {
	return currentImplementation.Getgid()
}
func Getgroups() (gids []int, err error) {
	return currentImplementation.Getgroups()
}
func Getpagesize() int {
	return currentImplementation.Getpagesize()
}
func Getpgid(pid int) (pgid int, err error) {
	return currentImplementation.Getpgid(pid)
}
func Getpgrp() (pid int) {
	return currentImplementation.Getpgrp()
}
func Getpid() (pid int) {
	return currentImplementation.Getpid()
}
func Getppid() (ppid int) {
	return currentImplementation.Getppid()
}
func Getpriority(which int, who int) (prio int, err error) {
	return currentImplementation.Getpriority(which, who)
}
func Getrlimit(resource int, rlim *Rlimit) (err error) {
	return currentImplementation.Getrlimit(resource, rlim)
}
func Getrusage(who int, rusage *Rusage) (err error) {
	return currentImplementation.Getrusage(who, rusage)
}
func GetsockoptInet4Addr(fd, level, opt int) {
	currentImplementation.GetsockoptInet4Addr(fd, level, opt)
}
func GetsockoptInt(fd, level, opt int) (value int, err error) {
	return currentImplementation.GetsockoptInt(fd, level, opt)
}
func Gettid() (tid int) {
	return currentImplementation.Gettid()
}
func Gettimeofday(tv *Timeval) (err error) {
	return currentImplementation.Gettimeofday(tv)
}
func Getuid() (uid int) {
	return currentImplementation.Getuid()
}
func Getwd() (wd string, err error) {
	return currentImplementation.Getwd()
}
func Getxattr(path string, attr string, dest []byte) (sz int, err error) {
	return currentImplementation.Getxattr(path, attr, dest)
}
func InotifyAddWatch(fd int, pathname string, mask uint32) (watchdesc int, err error) {
	return currentImplementation.InotifyAddWatch(fd, pathname, mask)
}
func InotifyInit() (fd int, err error) {
	return currentImplementation.InotifyInit()
}
func InotifyInit1(flags int) {
	currentImplementation.InotifyInit1(flags)
}
func InotifyRmWatch(fd int, watchdesc uint32) (success int, err error) {
	return currentImplementation.InotifyRmWatch(fd, watchdesc)
}
func Ioperm(from int, num int, on int) (err error) {
	return currentImplementation.Ioperm(from, num, on)
}
func Iopl(level int) (err error) {
	return currentImplementation.Iopl(level)
}
func Kill(pid int, sig Signal) (err error) {
	return currentImplementation.Kill(pid, sig)
}
func Klogctl(typ int, buf []byte) (n int, err error) {
	return currentImplementation.Klogctl(typ, buf)
}
func Lchown(path string, uid int, gid int) (err error) {
	return currentImplementation.Lchown(path, uid, gid)
}
func Link(oldpath string, newpath string) (err error) {
	return currentImplementation.Link(oldpath, newpath)
}
func Listen(s int, n int) (err error) {
	return currentImplementation.Listen(s, n)
}
func Listxattr(path string, dest []byte) (sz int, err error) {
	return currentImplementation.Listxattr(path, dest)
}
func LsfSocket(ifindex, proto int) (int, error) {
	return currentImplementation.LsfSocket(ifindex, proto)
}
func Lstat(path string, stat *Stat_t) (err error) {
	return currentImplementation.Lstat(path, stat)
}
func Madvise(b []byte, advice int) (err error) {
	return currentImplementation.Madvise(b, advice)
}
func Mkdir(path string, mode uint32) (err error) {
	return currentImplementation.Mkdir(path, mode)
}
func Mkdirat(dirfd int, path string, mode uint32) (err error) {
	return currentImplementation.Mkdirat(dirfd, path, mode)
}
func Mkfifo(path string, mode uint32) (err error) {
	return currentImplementation.Mkfifo(path, mode)
}
func Mknod(path string, mode uint32, dev int) (err error) {
	return currentImplementation.Mknod(path, mode, dev)
}
func Mknodat(dirfd int, path string, mode uint32, dev int) (err error) {
	return currentImplementation.Mknodat(dirfd, path, mode, dev)
}
func Mlock(b []byte) (err error) {
	return currentImplementation.Mlock(b)
}
func Mlockall(flags int) (err error) {
	return currentImplementation.Mlockall(flags)
}
func Mmap(fd int, offset int64, length int, prot int, flags int) (data []byte, err error) {
	return currentImplementation.Mmap(fd, offset, length, prot, flags)
}
func Mount(source string, target string, fstype string, flags uintptr, data string) (err error) {
	return currentImplementation.Mount(source, target, fstype, flags, data)
}
func Mprotect(b []byte, prot int) (err error) {
	return currentImplementation.Mprotect(b, prot)
}
func Munlock(b []byte) (err error) {
	return currentImplementation.Munlock(b)
}
func Munlockall() (err error) {
	return currentImplementation.Munlockall()
}
func Munmap(b []byte) (err error) {
	return currentImplementation.Munmap(b)
}
func Nanosleep(time *Timespec, leftover *Timespec) (err error) {
	return currentImplementation.Nanosleep(time, leftover)
}
func NetlinkRIB(proto, family int) ([]byte, error) {
	return currentImplementation.NetlinkRIB(proto, family)
}
func Open(path string, mode int, perm uint32) (fd int, err error) {
	return currentImplementation.Open(path, mode, perm)
}
func Openat(dirfd int, path string, flags int, mode uint32) (fd int, err error) {
	return currentImplementation.Openat(dirfd, path, flags, mode)
}
func ParseDirent(buf []byte, max int, names []string) (consumed int, count int, newnames []string) {
	return currentImplementation.ParseDirent(buf, max, names)
}
func ParseNetlinkMessage(b []byte) ([]NetlinkMessage, error) {
	return currentImplementation.ParseNetlinkMessage(b)
}
func ParseNetlinkRouteAttr(m *NetlinkMessage) ([]NetlinkRouteAttr, error) {
	return currentImplementation.ParseNetlinkRouteAttr(m)
}
func ParseSocketControlMessage(b []byte) ([]SocketControlMessage, error) {
	return currentImplementation.ParseSocketControlMessage(b)
}
func ParseUnixRights(m *SocketControlMessage) ([]int, error) {
	return currentImplementation.ParseUnixRights(m)
}
func Pause() (err error) {
	return currentImplementation.Pause()
}
func Pipe(p []int) (err error) {
	return currentImplementation.Pipe(p)
}
func Pipe2(p []int, flags int) {
	currentImplementation.Pipe2(p, flags)
}
func PivotRoot(newroot string, putold string) (err error) {
	return currentImplementation.PivotRoot(newroot, putold)
}
func Pread(fd int, p []byte, offset int64) (n int, err error) {
	return currentImplementation.Pread(fd, p, offset)
}
func PtraceAttach(pid int) (err error) {
	return currentImplementation.PtraceAttach(pid)
}
func PtraceCont(pid int, signal int) (err error) {
	return currentImplementation.PtraceCont(pid, signal)
}
func PtraceDetach(pid int) (err error) {
	return currentImplementation.PtraceDetach(pid)
}
func PtraceGetEventMsg(pid int) (msg uint, err error) {
	return currentImplementation.PtraceGetEventMsg(pid)
}
func PtraceGetRegs(pid int, regsout *PtraceRegs) (err error) {
	return currentImplementation.PtraceGetRegs(pid, regsout)
}
func PtracePeekData(pid int, addr uintptr, out []byte) (count int, err error) {
	return currentImplementation.PtracePeekData(pid, addr, out)
}
func PtracePeekText(pid int, addr uintptr, out []byte) (count int, err error) {
	return currentImplementation.PtracePeekText(pid, addr, out)
}
func PtracePokeData(pid int, addr uintptr, data []byte) (count int, err error) {
	return currentImplementation.PtracePokeData(pid, addr, data)
}
func PtracePokeText(pid int, addr uintptr, data []byte) (count int, err error) {
	return currentImplementation.PtracePokeText(pid, addr, data)
}
func PtraceSetOptions(pid int, options int) (err error) {
	return currentImplementation.PtraceSetOptions(pid, options)
}
func PtraceSetRegs(pid int, regs *PtraceRegs) (err error) {
	return currentImplementation.PtraceSetRegs(pid, regs)
}
func PtraceSingleStep(pid int) (err error) {
	return currentImplementation.PtraceSingleStep(pid)
}
func PtraceSyscall(pid int, signal int) (err error) {
	return currentImplementation.PtraceSyscall(pid, signal)
}
func Pwrite(fd int, p []byte, offset int64) (n int, err error) {
	return currentImplementation.Pwrite(fd, p, offset)
}
func RawSyscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err Errno) {
	return currentImplementation.RawSyscall(trap, a1, a2, a3)
}
func RawSyscall6(trap, a1, a2, a3, a4, a5, a6 uintptr) {
	currentImplementation.RawSyscall6(trap, a1, a2, a3, a4, a5, a6)
}
func Read(fd int, p []byte) (n int, err error) {
	return currentImplementation.Read(fd, p)
}
func ReadDirent(fd int, buf []byte) (n int, err error) {
	return currentImplementation.ReadDirent(fd, buf)
}
func Readlink(path string, buf []byte) (n int, err error) {
	return currentImplementation.Readlink(path, buf)
}
func Reboot(cmd int) (err error) {
	return currentImplementation.Reboot(cmd)
}
func Recvfrom(fd int, p []byte, flags int) (n int, from Sockaddr, err error) {
	return currentImplementation.Recvfrom(fd, p, flags)
}
func Recvmsg(fd int, p, oob []byte, flags int) (n, oobn int, recvflags int, from Sockaddr, err error) {
	return currentImplementation.Recvmsg(fd, p, oob, flags)
}
func Removexattr(path string, attr string) (err error) {
	return currentImplementation.Removexattr(path, attr)
}
func Rename(oldpath string, newpath string) (err error) {
	return currentImplementation.Rename(oldpath, newpath)
}
func Renameat(olddirfd int, oldpath string, newdirfd int, newpath string) (err error) {
	return currentImplementation.Renameat(olddirfd, oldpath, newdirfd, newpath)
}
func Rmdir(path string) (err error) {
	return currentImplementation.Rmdir(path)
}
func Seek(fd int, offset int64, whence int) (off int64, err error) {
	return currentImplementation.Seek(fd, offset, whence)
}
func Select(nfd int, r *FdSet, w *FdSet, e *FdSet, timeout *Timeval) (n int, err error) {
	return currentImplementation.Select(nfd, r, w, e, timeout)
}
func Sendfile(outfd int, infd int, offset *int64, count int) (written int, err error) {
	return currentImplementation.Sendfile(outfd, infd, offset, count)
}
func Sendmsg(fd int, p, oob []byte, to Sockaddr, flags int) (err error) {
	return currentImplementation.Sendmsg(fd, p, oob, to, flags)
}
func SendmsgN(fd int, p, oob []byte, to Sockaddr, flags int) (n int, err error) {
	return currentImplementation.SendmsgN(fd, p, oob, to, flags)
}
func Sendto(fd int, p []byte, flags int, to Sockaddr) (err error) {
	return currentImplementation.Sendto(fd, p, flags, to)
}
func SetLsfPromisc(name string, m bool) error {
	return currentImplementation.SetLsfPromisc(name, m)
}
func SetNonblock(fd int, nonblocking bool) (err error) {
	return currentImplementation.SetNonblock(fd, nonblocking)
}
func Setdomainname(p []byte) (err error) {
	return currentImplementation.Setdomainname(p)
}
func Setenv(key, value string) error {
	return currentImplementation.Setenv(key, value)
}
func Setfsgid(gid int) (err error) {
	return currentImplementation.Setfsgid(gid)
}
func Setfsuid(uid int) (err error) {
	return currentImplementation.Setfsuid(uid)
}
func Setgid(gid int) (err error) {
	return currentImplementation.Setgid(gid)
}
func Setgroups(gids []int) (err error) {
	return currentImplementation.Setgroups(gids)
}
func Sethostname(p []byte) (err error) {
	return currentImplementation.Sethostname(p)
}
func Setpgid(pid int, pgid int) (err error) {
	return currentImplementation.Setpgid(pid, pgid)
}
func Setpriority(which int, who int, prio int) (err error) {
	return currentImplementation.Setpriority(which, who, prio)
}
func Setregid(rgid int, egid int) (err error) {
	return currentImplementation.Setregid(rgid, egid)
}
func Setresgid(rgid int, egid int, sgid int) (err error) {
	return currentImplementation.Setresgid(rgid, egid, sgid)
}
func Setresuid(ruid int, euid int, suid int) (err error) {
	return currentImplementation.Setresuid(ruid, euid, suid)
}
func Setreuid(ruid int, euid int) (err error) {
	return currentImplementation.Setreuid(ruid, euid)
}
func Setrlimit(resource int, rlim *Rlimit) (err error) {
	return currentImplementation.Setrlimit(resource, rlim)
}
func Setsid() (pid int, err error) {
	return currentImplementation.Setsid()
}
func SetsockoptByte(fd, level, opt int, value byte) (err error) {
	return currentImplementation.SetsockoptByte(fd, level, opt, value)
}
func SetsockoptICMPv6Filter(fd, level, opt int, filter *ICMPv6Filter) error {
	return currentImplementation.SetsockoptICMPv6Filter(fd, level, opt, filter)
}
func SetsockoptIPMreq(fd, level, opt int, mreq *IPMreq) (err error) {
	return currentImplementation.SetsockoptIPMreq(fd, level, opt, mreq)
}
func SetsockoptIPMreqn(fd, level, opt int, mreq *IPMreqn) (err error) {
	return currentImplementation.SetsockoptIPMreqn(fd, level, opt, mreq)
}
func SetsockoptIPv6Mreq(fd, level, opt int, mreq *IPv6Mreq) (err error) {
	return currentImplementation.SetsockoptIPv6Mreq(fd, level, opt, mreq)
}
func SetsockoptInet4Addr(fd, level, opt int, value [4]byte) (err error) {
	return currentImplementation.SetsockoptInet4Addr(fd, level, opt, value)
}
func SetsockoptInt(fd, level, opt int, value int) (err error) {
	return currentImplementation.SetsockoptInt(fd, level, opt, value)
}
func SetsockoptLinger(fd, level, opt int, l *Linger) (err error) {
	return currentImplementation.SetsockoptLinger(fd, level, opt, l)
}
func SetsockoptString(fd, level, opt int, s string) (err error) {
	return currentImplementation.SetsockoptString(fd, level, opt, s)
}
func SetsockoptTimeval(fd, level, opt int, tv *Timeval) (err error) {
	return currentImplementation.SetsockoptTimeval(fd, level, opt, tv)
}
func Settimeofday(tv *Timeval) (err error) {
	return currentImplementation.Settimeofday(tv)
}
func Setuid(uid int) (err error) {
	return currentImplementation.Setuid(uid)
}
func Setxattr(path string, attr string, data []byte, flags int) (err error) {
	return currentImplementation.Setxattr(path, attr, data, flags)
}
func Shutdown(fd int, how int) (err error) {
	return currentImplementation.Shutdown(fd, how)
}
func SlicePtrFromStrings(ss []string) ([]*byte, error) {
	return currentImplementation.SlicePtrFromStrings(ss)
}
func Socket(domain, typ, proto int) (fd int, err error) {
	return currentImplementation.Socket(domain, typ, proto)
}
func Socketpair(domain, typ, proto int) (fd [2]int, err error) {
	return currentImplementation.Socketpair(domain, typ, proto)
}
func Splice(rfd int, roff *int64, wfd int, woff *int64, len int, flags int) (n int64, err error) {
	return currentImplementation.Splice(rfd, roff, wfd, woff, len, flags)
}
func StartProcess(argv0 string, argv []string, attr *ProcAttr) (pid int, handle uintptr, err error) {
	return currentImplementation.StartProcess(argv0, argv, attr)
}
func Stat(path string, stat *Stat_t) (err error) {
	return currentImplementation.Stat(path, stat)
}
func Statfs(path string, buf *Statfs_t) (err error) {
	return currentImplementation.Statfs(path, buf)
}
func StringSlicePtr(ss []string) []*byte {
	return currentImplementation.StringSlicePtr(ss)
}
func Symlink(oldpath string, newpath string) (err error) {
	return currentImplementation.Symlink(oldpath, newpath)
}
func Sync() {
	currentImplementation.Sync()
}
func SyncFileRange(fd int, off int64, n int64, flags int) (err error) {
	return currentImplementation.SyncFileRange(fd, off, n, flags)
}
func Syscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err Errno) {
	return currentImplementation.Syscall(trap, a1, a2, a3)
}
func Syscall6(trap, a1, a2, a3, a4, a5, a6 uintptr) {
	currentImplementation.Syscall6(trap, a1, a2, a3, a4, a5, a6)
}
func Sysinfo(info *Sysinfo_t) (err error) {
	return currentImplementation.Sysinfo(info)
}
func Tee(rfd int, wfd int, len int, flags int) (n int64, err error) {
	return currentImplementation.Tee(rfd, wfd, len, flags)
}
func Tgkill(tgid int, tid int, sig Signal) (err error) {
	return currentImplementation.Tgkill(tgid, tid, sig)
}
func Times(tms *Tms) (ticks uintptr, err error) {
	return currentImplementation.Times(tms)
}
func TimespecToNsec(ts Timespec) int64 {
	return currentImplementation.TimespecToNsec(ts)
}
func TimevalToNsec(tv Timeval) int64 {
	return currentImplementation.TimevalToNsec(tv)
}
func Truncate(path string, length int64) (err error) {
	return currentImplementation.Truncate(path, length)
}
func Umask(mask int) (oldmask int) {
	return currentImplementation.Umask(mask)
}
func Uname(buf *Utsname) (err error) {
	return currentImplementation.Uname(buf)
}
func UnixCredentials(ucred *Ucred) []byte {
	return currentImplementation.UnixCredentials(ucred)
}
func UnixRights(fds ...int) []byte {
	return currentImplementation.UnixRights(fds...)
}
func Unlink(path string) (err error) {
	return currentImplementation.Unlink(path)
}
func Unlinkat(dirfd int, path string) (err error) {
	return currentImplementation.Unlinkat(dirfd, path)
}
func Unmount(target string, flags int) (err error) {
	return currentImplementation.Unmount(target, flags)
}
func Unsetenv(key string) error {
	return currentImplementation.Unsetenv(key)
}
func Unshare(flags int) (err error) {
	return currentImplementation.Unshare(flags)
}
func Ustat(dev int, ubuf *Ustat_t) (err error) {
	return currentImplementation.Ustat(dev, ubuf)
}
func Utime(path string, buf *Utimbuf) (err error) {
	return currentImplementation.Utime(path, buf)
}
func Utimes(path string, tv []Timeval) (err error) {
	return currentImplementation.Utimes(path, tv)
}
func UtimesNano(path string, ts []Timespec) (err error) {
	return currentImplementation.UtimesNano(path, ts)
}
func Wait4(pid int, wstatus *WaitStatus, options int, rusage *Rusage) (wpid int, err error) {
	return currentImplementation.Wait4(pid, wstatus, options, rusage)
}
func Write(fd int, p []byte) (n int, err error) {
	return currentImplementation.Write(fd, p)
}

func GetsockoptICMPv6Filter(fd, level, opt int) (*ICMPv6Filter, error) {
	return currentImplementation.GetsockoptICMPv6Filter(fd, level, opt)
}
func GetsockoptIPMreq(fd, level, opt int) (*IPMreq, error) {
	return currentImplementation.GetsockoptIPMreq(fd, level, opt)
}
func GetsockoptIPMreqn(fd, level, opt int) (*IPMreqn, error) {
	return currentImplementation.GetsockoptIPMreqn(fd, level, opt)
}
func GetsockoptIPv6MTUInfo(fd, level, opt int) (*IPv6MTUInfo, error) {
	return currentImplementation.GetsockoptIPv6MTUInfo(fd, level, opt)
}
func GetsockoptIPv6Mreq(fd, level, opt int) (*IPv6Mreq, error) {
	return currentImplementation.GetsockoptIPv6Mreq(fd, level, opt)
}
func LsfJump(code, k, jt, jf int) *SockFilter {
	return currentImplementation.LsfJump(code, k, jt, jf)
}
func LsfStmt(code, k int) *SockFilter {
	return currentImplementation.LsfStmt(code, k)
}
func Getpeername(fd int) (sa Sockaddr, err error) {
	return currentImplementation.Getpeername(fd)
}
func Getsockname(fd int) (sa Sockaddr, err error) {
	return currentImplementation.Getsockname(fd)
}
func Time(t *Time_t) (tt Time_t, err error) {
	return currentImplementation.Time(t)
}
func NsecToTimespec(nsec int64) (ts Timespec) {
	return currentImplementation.NsecToTimespec(nsec)
}
func NsecToTimeval(nsec int64) (tv Timeval) {
	return currentImplementation.NsecToTimeval(nsec)
}
func GetsockoptUcred(fd, level, opt int) (*Ucred, error) {
	return currentImplementation.GetsockoptUcred(fd, level, opt)
}
func ParseUnixCredentials(m *SocketControlMessage) (*Ucred, error) {
	return currentImplementation.ParseUnixCredentials(m)
}

func SetReadDeadline(fd int, t int64) error {
	return currentImplementation.SetReadDeadline(fd, t)
}

func SetWriteDeadline(fd int, t int64) error {
	return currentImplementation.SetWriteDeadline(fd, t)
}

func StopIO(fd int) error {
	return currentImplementation.StopIO(fd)
}
