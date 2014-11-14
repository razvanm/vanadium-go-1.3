package syscall

// This file contains type definitions that are necessary for a syscall implementation.
// None of the methods on the types are implemented.

type Cmsghdr struct {
	Len          uint64
	Level        int32
	Type         int32
	X__cmsg_data [0]uint8
}

func (cmsg *Cmsghdr) SetLen(length int) {
	panic("SetLen() not implemented")
}

type Credential struct {
	Uid    uint32   // User ID.
	Gid    uint32   // Group ID.
	Groups []uint32 // Supplementary group IDs.
}
type Dirent struct {
	Ino       uint64
	Off       int64
	Reclen    uint16
	Type      uint8
	Name      [256]int8
	Pad_cgo_0 [5]byte
}
type EpollEvent struct {
	Events uint32
	Fd     int32
	Pad    int32
}
type FdSet struct {
	Bits [16]int64
}
type Flock_t struct {
	Type      int16
	Whence    int16
	Pad_cgo_0 [4]byte
	Start     int64
	Len       int64
	Pid       int32
	Pad_cgo_1 [4]byte
}
type Fsid struct {
	X__val [2]int32
}
type ICMPv6Filter struct {
	Data [8]uint32
}
type IPMreq struct {
	Multiaddr [4]byte /* in_addr */
	Interface [4]byte /* in_addr */
}
type IPMreqn struct {
	Multiaddr [4]byte /* in_addr */
	Address   [4]byte /* in_addr */
	Ifindex   int32
}
type IPv6MTUInfo struct {
	Addr RawSockaddrInet6
	Mtu  uint32
}
type IPv6Mreq struct {
	Multiaddr [16]byte /* in6_addr */
	Interface uint32
}
type IfAddrmsg struct {
	Family    uint8
	Prefixlen uint8
	Flags     uint8
	Scope     uint8
	Index     uint32
}
type IfInfomsg struct {
	Family     uint8
	X__ifi_pad uint8
	Type       uint16
	Index      int32
	Flags      uint32
	Change     uint32
}
type Inet4Pktinfo struct {
	Ifindex  int32
	Spec_dst [4]byte /* in_addr */
	Addr     [4]byte /* in_addr */
}
type Inet6Pktinfo struct {
	Addr    [16]byte /* in6_addr */
	Ifindex uint32
}
type InotifyEvent struct {
	Wd     int32
	Mask   uint32
	Cookie uint32
	Len    uint32
	Name   [0]uint8
}
type Iovec struct {
	Base *byte
	Len  uint64
}

func (iov *Iovec) SetLen(length int) {
	panic("() not implemented")
}

type Linger struct {
	Onoff  int32
	Linger int32
}
type Msghdr struct {
	Name       *byte
	Namelen    uint32
	Pad_cgo_0  [4]byte
	Iov        *Iovec
	Iovlen     uint64
	Control    *byte
	Controllen uint64
	Flags      int32
	Pad_cgo_1  [4]byte
}

func (msghdr *Msghdr) SetControllen(length int) {
	panic("() not implemented")
}

type NetlinkMessage struct {
	Header NlMsghdr
	Data   []byte
}
type NetlinkRouteAttr struct {
	Attr  RtAttr
	Value []byte
}
type NetlinkRouteRequest struct {
	Header NlMsghdr
	Data   RtGenmsg
}
type NlAttr struct {
	Len  uint16
	Type uint16
}
type NlMsgerr struct {
	Error int32
	Msg   NlMsghdr
}
type NlMsghdr struct {
	Len   uint32
	Type  uint16
	Flags uint16
	Seq   uint32
	Pid   uint32
}
type ProcAttr struct {
	Dir   string    // Current working directory.
	Env   []string  // Environment.
	Files []uintptr // File descriptors.
	Sys   *SysProcAttr
}
type PtraceRegs struct {
	R15      uint64
	R14      uint64
	R13      uint64
	R12      uint64
	Rbp      uint64
	Rbx      uint64
	R11      uint64
	R10      uint64
	R9       uint64
	R8       uint64
	Rax      uint64
	Rcx      uint64
	Rdx      uint64
	Rsi      uint64
	Rdi      uint64
	Orig_rax uint64
	Rip      uint64
	Cs       uint64
	Eflags   uint64
	Rsp      uint64
	Ss       uint64
	Fs_base  uint64
	Gs_base  uint64
	Ds       uint64
	Es       uint64
	Fs       uint64
	Gs       uint64
}

