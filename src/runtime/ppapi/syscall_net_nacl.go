// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppapi

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"syscall"
)

type netConnFile struct {
	conn      net.Conn
	listener  net.Listener
	protoType string
	addr      string // set by Bind()
}

func (*netConnFile) stat(*syscall.Stat_t) error {
	panic("stat not yet implemented")
}
func (n *netConnFile) read(b []byte) (int, error) {
	count, err := n.conn.Read(b)
	if err != nil && err != io.EOF {
		switch err.Error() {
		case "connection reset", "aborted", "Reading from closed connection":
			return 0, err
		}
		panic(fmt.Sprintf("Failed to read: %v", err))
	}
	remoteaddr := n.conn.RemoteAddr()
	if err == io.EOF {
		fmt.Printf("Network read to %v resulted in EOF!", remoteaddr)
	} else {
		fmt.Printf("Network read to %v: %s", remoteaddr, string(b[:count]))
	}
	return count, err
}
func (n *netConnFile) write(b []byte) (int, error) {
	count, err := n.conn.Write(b)
	if err != nil && err != io.EOF {
		switch err.Error() {
		case "connection reset", "aborted", "Reading from closed connection":
			return 0, err
		}
		panic(fmt.Sprintf("Failed to write: %v", err))
	}
	remoteaddr := n.conn.RemoteAddr()
	if err == io.EOF {
		fmt.Printf("Network write to %v resulted in EOF!", remoteaddr)
	} else {
		fmt.Printf("Network write to %v: %s", remoteaddr, string(b[:count]))
	}
	return count, err
}
func (*netConnFile) seek(int64, int) (int64, error) {
	panic("seek not yet implemented")
}
func (*netConnFile) pread([]byte, int64) (int, error) {
	panic("pread not yet implemented")
}
func (*netConnFile) pwrite([]byte, int64) (int, error) {
	panic("pwrite not yet implemented")
}
func (n *netConnFile) close() error {
	if n.conn != nil {
		return n.conn.Close()
	} else if n.listener != nil {
		return n.listener.Close()
	}
	return nil
}

func fdToNetConnFile(fd int) (*netConnFile, error) {
	f, err := fdToFile(fd)
	if err != nil {
		return nil, err
	}
	if netFile, ok := f.(*netConnFile); ok {
		return netFile, nil
	}
	return nil, fmt.Errorf("FD: %v Resolved to a file that does not represent a net connection. Type: %T", fd, f)
}

func sockaddrToStr(sa syscall.Sockaddr, allowZeroPort bool) (string, error) {
	switch vsa := sa.(type) {
	case *syscall.SockaddrInet4:
		ip := net.IP(vsa.Addr[:])
		if vsa.Port == 0 && !allowZeroPort {
			return "", fmt.Errorf("No port specified for %s", ip.String())
		}
		return fmt.Sprintf("%v:%d", ip, vsa.Port), nil
	case *syscall.SockaddrInet6:
		ip := net.IP(vsa.Addr[:])
		if vsa.Port == 0 && !allowZeroPort {
			return "", fmt.Errorf("No port specified for %s", ip.String())
		}
		return fmt.Sprintf("[%v]:%d", ip, vsa.Port), nil
	default:
		panic(fmt.Sprintf("Unsupported Sockaddr type: %v", vsa))
	}
}

func strToSockaddr(addr string) (syscall.Sockaddr, error) {
	hostStr, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		hostStr = addr
		portStr = ""
	}

	ip := net.ParseIP(hostStr)
	if ip == nil {
		return nil, fmt.Errorf("Failed to parse address: %s", addr)
	}

	uport, err := strconv.ParseUint(portStr, 0, 16)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse port: %d", uport)
	}

	port := int(uport)

	ipv4 := ip.To4()
	if ipv4 != nil {
		if len(ipv4) != 4 {
			panic("Invalid length for ipv4 address")
		}
		var addrArray [4]byte
		copy(addrArray[:], []byte(ipv4))
		return &syscall.SockaddrInet4{
			Port: port,
			Addr: addrArray,
		}, nil
	}

	ipv6 := ip.To16()
	if ipv6 != nil {
		if len(ipv6) != 16 {
			panic("Invalid length for ipv6 address")
		}
		var addrArray [16]byte
		copy(addrArray[:], []byte(ipv6))
		return &syscall.SockaddrInet6{
			Port: port,
			Addr: addrArray,
		}, nil
	}

	panic("Invalid parsed address")
}

