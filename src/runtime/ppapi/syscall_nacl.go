package ppapi

import (
	"fmt"
	"strings"
	"syscall"
)

type PPAPISyscallImpl struct {
	Instance
}

var envVars map[string]string

func initEnvVars() {
	if envVars == nil {
		envVars = map[string]string{
			"TMPDIR":          "/tmp",
		}
	}
}

func (PPAPISyscallImpl) Accept4(fd int, flags int) (nfd int, sa syscall.Sockaddr, err error) {
	panic("Accept() not implemented")
}
func (PPAPISyscallImpl) Access(path string, mode uint32) (err error) {
	panic("Access() not implemented")
}
func (PPAPISyscallImpl) Acct(path string) (err error) {
	panic("Acct() not implemented")
}
func (PPAPISyscallImpl) Adjtimex(buf *syscall.Timex) (state int, err error) {
	panic("Adjtimex() not implemented")
}
func (PPAPISyscallImpl) AttachLsf(fd int, i []syscall.SockFilter) error {
	panic("AttachLsf() not implemented")
}
func (PPAPISyscallImpl) BindToDevice(fd int, device string) (err error) {
	panic("BindToDevice() not implemented")
}
func (PPAPISyscallImpl) Chdir(path string) (err error) {
	panic("Chdir() not implemented")
}
func (PPAPISyscallImpl) Chmod(path string, mode uint32) (err error) {
	panic("Chmod() not implemented")
}
func (PPAPISyscallImpl) Chown(path string, uid int, gid int) (err error) {
	panic("Chown() not implemented")
}
func (PPAPISyscallImpl) Chroot(path string) (err error) {
	panic("Chroot() not implemented")
}
func (PPAPISyscallImpl) Clearenv() {
	panic("Clearenv() not implemented")
}
func (PPAPISyscallImpl) CmsgLen(datalen int) int {
	panic("CmsgLen() not implemented")
}
func (PPAPISyscallImpl) CmsgSpace(datalen int) int {
	panic("CmsgSpace() not implemented")
}
func (PPAPISyscallImpl) Creat(path string, mode uint32) (fd int, err error) {
	panic("Creat() not implemented")
}
func (PPAPISyscallImpl) DetachLsf(fd int) error {
	panic("DetachLsf() not implemented")
}
func (PPAPISyscallImpl) Dup3(oldfd int, newfd int, flags int) (err error) {
	panic("Dup() not implemented")
}
func (PPAPISyscallImpl) Environ() []string {
	initEnvVars()
	env := []string{}
	for key, val := range envVars {
		env = append(env, key+"="+val)
	}
	return env
}
func (PPAPISyscallImpl) EpollCreate(size int) (fd int, err error) {
	panic("EpollCreate() not implemented")
}
func (PPAPISyscallImpl) EpollCreate1(flag int) (fd int, err error) {
	panic("EpollCreate() not implemented")
}
func (PPAPISyscallImpl) EpollCtl(epfd int, op int, fd int, event *syscall.EpollEvent) (err error) {
	panic("EpollCtl() not implemented")
}
func (PPAPISyscallImpl) EpollWait(epfd int, events []syscall.EpollEvent, msec int) (n int, err error) {
	panic("EpollWait() not implemented")
}
func (PPAPISyscallImpl) Exec(argv0 string, argv []string, envv []string) (err error) {
	panic("Exec() not implemented")
}
func (PPAPISyscallImpl) Exit(code int) {
	// Wait for log outputs to finish.
	// runLock.Lock()
	// runLock.Unlock()
	if code != 0 {
		panic(fmt.Sprintf("Exited with non-zero code %d", code))
	}
	var c chan bool = make(chan bool)
	fmt.Printf("Exited with code 0, going to sleep forever.")
	<-c
}
func (PPAPISyscallImpl) Faccessat(dirfd int, path string, mode uint32, flags int) (err error) {
	panic("Faccessat() not implemented")
}
func (PPAPISyscallImpl) Fallocate(fd int, mode uint32, off int64, len int64) (err error) {
	panic("Fallocate() not implemented")
}
func (PPAPISyscallImpl) Fchdir(fd int) (err error) {
	panic("Fchdir() not implemented")
}
func (PPAPISyscallImpl) Fchmod(fd int, mode uint32) (err error) {
	panic("Fchmod() not implemented")
}
func (PPAPISyscallImpl) Fchmodat(dirfd int, path string, mode uint32, flags int) (err error) {
	panic("Fchmodat() not implemented")
}
func (PPAPISyscallImpl) Fchown(fd int, uid int, gid int) (err error) {
	panic("Fchown() not implemented")
}
func (PPAPISyscallImpl) Fchownat(dirfd int, path string, uid int, gid int, flags int) (err error) {
	panic("Fchownat() not implemented")
}
func (PPAPISyscallImpl) FcntlFlock(fd uintptr, cmd int, lk *syscall.Flock_t) error {
	panic("FcntlFlock() not implemented")
}
func (PPAPISyscallImpl) Fdatasync(fd int) (err error) {
	panic("Fdatasync() not implemented")
}
func (PPAPISyscallImpl) Flock(fd int, how int) (err error) {
	panic("Flock() not implemented")
}
func (PPAPISyscallImpl) ForkExec(argv0 string, argv []string, attr *syscall.ProcAttr) (pid int, err error) {
	panic("ForkExec() not implemented")
}
func (PPAPISyscallImpl) Fstatfs(fd int, buf *syscall.Statfs_t) (err error) {
	panic("Fstatfs() not implemented")
}
func (PPAPISyscallImpl) Ftruncate(fd int, length int64) (err error) {
	panic("Ftruncate() not implemented")
}
func (PPAPISyscallImpl) Futimes(fd int, tv []syscall.Timeval) (err error) {
	panic("Futimes() not implemented")
}
func (PPAPISyscallImpl) Futimesat(dirfd int, path string, tv []syscall.Timeval) (err error) {
	panic("Futimesat() not implemented")
}
func (PPAPISyscallImpl) Getcwd(buf []byte) (n int, err error) {
	panic("Getcwd() not implemented")
}
func (PPAPISyscallImpl) Getdents(fd int, buf []byte) (n int, err error) {
	panic("Getdents() not implemented")
}
func (PPAPISyscallImpl) Getegid() (egid int) {
	panic("Getegid() not implemented")
}
func (PPAPISyscallImpl) Getenv(key string) (value string, found bool) {
	initEnvVars()

	key = strings.ToUpper(key)
	if val, ok := envVars[key]; ok {
		return val, true
	}
	switch key {
	case "ZONEINFO", "HTTP_PROXY", "TZ", "VEYRON_PUBLICID_STORE", "VEYRON_EXEC_VERSION", "PARENT_NODE_MANAGER_NAME", "VEYRON_AGENT_FD", "VEYRON_CREDENTIALS", "VEYRON_IDENTITY":
	default:
		fmt.Printf("Getenv called on unexpected key %s (should we handle this?)", key)
	}
	return "", false
}
func (PPAPISyscallImpl) Geteuid() (euid int) {
	panic("Geteuid() not implemented")
}
func (PPAPISyscallImpl) Getgid() (gid int) {
	panic("Getgid() not implemented")
}
func (PPAPISyscallImpl) Getgroups() (gids []int, err error) {
	panic("Getgroups() not implemented")
}
func (PPAPISyscallImpl) Getpagesize() int {
	panic("Getpagesize() not implemented")
}
func (PPAPISyscallImpl) Getpgid(pid int) (pgid int, err error) {
	panic("Getpgid() not implemented")
}
func (PPAPISyscallImpl) Getpgrp() (pid int) {
	panic("Getpgrp() not implemented")
}
func (PPAPISyscallImpl) Getpid() (pid int) {
	// Same PID every time.
	return 1234
}
func (PPAPISyscallImpl) Getppid() (ppid int) {
	panic("Getppid() not implemented")
}
func (PPAPISyscallImpl) Getpriority(which int, who int) (prio int, err error) {
	panic("Getpriority() not implemented")
}
func (PPAPISyscallImpl) Getrlimit(resource int, rlim *syscall.Rlimit) (err error) {
	panic("Getrlimit() not implemented")
}
func (PPAPISyscallImpl) Getrusage(who int, rusage *syscall.Rusage) (err error) {
	panic("Getrusage() not implemented")
}
func (PPAPISyscallImpl) GetsockoptInet4Addr(fd, level, opt int) (value [4]byte, err error) {
	panic("GetsockoptInet() not implemented")
}
func (PPAPISyscallImpl) GetsockoptInt(fd, level, opt int) (value int, err error) {
	panic("GetsockoptInt() not implemented")
}
func (PPAPISyscallImpl) Gettid() (tid int) {
	panic("Gettid() not implemented")
}
func (PPAPISyscallImpl) Gettimeofday(tv *syscall.Timeval) (err error) {
	panic("Gettimeofday() not implemented")
}
func (PPAPISyscallImpl) Getuid() (uid int) {
	panic("Getuid() not implemented")
}
func (PPAPISyscallImpl) Getwd() (wd string, err error) {
	panic("Getwd() not implemented")
}
func (PPAPISyscallImpl) Getxattr(path string, attr string, dest []byte) (sz int, err error) {
	panic("Getxattr() not implemented")
}
func (PPAPISyscallImpl) InotifyAddWatch(fd int, pathname string, mask uint32) (watchdesc int, err error) {
	panic("InotifyAddWatch() not implemented")
}
func (PPAPISyscallImpl) InotifyInit() (fd int, err error) {
	panic("InotifyInit() not implemented")
}
func (PPAPISyscallImpl) InotifyInit1(flags int) (fd int, err error) {
	panic("InotifyInit() not implemented")
}
func (PPAPISyscallImpl) InotifyRmWatch(fd int, watchdesc uint32) (success int, err error) {
	panic("InotifyRmWatch() not implemented")
}
func (PPAPISyscallImpl) Ioperm(from int, num int, on int) (err error) {
	panic("Ioperm() not implemented")
}
func (PPAPISyscallImpl) Iopl(level int) (err error) {
	panic("Iopl() not implemented")
}
func (PPAPISyscallImpl) Kill(pid int, sig syscall.Signal) (err error) {
	panic("Kill() not implemented")
}
func (PPAPISyscallImpl) Klogctl(typ int, buf []byte) (n int, err error) {
	panic("Klogctl() not implemented")
}
func (PPAPISyscallImpl) Lchown(path string, uid int, gid int) (err error) {
	panic("Lchown() not implemented")
}
func (PPAPISyscallImpl) Link(oldpath string, newpath string) (err error) {
	panic("Link() not implemented")
}
func (PPAPISyscallImpl) Listxattr(path string, dest []byte) (sz int, err error) {
	panic("Listxattr() not implemented")
}
func (PPAPISyscallImpl) LsfSocket(ifindex, proto int) (int, error) {
	panic("LsfSocket() not implemented")
}
func (PPAPISyscallImpl) Lstat(path string, stat *syscall.Stat_t) (err error) {
	panic("Lstat() not implemented")
}
func (PPAPISyscallImpl) Madvise(b []byte, advice int) (err error) {
	panic("Madvise() not implemented")
}
func (PPAPISyscallImpl) Mkdirat(dirfd int, path string, mode uint32) (err error) {
	panic("Mkdirat() not implemented")
}
func (PPAPISyscallImpl) Mkfifo(path string, mode uint32) (err error) {
	panic("Mkfifo() not implemented")
}
func (PPAPISyscallImpl) Mknod(path string, mode uint32, dev int) (err error) {
	panic("Mknod() not implemented")
}
func (PPAPISyscallImpl) Mknodat(dirfd int, path string, mode uint32, dev int) (err error) {
	panic("Mknodat() not implemented")
}
func (PPAPISyscallImpl) Mlock(b []byte) (err error) {
	panic("Mlock() not implemented")
}
func (PPAPISyscallImpl) Mlockall(flags int) (err error) {
	panic("Mlockall() not implemented")
}
func (PPAPISyscallImpl) Mmap(fd int, offset int64, length int, prot int, flags int) (data []byte, err error) {
	panic("Mmap() not implemented")
}
func (PPAPISyscallImpl) Mount(source string, target string, fstype string, flags uintptr, data string) (err error) {
	panic("Mount() not implemented")
}
func (PPAPISyscallImpl) Mprotect(b []byte, prot int) (err error) {
	panic("Mprotect() not implemented")
}
func (PPAPISyscallImpl) Munlock(b []byte) (err error) {
	panic("Munlock() not implemented")
}
func (PPAPISyscallImpl) Munlockall() (err error) {
	panic("Munlockall() not implemented")
}
func (PPAPISyscallImpl) Munmap(b []byte) (err error) {
	panic("Munmap() not implemented")
}
func (PPAPISyscallImpl) Nanosleep(time *syscall.Timespec, leftover *syscall.Timespec) (err error) {
	panic("Nanosleep() not implemented")
}
func (PPAPISyscallImpl) NetlinkRIB(proto, family int) ([]byte, error) {
	panic("NetlinkRIB() not implemented")
}
func (PPAPISyscallImpl) Openat(dirfd int, path string, flags int, mode uint32) (fd int, err error) {
	panic("Openat() not implemented")
}
func (PPAPISyscallImpl) ParseDirent(buf []byte, max int, names []string) (consumed int, count int, newnames []string) {
	panic("ParseDirent() not implemented")
}
func (PPAPISyscallImpl) ParseNetlinkMessage(b []byte) ([]syscall.NetlinkMessage, error) {
	panic("ParseNetlinkMessage() not implemented")
}
func (PPAPISyscallImpl) ParseNetlinkRouteAttr(m *syscall.NetlinkMessage) ([]syscall.NetlinkRouteAttr, error) {
	panic("ParseNetlinkRouteAttr() not implemented")
}
func (PPAPISyscallImpl) ParseSocketControlMessage(b []byte) ([]syscall.SocketControlMessage, error) {
	panic("ParseSocketControlMessage() not implemented")
}
func (PPAPISyscallImpl) ParseUnixRights(m *syscall.SocketControlMessage) ([]int, error) {
	panic("ParseUnixRights() not implemented")
}
func (PPAPISyscallImpl) Pause() (err error) {
	panic("Pause() not implemented")
}
func (PPAPISyscallImpl) Pipe(p []int) (err error) {
	panic("Pipe() not implemented")
}
func (PPAPISyscallImpl) Pipe2(p []int, flags int) (err error) {
	panic("Pipe() not implemented")
}
func (PPAPISyscallImpl) PivotRoot(newroot string, putold string) (err error) {
	panic("PivotRoot() not implemented")
}
func (PPAPISyscallImpl) PtraceAttach(pid int) (err error) {
	panic("PtraceAttach() not implemented")
}
func (PPAPISyscallImpl) PtraceCont(pid int, signal int) (err error) {
	panic("PtraceCont() not implemented")
}
func (PPAPISyscallImpl) PtraceDetach(pid int) (err error) {
	panic("PtraceDetach() not implemented")
}
func (PPAPISyscallImpl) PtraceGetEventMsg(pid int) (msg uint, err error) {
	panic("PtraceGetEventMsg() not implemented")
}
func (PPAPISyscallImpl) PtraceGetRegs(pid int, regsout *syscall.PtraceRegs) (err error) {
	panic("PtraceGetRegs() not implemented")
}
func (PPAPISyscallImpl) PtracePeekData(pid int, addr uintptr, out []byte) (count int, err error) {
	panic("PtracePeekData() not implemented")
}
func (PPAPISyscallImpl) PtracePeekText(pid int, addr uintptr, out []byte) (count int, err error) {
	panic("PtracePeekText() not implemented")
}
func (PPAPISyscallImpl) PtracePokeData(pid int, addr uintptr, data []byte) (count int, err error) {
	panic("PtracePokeData() not implemented")
}
func (PPAPISyscallImpl) PtracePokeText(pid int, addr uintptr, data []byte) (count int, err error) {
	panic("PtracePokeText() not implemented")
}
func (PPAPISyscallImpl) PtraceSetOptions(pid int, options int) (err error) {
	panic("PtraceSetOptions() not implemented")
}
func (PPAPISyscallImpl) PtraceSetRegs(pid int, regs *syscall.PtraceRegs) (err error) {
	panic("PtraceSetRegs() not implemented")
}
func (PPAPISyscallImpl) PtraceSingleStep(pid int) (err error) {
	panic("PtraceSingleStep() not implemented")
}
func (PPAPISyscallImpl) PtraceSyscall(pid int, signal int) (err error) {
	panic("PtraceSyscall() not implemented")
}
func (PPAPISyscallImpl) RawSyscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno) {
	panic("RawSyscall() not implemented")
}
func (PPAPISyscallImpl) RawSyscall6(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err syscall.Errno) {
	panic("RawSyscall() not implemented")
}
func (PPAPISyscallImpl) ReadDirent(fd int, buf []byte) (n int, err error) {
	panic("ReadDirent() not implemented")
}
func (PPAPISyscallImpl) Readlink(path string, buf []byte) (n int, err error) {
	panic("Readlink() not implemented")
}
func (PPAPISyscallImpl) Reboot(cmd int) (err error) {
	panic("Reboot() not implemented")
}
func (PPAPISyscallImpl) Recvfrom(fd int, p []byte, flags int) (n int, from syscall.Sockaddr, err error) {
	panic("Recvfrom() not implemented")
}
func (PPAPISyscallImpl) Recvmsg(fd int, p, oob []byte, flags int) (n, oobn int, recvflags int, from syscall.Sockaddr, err error) {
	panic("Recvmsg() not implemented")
}
func (PPAPISyscallImpl) Removexattr(path string, attr string) (err error) {
	panic("Removexattr() not implemented")
}
func (PPAPISyscallImpl) Rename(oldpath string, newpath string) (err error) {
	panic("Rename() not implemented")
}
func (PPAPISyscallImpl) Renameat(olddirfd int, oldpath string, newdirfd int, newpath string) (err error) {
	panic("Renameat() not implemented")
}
func (PPAPISyscallImpl) Select(nfd int, r *syscall.FdSet, w *syscall.FdSet, e *syscall.FdSet, timeout *syscall.Timeval) (n int, err error) {
	panic("Select() not implemented")
}
func (PPAPISyscallImpl) Sendfile(outfd int, infd int, offset *int64, count int) (written int, err error) {
	panic("Sendfile() not implemented")
}
func (PPAPISyscallImpl) Sendmsg(fd int, p, oob []byte, to syscall.Sockaddr, flags int) (err error) {
	panic("Sendmsg() not implemented")
}
func (PPAPISyscallImpl) SendmsgN(fd int, p, oob []byte, to syscall.Sockaddr, flags int) (n int, err error) {
	panic("SendmsgN() not implemented")
}
func (PPAPISyscallImpl) Sendto(fd int, p []byte, flags int, to syscall.Sockaddr) (err error) {
	panic("Sendto() not implemented")
}
func (PPAPISyscallImpl) SetLsfPromisc(name string, m bool) error {
	panic("SetLsfPromisc() not implemented")
}
func (PPAPISyscallImpl) SetNonblock(fd int, nonblocking bool) (err error) {
	//Ignore...
	//panic("SetNonblock() not implemented")
	return nil
}
func (PPAPISyscallImpl) Setdomainname(p []byte) (err error) {
	panic("Setdomainname() not implemented")
}
func (PPAPISyscallImpl) Setenv(key, value string) error {
	key = strings.ToUpper(key)
	envVars[key] = value
	return nil
}
func (PPAPISyscallImpl) Setfsgid(gid int) (err error) {
	panic("Setfsgid() not implemented")
}
func (PPAPISyscallImpl) Setfsuid(uid int) (err error) {
	panic("Setfsuid() not implemented")
}
func (PPAPISyscallImpl) Setgid(gid int) (err error) {
	panic("Setgid() not implemented")
}
func (PPAPISyscallImpl) Setgroups(gids []int) (err error) {
	panic("Setgroups() not implemented")
}
func (PPAPISyscallImpl) Sethostname(p []byte) (err error) {
	panic("Sethostname() not implemented")
}
func (PPAPISyscallImpl) Setpgid(pid int, pgid int) (err error) {
	panic("Setpgid() not implemented")
}
func (PPAPISyscallImpl) Setpriority(which int, who int, prio int) (err error) {
	panic("Setpriority() not implemented")
}
func (PPAPISyscallImpl) Setregid(rgid int, egid int) (err error) {
	panic("Setregid() not implemented")
}
func (PPAPISyscallImpl) Setresgid(rgid int, egid int, sgid int) (err error) {
	panic("Setresgid() not implemented")
}
func (PPAPISyscallImpl) Setresuid(ruid int, euid int, suid int) (err error) {
	panic("Setresuid() not implemented")
}
func (PPAPISyscallImpl) Setreuid(ruid int, euid int) (err error) {
	panic("Setreuid() not implemented")
}
func (PPAPISyscallImpl) Setrlimit(resource int, rlim *syscall.Rlimit) (err error) {
	panic("Setrlimit() not implemented")
}
func (PPAPISyscallImpl) Setsid() (pid int, err error) {
	panic("Setsid() not implemented")
}
func (PPAPISyscallImpl) SetsockoptByte(fd, level, opt int, value byte) (err error) {
	panic("SetsockoptByte() not implemented")
}
func (PPAPISyscallImpl) SetsockoptICMPv6Filter(fd, level, opt int, filter *syscall.ICMPv6Filter) error {
	panic("SetsockoptICMPv() not implemented")
}
func (PPAPISyscallImpl) SetsockoptIPMreq(fd, level, opt int, mreq *syscall.IPMreq) (err error) {
	panic("SetsockoptIPMreq() not implemented")
}
func (PPAPISyscallImpl) SetsockoptIPMreqn(fd, level, opt int, mreq *syscall.IPMreqn) (err error) {
	panic("SetsockoptIPMreqn() not implemented")
}
func (PPAPISyscallImpl) SetsockoptIPv6Mreq(fd, level, opt int, mreq *syscall.IPv6Mreq) (err error) {
	panic("SetsockoptIPv() not implemented")
}
func (PPAPISyscallImpl) SetsockoptInet4Addr(fd, level, opt int, value [4]byte) (err error) {
	panic("SetsockoptInet() not implemented")
}
func (PPAPISyscallImpl) SetsockoptInt(fd, level, opt int, value int) (err error) {
	return nil
}
func (PPAPISyscallImpl) SetsockoptLinger(fd, level, opt int, l *syscall.Linger) (err error) {
	panic("SetsockoptLinger() not implemented")
}
func (PPAPISyscallImpl) SetsockoptString(fd, level, opt int, s string) (err error) {
	panic("SetsockoptString() not implemented")
}
func (PPAPISyscallImpl) SetsockoptTimeval(fd, level, opt int, tv *syscall.Timeval) (err error) {
	panic("SetsockoptTimeval() not implemented")
}
func (PPAPISyscallImpl) Settimeofday(tv *syscall.Timeval) (err error) {
	panic("Settimeofday() not implemented")
}
func (PPAPISyscallImpl) Setuid(uid int) (err error) {
	panic("Setuid() not implemented")
}
func (PPAPISyscallImpl) Setxattr(path string, attr string, data []byte, flags int) (err error) {
	panic("Setxattr() not implemented")
}
func (PPAPISyscallImpl) Shutdown(fd int, how int) (err error) {
	panic("Shutdown() not implemented")
}
func (PPAPISyscallImpl) SlicePtrFromStrings(ss []string) ([]*byte, error) {
	panic("SlicePtrFromStrings() not implemented")
}
func (PPAPISyscallImpl) Socketpair(domain, typ, proto int) (fd [2]int, err error) {
	panic("Socketpair() not implemented")
}
func (PPAPISyscallImpl) Splice(rfd int, roff *int64, wfd int, woff *int64, len int, flags int) (n int64, err error) {
	panic("Splice() not implemented")
}
func (PPAPISyscallImpl) StartProcess(argv0 string, argv []string, attr *syscall.ProcAttr) (pid int, handle uintptr, err error) {
	panic("StartProcess() not implemented")
}
func (PPAPISyscallImpl) Statfs(path string, buf *syscall.Statfs_t) (err error) {
	panic("Statfs() not implemented")
}
func (PPAPISyscallImpl) StringSlicePtr(ss []string) []*byte {
	panic("StringSlicePtr() not implemented")
}
func (PPAPISyscallImpl) Sync() {
	panic("Sync() not implemented")
}
func (PPAPISyscallImpl) SyncFileRange(fd int, off int64, n int64, flags int) (err error) {
	panic("SyncFileRange() not implemented")
}
func (PPAPISyscallImpl) Syscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno) {
	panic("Syscall() not implemented")
}
func (PPAPISyscallImpl) Syscall6(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err syscall.Errno) {
	panic("Syscall() not implemented")
}
func (PPAPISyscallImpl) Sysinfo(info *syscall.Sysinfo_t) (err error) {
	panic("Sysinfo() not implemented")
}
func (PPAPISyscallImpl) Tee(rfd int, wfd int, len int, flags int) (n int64, err error) {
	panic("Tee() not implemented")
}
func (PPAPISyscallImpl) Tgkill(tgid int, tid int, sig syscall.Signal) (err error) {
	panic("Tgkill() not implemented")
}
func (PPAPISyscallImpl) Times(tms *syscall.Tms) (ticks uintptr, err error) {
	panic("Times() not implemented")
}
func (PPAPISyscallImpl) TimespecToNsec(ts syscall.Timespec) int64 {
	panic("TimespecToNsec() not implemented")
}
func (PPAPISyscallImpl) TimevalToNsec(tv syscall.Timeval) int64 {
	panic("TimevalToNsec() not implemented")
}
func (PPAPISyscallImpl) Truncate(path string, length int64) (err error) {
	panic("Truncate() not implemented")
}
func (PPAPISyscallImpl) Umask(mask int) (oldmask int) {
	panic("Umask() not implemented")
}
func (PPAPISyscallImpl) Uname(buf *syscall.Utsname) (err error) {
	panic("Uname() not implemented")
}
func (PPAPISyscallImpl) UnixCredentials(ucred *syscall.Ucred) []byte {
	panic("UnixCredentials() not implemented")
}
func (PPAPISyscallImpl) UnixRights(fds ...int) []byte {
	panic("UnixRights() not implemented")
}
func (PPAPISyscallImpl) Unlinkat(dirfd int, path string) (err error) {
	panic("Unlinkat() not implemented")
}
func (PPAPISyscallImpl) Unmount(target string, flags int) (err error) {
	panic("Unmount() not implemented")
}
func (PPAPISyscallImpl) Unsetenv(key string) error {
	key = strings.ToUpper(key)
	delete(envVars, key)
	return nil
}
func (PPAPISyscallImpl) Unshare(flags int) (err error) {
	panic("Unshare() not implemented")
}
func (PPAPISyscallImpl) Ustat(dev int, ubuf *syscall.Ustat_t) (err error) {
	panic("Ustat() not implemented")
}
func (PPAPISyscallImpl) Utime(path string, buf *syscall.Utimbuf) (err error) {
	panic("Utime() not implemented")
}
func (PPAPISyscallImpl) Utimes(path string, tv []syscall.Timeval) (err error) {
	panic("Utimes() not implemented")
}
func (PPAPISyscallImpl) UtimesNano(path string, ts []syscall.Timespec) (err error) {
	panic("UtimesNano() not implemented")
}
func (PPAPISyscallImpl) Wait4(pid int, wstatus *syscall.WaitStatus, options int, rusage *syscall.Rusage) (wpid int, err error) {
	panic("Wait() not implemented")
}