func (r *PtraceRegs) PC() uint64 {
	panic("PC() not implemented")
}
func (r *PtraceRegs) SetPC(pc uint64) {
	panic("SetPC() not implemented")
}

type RawSockaddr struct {
	Family uint16
	Data   [14]int8
}
type RawSockaddrAny struct {
	Addr RawSockaddr
	Pad  [96]int8
}
type RawSockaddrInet4 struct {
	Family uint16
	Port   uint16
	Addr   [4]byte /* in_addr */
	Zero   [8]uint8
}
type RawSockaddrInet6 struct {
	Family   uint16
	Port     uint16
	Flowinfo uint32
	Addr     [16]byte /* in6_addr */
	Scope_id uint32
}
type RawSockaddrLinklayer struct {
	Family   uint16
	Protocol uint16
	Ifindex  int32
	Hatype   uint16
	Pkttype  uint8
	Halen    uint8
	Addr     [8]uint8
}
type RawSockaddrNetlink struct {
	Family uint16
	Pad    uint16
	Pid    uint32
	Groups uint32
}
type RawSockaddrUnix struct {
	Family uint16
	Path   [108]int8
}
type Rlimit struct {
	Cur uint64
	Max uint64
}
type RtAttr struct {
	Len  uint16
	Type uint16
}
type RtGenmsg struct {
	Family uint8
}
type RtMsg struct {
	Family   uint8
	Dst_len  uint8
	Src_len  uint8
	Tos      uint8
	Table    uint8
	Protocol uint8
	Scope    uint8
	Type     uint8
	Flags    uint32
}
type RtNexthop struct {
	Len     uint16
	Flags   uint8
	Hops    uint8
	Ifindex int32
}
type Rusage struct {
	Utime    Timeval
	Stime    Timeval
	Maxrss   int64
	Ixrss    int64
	Idrss    int64
	Isrss    int64
	Minflt   int64
	Majflt   int64
	Nswap    int64
	Inblock  int64
	Oublock  int64
	Msgsnd   int64
	Msgrcv   int64
	Nsignals int64
	Nvcsw    int64
	Nivcsw   int64
}

type SockFilter struct {
	Code uint16
	Jt   uint8
	Jf   uint8
	K    uint32
}
type SockFprog struct {
	Len       uint16
	Pad_cgo_0 [6]byte
	Filter    *SockFilter
}
type Sockaddr interface {
}
type SockaddrInet4 struct {
	Port int
	Addr [4]byte
	// contains filtered or unexported fields
}
type SockaddrInet6 struct {
	Port   int
	ZoneId uint32
	Addr   [16]byte
	// contains filtered or unexported fields
}
type SockaddrLinklayer struct {
	Protocol uint16
	Ifindex  int
	Hatype   uint16
	Pkttype  uint8
	Halen    uint8
	Addr     [8]byte
	// contains filtered or unexported fields
}
type SockaddrNetlink struct {
	Family uint16
	Pad    uint16
	Pid    uint32
	Groups uint32
	// contains filtered or unexported fields
}
type SockaddrUnix struct {
	Name string
	// contains filtered or unexported fields
}
type SocketControlMessage struct {
	Header Cmsghdr
	Data   []byte
}
type Stat_t struct {
	Dev       int64
	Ino       uint64
	Mode      uint32
	Nlink     uint32
	Uid       uint32
	Gid       uint32
	Rdev      int64
	Size      int64
	Blksize   int32
	Blocks    int32
	Atime     int64
	AtimeNsec int64
	Mtime     int64
	MtimeNsec int64
	Ctime     int64
	CtimeNsec int64
}
type Statfs_t struct {
	Type    int64
	Bsize   int64
	Blocks  uint64
	Bfree   uint64
	Bavail  uint64
	Files   uint64
	Ffree   uint64
	Fsid    Fsid
	Namelen int64
	Frsize  int64
	Flags   int64
	Spare   [4]int64
}
type SysProcAttr struct {
	Chroot     string      // Chroot.
	Credential *Credential // Credential.
	Ptrace     bool        // Enable tracing.
	Setsid     bool        // Create session.
	Setpgid    bool        // Set process group ID to new pid (SYSV setpgrp)
	Setctty    bool        // Set controlling terminal to fd Ctty (only meaningful if Setsid is set)
	Noctty     bool        // Detach fd 0 from controlling terminal
	Ctty       int         // Controlling TTY fd (Linux only)
	Pdeathsig  Signal      // Signal that the process will get when its parent dies (Linux only)
	Cloneflags uintptr     // Flags for clone calls (Linux only)
}
type Sysinfo_t struct {
	Uptime    int64
	Loads     [3]uint64
	Totalram  uint64
	Freeram   uint64
	Sharedram uint64
	Bufferram uint64
	Totalswap uint64
	Freeswap  uint64
	Procs     uint16
	Pad       uint16
	Pad_cgo_0 [4]byte
	Totalhigh uint64
	Freehigh  uint64
	Unit      uint32
	X_f       [0]byte
	Pad_cgo_1 [4]byte
}
type TCPInfo struct {
	State          uint8
	Ca_state       uint8
	Retransmits    uint8
	Probes         uint8
	Backoff        uint8
	Options        uint8
	Pad_cgo_0      [2]byte
	Rto            uint32
	Ato            uint32
	Snd_mss        uint32
	Rcv_mss        uint32
	Unacked        uint32
	Sacked         uint32
	Lost           uint32
	Retrans        uint32
	Fackets        uint32
	Last_data_sent uint32
	Last_ack_sent  uint32
	Last_data_recv uint32
	Last_ack_recv  uint32
	Pmtu           uint32
	Rcv_ssthresh   uint32
	Rtt            uint32
	Rttvar         uint32
	Snd_ssthresh   uint32
	Snd_cwnd       uint32
	Advmss         uint32
	Reordering     uint32
	Rcv_rtt        uint32
	Rcv_space      uint32
	Total_retrans  uint32
}
type Termios struct {
	Iflag     uint32
	Oflag     uint32
	Cflag     uint32
	Lflag     uint32
	Line      uint8
	Cc        [32]uint8
	Pad_cgo_0 [3]byte
	Ispeed    uint32
	Ospeed    uint32
}
type Time_t int64
type Timespec struct {
	Sec  int64
	Nsec int64
}