func (PPAPISyscallImpl) Socket(domain, typ, proto int) (fd int, err error) {
	fmt.Printf("Socket(%v, %v, %v)", domain, typ, proto)
	switch domain {
	case syscall.AF_INET, syscall.AF_INET6:
	default:
		return -1, syscall.EPROTONOSUPPORT
	}
	var protoType string
	if typ == syscall.SOCK_STREAM {
		protoType = "tcp"
	} else if typ == syscall.SOCK_DGRAM {
		protoType = "udp"
	} else {
		return -1, syscall.ESOCKTNOSUPPORT
	}
	if proto != 0 {
		return -1, syscall.EPROTONOSUPPORT
	}
	f := &netConnFile{
		protoType: protoType,
	}
	return newFD(f), nil
}

func (PPAPISyscallImpl) Bind(fd int, sa syscall.Sockaddr) (err error) {
	netFile, err := fdToNetConnFile(fd)
	if err != nil {
		return err
	}
	addrStr, err := sockaddrToStr(sa, true)
	if err != nil {
		return err
	}
	fmt.Printf("Binding to %s", addrStr)
	netFile.addr = addrStr
	return nil
}

func (pi PPAPISyscallImpl) Listen(fd int, backlog int) (err error) {
	// backlog is ignored
	netFile, err := fdToNetConnFile(fd)
	if err != nil {
		return err
	}
	if netFile.addr == "" {
		panic("Bind() must be called before listen")
	}
	fmt.Printf("Listening on address: %s", netFile.addr)
	listener, err := pi.Instance.Listen(netFile.protoType, netFile.addr)
	if err != nil {
		return err
	}
	netFile.listener = listener
	return nil
}

func (PPAPISyscallImpl) Accept(fd int) (nfd int, sa syscall.Sockaddr, err error) {
	netFile, err := fdToNetConnFile(fd)
	if err != nil {
		return -1, nil, err
	}
	conn, err := netFile.listener.Accept()
	if err != nil {
		return -1, nil, err
	}
	connFd := newFD(&netConnFile{
		conn: conn,
	})
	addr := conn.RemoteAddr()
	if addr == nil {
		panic("Failed to get remote endpoint address in accept. This should always be possible.")
	}
	straddr := addr.String()
	fmt.Printf("straddr: %v", straddr)
	sa, err = strToSockaddr(straddr)
	if err != nil {
		return -1, nil, err
	}
	return connFd, sa, nil
}

func (pi PPAPISyscallImpl) Connect(fd int, sa syscall.Sockaddr) (err error) {
	netFile, err := fdToNetConnFile(fd)
	if err != nil {
		return err
	}
	addrStr, err := sockaddrToStr(sa, false)
	if err != nil {
		return err
	}
	fmt.Printf("Connecting to %s", addrStr)
	conn, err := pi.Instance.Dial(netFile.protoType, addrStr)
	if err != nil {
		fmt.Printf("Error connecting to %s: %v", addrStr, err)
		return err
	}
	netFile.conn = conn
	return nil
}

func (PPAPISyscallImpl) Getsockname(fd int) (sa syscall.Sockaddr, err error) {
	netFile, err := fdToNetConnFile(fd)
	if err != nil {
		return nil, err
	}
	if netFile.conn == nil && netFile.listener == nil {
	}
	var addr net.Addr
	if netFile.listener != nil {
		addr = netFile.listener.Addr()
	} else if netFile.conn != nil {
		addr = netFile.conn.LocalAddr()
	} else {
		return nil, fmt.Errorf("Cannot Getsockname on unconnected socket.")
	}
	if addr == nil {
		panic("Addr is nil")
		//return nil, nil
	}
	return strToSockaddr(addr.String())
}

func (PPAPISyscallImpl) Getpeername(fd int) (sa syscall.Sockaddr, err error) {
	netFile, err := fdToNetConnFile(fd)
	if err != nil {
		return nil, err
	}
	var addr net.Addr
	if netFile.conn != nil {
		addr = netFile.conn.RemoteAddr()
	} else {
		return nil, fmt.Errorf("Cannot Getpeername on unconnected socket.")
	}
	if addr == nil {
		panic("Addr is nil")
		//return nil, nil
	}
	return strToSockaddr(addr.String())
}

func (PPAPISyscallImpl) StopIO(fd int) error {
	f, err := fdToNetConnFile(fd)
	if err != nil {
		return err
	}
	f.close()
	return nil
}