func (PPAPISyscallImpl) GetsockoptICMPv6Filter(fd, level, opt int) (*syscall.ICMPv6Filter, error) {
	panic("GetsockoptICMPv() not implemented")
}
func (PPAPISyscallImpl) GetsockoptIPMreq(fd, level, opt int) (*syscall.IPMreq, error) {
	panic("GetsockoptIPMreq() not implemented")
}
func (PPAPISyscallImpl) GetsockoptIPMreqn(fd, level, opt int) (*syscall.IPMreqn, error) {
	panic("GetsockoptIPMreqn() not implemented")
}
func (PPAPISyscallImpl) GetsockoptIPv6MTUInfo(fd, level, opt int) (*syscall.IPv6MTUInfo, error) {
	panic("GetsockoptIPv() not implemented")
}
func (PPAPISyscallImpl) GetsockoptIPv6Mreq(fd, level, opt int) (*syscall.IPv6Mreq, error) {
	panic("GetsockoptIPv() not implemented")
}
func (PPAPISyscallImpl) LsfJump(code, k, jt, jf int) *syscall.SockFilter {
	panic("LsfJump() not implemented")
}
func (PPAPISyscallImpl) LsfStmt(code, k int) *syscall.SockFilter {
	panic("LsfStmt() not implemented")
}
func (PPAPISyscallImpl) Time(t *syscall.Time_t) (tt syscall.Time_t, err error) {
	panic("Time() not implemented")
}
func (PPAPISyscallImpl) NsecToTimespec(nsec int64) (ts syscall.Timespec) {
	panic("NsecToTimespec() not implemented")
}
func (PPAPISyscallImpl) NsecToTimeval(nsec int64) (tv syscall.Timeval) {
	panic("NsecToTimeval() not implemented")
}
func (PPAPISyscallImpl) GetsockoptUcred(fd, level, opt int) (*syscall.Ucred, error) {
	panic("GetsockoptUcred() not implemented")
}
func (PPAPISyscallImpl) ParseUnixCredentials(m *syscall.SocketControlMessage) (*syscall.Ucred, error) {
	panic("ParseUnixCredentials() not implemented")
}

func (PPAPISyscallImpl) SetReadDeadline(fd int, t int64) error {
	//panic("SetReadDeadline() not implemented")
	fmt.Printf("SetReadDeadline() ignored")
	return nil
}
func (PPAPISyscallImpl) SetWriteDeadline(fd int, t int64) error {
	//panic("SetWriteDeadline() not implemented")
	fmt.Printf("SetWriteDeadline() ignored")
	return nil
}

// For type checking
var _ syscall.SyscallImplementation = PPAPISyscallImpl{}