type Timeval struct {
	Sec  int64
	Usec int64
}

type Timex struct {
	Modes     uint32
	Pad_cgo_0 [4]byte
	Offset    int64
	Freq      int64
	Maxerror  int64
	Esterror  int64
	Status    int32
	Pad_cgo_1 [4]byte
	Constant  int64
	Precision int64
	Tolerance int64
	Time      Timeval
	Tick      int64
	Ppsfreq   int64
	Jitter    int64
	Shift     int32
	Pad_cgo_2 [4]byte
	Stabil    int64
	Jitcnt    int64
	Calcnt    int64
	Errcnt    int64
	Stbcnt    int64
	Tai       int32
	Pad_cgo_3 [44]byte
}
type Tms struct {
	Utime  int64
	Stime  int64
	Cutime int64
	Cstime int64
}
type Ucred struct {
	Pid int32
	Uid uint32
	Gid uint32
}
type Ustat_t struct {
	Tfree     int32
	Pad_cgo_0 [4]byte
	Tinode    uint64
	Fname     [6]int8
	Fpack     [6]int8
	Pad_cgo_1 [4]byte
}
type Utimbuf struct {
	Actime  int64
	Modtime int64
}
type Utsname struct {
	Sysname    [65]int8
	Nodename   [65]int8
	Release    [65]int8
	Version    [65]int8
	Machine    [65]int8
	Domainname [65]int8
}
type WaitStatus uint32

func (w WaitStatus) Continued() bool {
	panic("() not implemented")
}
func (w WaitStatus) CoreDump() bool {
	panic("() not implemented")
}
func (w WaitStatus) ExitStatus() int {
	panic("() not implemented")
}
func (w WaitStatus) Exited() bool {
	panic("() not implemented")
}
func (w WaitStatus) Signal() Signal {
	panic("() not implemented")
}
func (w WaitStatus) Signaled() bool {
	panic("() not implemented")
}
func (w WaitStatus) StopSignal() Signal {
	panic("() not implemented")
}
func (w WaitStatus) Stopped() bool {
	panic("() not implemented")
}
func (w WaitStatus) TrapCause() int {
	panic("() not implemented")
}

// A Signal is a number describing a process signal.
// It implements the os.Signal interface.
type Signal int

func (s Signal) Signal() {}

func (s Signal) String() string {
	if 0 <= s && int(s) < len(signals) {
		str := signals[s]
		if str != "" {
			return str
		}
	}
	return "signal " + itoa(int(s))
}
