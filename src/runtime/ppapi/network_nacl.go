// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppapi

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

var (
	errCreateAddressFailed      = errors.New("CreateAddress failed")
	errMalformedAddress         = errors.New("malformed network address")
	errLocalAddrFailed          = errors.New("LocalAddr failed")
	errRemoteAddrFailed         = errors.New("RemoteAddr failed")
	errDeadlineNotSupported     = errors.New("deadlines are not supported")
	errCreateTCPSocketFailed    = errors.New("could not create TCP socket")
	errCreateUDPSocketFailed    = errors.New("could not create UDP socket")
	errNotConnected             = errors.New("not connected")
	errCreateHostResolverFailed = errors.New("CreateHostResolver failed")
	errCanonicalNameFailed      = errors.New("CanonicalName failed")
	errCanonNameFlagNotSet      = errors.New("PP_HOSTRESOLVER_FLAG_CANONNAME not set in HostResolverHint")
	errHostResolverFailed       = errors.New("host resolution failed")
)

// HostResolver supports host name resolution.
//
// It isn't normally necessary to use the HostResolver directly, since the Dial
// and Listen methods perform host resolution automatically.
type HostResolver struct {
	Resource
}

// HostResolveHint represents hints for host resolution.
type HostResolverHint struct {
	Family NetAddressFamily
	Flags  int32
}

// CreateHostResolver creates a HostResolver.
func (inst Instance) CreateHostResolver() (resolver HostResolver, err error) {
	resolver.id = ppb_hostresolver_create(inst.id)
	if resolver.id == 0 {
		err = errCreateHostResolverFailed
	}
	return
}

// Resolve requests resolution of a host name.
//
// If the call completes successfully, the results can be retrieved by
// CanonicalName(), NetAddressCount() and NetAddress().
func (resolver HostResolver) Resolve(host string, port uint16, hint *HostResolverHint) error {
	name := append([]byte(host), 0)
	code := ppb_hostresolver_resolve(
		resolver.id, &name[0], port,
		(*pp_HostResolverHint)(unsafe.Pointer(hint)), ppNullCompletionCallback)
	return decodeError(Error(code))
}

// CanonicalName gets the canonical name of the host.
func (resolver HostResolver) GetCanonicalName() (s string, err error) {
	var ppVar pp_Var
	ppb_hostresolver_get_canonical_name(&ppVar, resolver.id)
	var v Var
	v.fromPP(ppVar)
	s, err = v.AsString()
	if err != nil {
		err = errCanonicalNameFailed
		return
	}
	if s == "" {
		err = errCanonNameFlagNotSet
		return
	}
	return
}

// NetAddresses returns the resolved network addresses.
func (resolver HostResolver) NetAddresses() []NetAddress {
	count := ppb_hostresolver_get_net_address_count(resolver.id)
	var addresses []NetAddress
	for i := uint32(0); i != count; i++ {
		var addr NetAddress
		addr.id = ppb_hostresolver_get_net_address(resolver.id, i)
		if addr.id != 0 {
			addresses = append(addresses, addr)
		}
	}
	return addresses
}

// ResolveNetAddress resolves a network address.
func (inst Instance) ResolveNetAddress(network, addr string) (na NetAddress, err error) {
	if strings.HasPrefix(addr, "[::]:") {
		// TODO(bprosnitz) We really shouldn't have to do this. Chrome won't resolve the IPv6
		// address for some reason. Fix this.
		addr = strings.Replace(addr, "[::]:", "127.0.0.1:", 1)
	}
	host, strport, err := net.SplitHostPort(addr)
	if err != nil {
		panic(fmt.Sprintf("Failed to resolve 1 %s: %s", addr, err))
		return
	}
	port, err := strconv.Atoi(strport)
	if err != nil {
		panic(fmt.Sprintf("Failed to resolve 2 %s: %s", addr, err))

		return
	}
	var hint HostResolverHint
	switch network {
	case "tcp", "udp":
		hint.Family = PP_NETADDRESS_FAMILY_UNSPECIFIED
	case "tcp4", "udp4":
		hint.Family = PP_NETADDRESS_FAMILY_IPV4
	case "tcp6", "udp6":
		hint.Family = PP_NETADDRESS_FAMILY_IPV6
	default:
		err = fmt.Errorf("unsupported network: %q", network)
		return
	}
	resolver, err := inst.CreateHostResolver()
	if err != nil {
		panic(fmt.Sprintf("Failed to resolve 3 %s: %s", addr, err))

		return
	}
	defer resolver.Release()
	if err = resolver.Resolve(host, uint16(port), &hint); err != nil {
		panic(fmt.Sprintf("Failed to resolve 4 %s: %s", addr, err))
		return
	}
	addrs := resolver.NetAddresses()
	if len(addrs) == 0 {
		err = errHostResolverFailed
		panic(fmt.Sprintf("Failed to resolve 5 %s: %s", addr, err))
		return
	}
	for _, addr := range addrs[1:] {
		addr.Release()
	}
	na = addrs[0]
	return
}

// ResolveTCPAddr parses addr as a TCP address of the form "host:port" or
// "[ipv6-host%zone]:port" and resolves a pair of domain name and port name on
// the network net, which must be "tcp", "tcp4" or "tcp6". A literal address or
// host name for IPv6 must be enclosed in square brackets, as in "[::1]:80",
// "[ipv6-host]:http" or "[ipv6-host%zone]:80".
func (inst Instance) ResolveTCPAddr(network, addr string) (*net.TCPAddr, error) {
	na, err := inst.ResolveNetAddress(network, addr)
	if err != nil {
		return nil, err
	}
	defer na.Release()
	return na.TCPAddr()
}

// ResolveUDPAddr parses addr as a UDP address of the form "host:port" or
// "[ipv6-host%zone]:port" and resolves a pair of domain name and port name on
// the network net, which must be "udp", "udp4" or "udp6". A literal address or
// host name for IPv6 must be enclosed in square brackets, as in "[::1]:80",
// "[ipv6-host]:http" or "[ipv6-host%zone]:80".
func (inst Instance) ResolveUDPAddr(network, addr string) (*net.UDPAddr, error) {
	na, err := inst.ResolveNetAddress(network, addr)
	if err != nil {
		return nil, err
	}
	defer na.Release()
	return na.UDPAddr()
}

// NetAddress represents a network address.
type NetAddress struct {
	Resource
}

// createTCPAddress creates an address from a net.TCPAddr.
func (inst Instance) CreateTCPNetAddress(net string, addr *net.TCPAddr) (na NetAddress, err error) {
	if net == "tcp" || net == "tcp4" {
		if ipv4 := addr.IP.To4(); ipv4 != nil {
			var ppNetAddress pp_NetAddress_IPv4
			binary.BigEndian.PutUint16(ppNetAddress[0:2], uint16(addr.Port))
			copy(ppNetAddress[2:6], ipv4)
			na.id = ppb_netaddress_create_from_ipv4_address(inst.id, &ppNetAddress)
			if na.id == 0 {
				err = errCreateAddressFailed
			}
			return
		}
	}

	if net == "tcp" || net == "tcp6" {
		if ipv6 := addr.IP.To16(); ipv6 != nil {
			var ppNetAddress pp_NetAddress_IPv6
			binary.BigEndian.PutUint16(ppNetAddress[0:2], uint16(addr.Port))
			copy(ppNetAddress[2:18], ipv6)
			na.id = ppb_netaddress_create_from_ipv6_address(inst.id, &ppNetAddress)
			if na.id == 0 {
				err = errCreateAddressFailed
			}
			return
		}
	}

	err = errMalformedAddress
	return
}

// decodeTCPAddress returns the TCP address.
func (na NetAddress) TCPAddr() (addr *net.TCPAddr, err error) {
	family := ppb_netaddress_get_family(na.id)
	switch family {
	case PP_NETADDRESS_FAMILY_UNSPECIFIED:
		err = errMalformedAddress
	case PP_NETADDRESS_FAMILY_IPV4:
		var ipv4 pp_NetAddress_IPv4
		if ok := ppb_netaddress_describe_as_ipv4_address(na.id, &ipv4); ok == ppFalse {
			err = errMalformedAddress
			return
		}
		addr = &net.TCPAddr{Port: int(binary.BigEndian.Uint16(ipv4[0:2])), IP: make([]byte, 4)}
		copy(addr.IP, ipv4[2:6])
	case PP_NETADDRESS_FAMILY_IPV6:
		var ipv6 pp_NetAddress_IPv6
		if ok := ppb_netaddress_describe_as_ipv6_address(na.id, &ipv6); ok == ppFalse {
			err = errMalformedAddress
			return
		}
		addr = &net.TCPAddr{Port: int(binary.BigEndian.Uint16(ipv6[0:2])), IP: make([]byte, 16)}
		copy(addr.IP, ipv6[2:18])
	}
	return
}

// createUDPAddress creates an address from a net.UDPAddr.
func (inst Instance) CreateUDPNetAddress(net string, addr *net.UDPAddr) (na NetAddress, err error) {
	if net == "udp" || net == "udp4" {
		if ipv4 := addr.IP.To4(); ipv4 != nil {
			var ppNetAddress pp_NetAddress_IPv4
			binary.BigEndian.PutUint16(ppNetAddress[0:2], uint16(addr.Port))
			copy(ppNetAddress[2:6], ipv4)
			na.id = ppb_netaddress_create_from_ipv4_address(inst.id, &ppNetAddress)
			if na.id == 0 {
				err = errCreateAddressFailed
			}
			return
		}
	}

	if net == "udp" || net == "udp6" {
		if ipv6 := addr.IP.To16(); ipv6 != nil {
			var ppNetAddress pp_NetAddress_IPv6
			binary.BigEndian.PutUint16(ppNetAddress[0:2], uint16(addr.Port))
			copy(ppNetAddress[2:18], ipv6)
			na.id = ppb_netaddress_create_from_ipv6_address(inst.id, &ppNetAddress)
			if na.id == 0 {
				err = errCreateAddressFailed
			}
			return
		}
	}

	err = errMalformedAddress
	return
}

// decodeUDPAddress returns the UDP address.
func (na NetAddress) UDPAddr() (addr *net.UDPAddr, err error) {
	family := ppb_netaddress_get_family(na.id)
	switch family {
	case PP_NETADDRESS_FAMILY_UNSPECIFIED:
		err = errMalformedAddress
	case PP_NETADDRESS_FAMILY_IPV4:
		var ipv4 pp_NetAddress_IPv4
		if ok := ppb_netaddress_describe_as_ipv4_address(na.id, &ipv4); ok == ppFalse {
			err = errMalformedAddress
			return
		}
		addr = &net.UDPAddr{Port: int(binary.BigEndian.Uint16(ipv4[0:2])), IP: make([]byte, 4)}
		copy(addr.IP, ipv4[2:6])
	case PP_NETADDRESS_FAMILY_IPV6:
		var ipv6 pp_NetAddress_IPv6
		if ok := ppb_netaddress_describe_as_ipv6_address(na.id, &ipv6); ok == ppFalse {
			err = errMalformedAddress
			return
		}
		addr = &net.UDPAddr{Port: int(binary.BigEndian.Uint16(ipv6[0:2])), IP: make([]byte, 16)}
		copy(addr.IP, ipv6[2:18])
	}
	return
}

// tcpSocket implements functions common to both dialed and lister sockets.
type tcpSocket struct {
	Resource
	closed bool
}

// CreateTCPConn creates a fresh, unconnected socket.
//
// Permissions: Apps permission socket with subrule tcp-connect is required for
// Connect(); subrule tcp-listen is required for Listen(). For more details
// about network communication permissions, please see:
// http://developer.chrome.com/apps/app_network.html
func (inst Instance) createTCPConn() (s tcpSocket, err error) {
	id := ppb_tcpsocket_create(inst.id)
	if id == 0 {
		err = errCreateTCPSocketFailed
		return
	}
	s.id = id
	return
}

// Bind the socket to an address.
func (sock *tcpSocket) bind(addr *NetAddress) error {
	code := ppb_tcpsocket_bind(sock.id, addr.id, ppNullCompletionCallback)
	return decodeError(Error(code))
}

// Accept a connection.
func (sock *tcpSocket) accept() (accepted *TCPConn, err error) {
	accepted = &TCPConn{}
	code := ppb_tcpsocket_accept(sock.id, &accepted.id, ppNullCompletionCallback)
	if Error(code) == PP_ERROR_ABORTED {
		err = io.EOF
	} else if code < 0 {
		err = decodeError(Error(code))
	}
	return
}

// Connects the socket to the given address.
//
// The socket must not be listening. Binding the socket beforehand is optional.
func (sock *tcpSocket) connect(addr *NetAddress) error {
	code := ppb_tcpsocket_connect(sock.id, addr.id, ppNullCompletionCallback)
	return decodeError(Error(code))
}

// Starts listening.
//
// The socket must be bound and not connected.
func (sock *tcpSocket) listen(backlog int) error {
	code := ppb_tcpsocket_listen(sock.id, int32(backlog), ppNullCompletionCallback)
	return decodeError(Error(code))
}

// localAddr returns the local address of the socket, if it is bound.
func (sock *tcpSocket) localAddr() net.Addr {
	var local NetAddress
	local.id = ppb_tcpsocket_get_local_address(sock.id)
	if local.id == 0 {
		return nil
	}
	defer local.Release()

	tcpAddr, err := local.TCPAddr()
	if err != nil {
		return nil
	}
	return tcpAddr
}

// remoteAddr returns the remote address of the socket, if it is connected.
func (sock *tcpSocket) remoteAddr() net.Addr {
	var local NetAddress
	local.id = ppb_tcpsocket_get_remote_address(sock.id)
	if local.id == 0 {
		return nil
	}
	defer local.Release()

	tcpAddr, err := local.TCPAddr()
	if err != nil {
		return nil
	}
	return tcpAddr
}

// TCPConn is a TCP socket.
type TCPConn struct {
	tcpSocket
}

// DialTCP connects to the remote address raddr on the network net, which must be
// "tcp", "tcp4", or "tcp6". If laddr is not nil, it is used as the local
// address for the connection.
func (inst Instance) DialTCP(net string, laddr, raddr *net.TCPAddr) (conn *TCPConn, err error) {
	conn = &TCPConn{}
	conn.tcpSocket, err = inst.createTCPConn()
	if err != nil {
		return
	}
	if laddr != nil {
		var addr NetAddress
		addr, err = inst.CreateTCPNetAddress(net, laddr)
		if err != nil {
			conn.Release()
			return
		}
		defer addr.Release()
		if err = conn.bind(&addr); err != nil {
			conn.Release()
			return
		}
	}
	var addr NetAddress
	addr, err = inst.CreateTCPNetAddress(net, raddr)
	if err != nil {
		conn.Release()
		return
	}
	defer addr.Release()
	if err = conn.connect(&addr); err != nil {
		conn.Release()
		return
	}
	return
}

// Close closes the connection.
//
// Any pending callbacks will still run, reporting PP_ERROR_ABORTED if pending
// IO was interrupted. After a call to this method, no output buffer pointers
// passed into previous Read() or Accept() calls will be accessed. It is not
// valid to call Connect() or Listen() again.
func (conn *TCPConn) Close() error {
	ppb_tcpsocket_close(conn.id)
	conn.Release()
	conn.closed = true
	return nil
}

// localAddr returns the local address of the socket, if it is bound.
func (conn *TCPConn) LocalAddr() net.Addr {
	return conn.localAddr()
}

// RemoteAddr returns the remote address of the socket, if it is connected.
func (conn *TCPConn) RemoteAddr() net.Addr {
	return conn.remoteAddr()
}

// Read reads data from the connection.  Deadlines are not supported.
func (conn *TCPConn) Read(buf []byte) (n int, err error) {
	if conn.closed {
		return 0, fmt.Errorf("Reading from closed connection")
	}
	code := ppb_tcpsocket_read(conn.id, &buf[0], int32(len(buf)), ppNullCompletionCallback)
	if code < 0 {
		err = decodeError(Error(code))
		return
	}
	if code == 0 {
		err = io.EOF
		return
	}
	n = int(code)
	return
}

// Write writes data to the connection.  Deadlines are not supported.
func (conn *TCPConn) Write(buf []byte) (n int, err error) {
	if conn.closed {
		return 0, fmt.Errorf("Writing to closed connection")
	}
	for len(buf) != 0 {
		code := ppb_tcpsocket_write(conn.id, &buf[0], int32(len(buf)), ppNullCompletionCallback)
		if code < 0 {
			err = decodeError(Error(code))
			return
		}
		if code == 0 {
			err = io.EOF
			return
		}
		amount := int(code)
		n += amount
		buf = buf[amount:]
	}
	return
}

// SetReadBuffer sets the size of the operating system's receive buffer
// associated with the connection.
func (conn *TCPConn) SetReadBuffer(bytes int) error {
	v := VarFromInt(int32(bytes))
	code := ppb_tcpsocket_set_option(conn.id, PP_TCPSOCKET_OPTION_RECV_BUFFER_SIZE, v.toPPVar(), ppNullCompletionCallback)
	return decodeError(Error(code))
}

// SetWriteBuffer sets the size of the operating system's receive buffer
// associated with the connection.
func (conn *TCPConn) SetWriteBuffer(bytes int) error {
	v := VarFromInt(int32(bytes))
	code := ppb_tcpsocket_set_option(conn.id, PP_TCPSOCKET_OPTION_SEND_BUFFER_SIZE, v.toPPVar(), ppNullCompletionCallback)
	return decodeError(Error(code))
}

// SetDeadline sets the read and write deadlines associated
// with the connection.  Not supported.
func (conn *TCPConn) SetDeadline(t time.Time) error {
	return errDeadlineNotSupported
}

// SetReadDeadline sets the deadline for future Read calls.
// Not supported.
func (conn *TCPConn) SetReadDeadline(t time.Time) error {
	return errDeadlineNotSupported
}

// SetWriteDeadline sets the deadline for future Write calls.
// Not supported.
func (conn *TCPConn) SetWriteDeadline(t time.Time) error {
	return errDeadlineNotSupported
}

// TCPListener is a TCP network listener.
type TCPListener struct {
	tcpSocket
}

// ListenTCP announces on the TCP address laddr and returns a TCP listener. Net
// must be "tcp", "tcp4", or "tcp6". If laddr has a port of 0, ListenTCP will
// choose an available port. The caller can use the Addr method of TCPListener
// to retrieve the chosen address.
func (inst Instance) ListenTCP(net string, laddr *net.TCPAddr) (l *TCPListener, err error) {
	l = &TCPListener{}
	l.tcpSocket, err = inst.createTCPConn()
	if err != nil {
		return
	}
	if laddr != nil {
		var addr NetAddress
		addr, err = inst.CreateTCPNetAddress(net, laddr)
		if err != nil {
			l.Release()
			return
		}
		defer addr.Release()
		if err = l.bind(&addr); err != nil {
			l.Release()
			return
		}
	}
	// TODO(jyh): where does the backlog 5 come from?
	if err = l.listen(5); err != nil {
		l.Release()
		return
	}
	return
}

// Accept implements the Accept method in the Listener interface; it waits for
// the next call and returns a generic Conn.
func (l *TCPListener) Accept() (net.Conn, error) {
	return l.AcceptTCP()
}

// AcceptTCP accepts the next incoming call and returns the new connection.
func (l *TCPListener) AcceptTCP() (*TCPConn, error) {
	return l.accept()
}

// Addr returns the listener's network address, a *TCPAddr.
func (l *TCPListener) Addr() net.Addr {
	return l.localAddr()
}

// Close stops listening on the TCP address. Already accepted connections are
// not closed.
func (l *TCPListener) Close() error {
	ppb_tcpsocket_close(l.id)
	return nil
}

// UDPConn is the implementation of the Conn and PacketConn interfaces for UDP network connections.
type UDPConn struct {
	Resource
	inst       Instance
	net        string
	remoteAddr NetAddress
}

// Release the resources associated with the UDPConn.
func (conn *UDPConn) Release() {
	conn.remoteAddr.Release()
	conn.Resource.Release()
}

// DialUDP connects to the remote address raddr on the network net, which must
// be "udp", "udp4", or "udp6". If laddr is not nil, it is used as the local
// address for the connection.
func (inst Instance) DialUDP(net string, laddr, raddr *net.UDPAddr) (conn *UDPConn, err error) {
	conn, err = inst.ListenUDP(net, laddr)
	if err != nil {
		return
	}
	conn.remoteAddr, err = inst.CreateUDPNetAddress(net, raddr)
	if err != nil {
		conn.Release()
		return
	}
	return
}

// ListenUDP listens for incoming UDP packets addressed to the local address
// laddr. Net must be "udp", "udp4", or "udp6". If laddr has a port of 0,
// ListenUDP will choose an available port. The LocalAddr method of the returned
// UDPConn can be used to discover the port. The returned connection's ReadFrom
// and WriteTo methods can be used to receive and send UDP packets with
// per-packet addressing.
func (inst Instance) ListenUDP(net string, laddr *net.UDPAddr) (conn *UDPConn, err error) {
	conn = &UDPConn{}
	conn.inst = inst
	conn.id = ppb_udpsocket_create(inst.id)
	if conn.id == 0 {
		err = errCreateUDPSocketFailed
		return
	}
	if laddr != nil {
		var addr NetAddress
		addr, err = inst.CreateUDPNetAddress(net, laddr)
		if err != nil {
			conn.Release()
			return
		}
		defer addr.Release()
		code := ppb_udpsocket_bind(conn.id, addr.id, ppNullCompletionCallback)
		if code < 0 {
			conn.Release()
			err = decodeError(Error(code))
			return
		}
	}
	conn.net = net
	return
}

// Close closes the connection.
func (conn *UDPConn) Close() error {
	ppb_udpsocket_close(conn.id)
	conn.Release()
	return nil
}

// LocalAddr returns the local network address.
func (conn *UDPConn) LocalAddr() net.Addr {
	var na NetAddress
	na.id = ppb_udpsocket_get_bound_address(conn.id)
	if na.id == 0 {
		return nil
	}
	defer na.Release()
	addr, _ := na.UDPAddr()
	return addr
}

// RemoteAddr returns the remote network address.
func (conn *UDPConn) RemoteAddr() net.Addr {
	addr, _ := conn.remoteAddr.UDPAddr()
	return addr
}

// Read implements the net.Conn Read method.
func (conn *UDPConn) Read(buf []byte) (n int, err error) {
	n, _, err = conn.ReadFrom(buf)
	return
}

// ReadFrom implements the net.PacketConn ReadFrom method.
func (conn *UDPConn) ReadFrom(buf []byte) (n int, addr net.Addr, err error) {
	n, addr, err = conn.ReadFromUDP(buf)
	return
}

// ReadFromUDP reads a UDP packet from conn, copying the payload into buf. It
// returns the number of bytes copied into buf and the return address that was
// on the packet.
func (conn *UDPConn) ReadFromUDP(buf []byte) (n int, addr *net.UDPAddr, err error) {
	var na NetAddress
	defer na.Release()
	code := ppb_udpsocket_recvfrom(conn.id, &buf[0], int32(len(buf)), &na.id, ppNullCompletionCallback)
	if Error(code) == PP_ERROR_ABORTED {
		err = io.EOF
	} else if code < 0 {
		err = decodeError(Error(code))
		return
	}
	n = int(code)
	addr, _ = na.UDPAddr()
	return
}

// Write implements the Conn Write method.
func (conn *UDPConn) Write(buf []byte) (int, error) {
	return conn.WriteToAddress(buf, conn.remoteAddr)
}

// WriteTo implements the PacketConn WriteTo method.
func (conn *UDPConn) WriteTo(buf []byte, addr net.Addr) (int, error) {
	na, err := conn.inst.ResolveNetAddress(addr.Network(), addr.String())
	if err != nil {
		return 0, err
	}
	defer na.Release()
	return conn.WriteToAddress(buf, na)
}

// WriteToUDP writes a UDP packet to addr via conn, copying the payload from buf.
func (conn *UDPConn) WriteToUDP(buf []byte, addr *net.UDPAddr) (int, error) {
	na, err := conn.inst.CreateUDPNetAddress(conn.net, addr)
	if err != nil {
		return 0, err
	}
	defer na.Release()
	return conn.WriteToAddress(buf, na)
}

// WriteToAddress writes a UDP packet to addr via conn, copying the payload from buf.
func (conn *UDPConn) WriteToAddress(buf []byte, na NetAddress) (n int, err error) {
	code := ppb_udpsocket_sendto(conn.id, &buf[0], int32(len(buf)), na.id, ppNullCompletionCallback)
	if code < 0 {
		err = decodeError(Error(code))
		return
	}
	n = int(code)
	return
}

// SetDeadline sets the read and write deadlines associated
// with the connection.  Not supported.
func (conn *UDPConn) SetDeadline(t time.Time) error {
	return errDeadlineNotSupported
}

// SetReadDeadline sets the deadline for future Read calls.
// Not supported.
func (conn *UDPConn) SetReadDeadline(t time.Time) error {
	return errDeadlineNotSupported
}

// SetWriteDeadline sets the deadline for future Write calls.
// Not supported.
func (conn *UDPConn) SetWriteDeadline(t time.Time) error {
	return errDeadlineNotSupported
}

// Dial connects to the address on the named network.
//
// Known networks are "tcp", "tcp4" (IPv4-only), "tcp6" (IPv6-only), "udp",
// "udp4" (IPv4-only), "udp6" (IPv6-only).
//
// For TCP and UDP networks, addresses have the form host:port. If host is a
// literal IPv6 address or host name, it must be enclosed in square brackets as
// in "[::1]:80" or "[ipv6-host%zone]:80".
func (inst Instance) Dial(network, address string) (net.Conn, error) {
	switch network {
	case "tcp", "tcp4", "tcp6":
		raddr, err := inst.ResolveTCPAddr(network, address)
		if err != nil {
			return nil, err
		}
		return inst.DialTCP(network, nil, raddr)
	case "udp", "udp4", "udp6":
		raddr, err := inst.ResolveUDPAddr(network, address)
		if err != nil {
			return nil, err
		}
		return inst.DialUDP(network, &net.UDPAddr{IP: net.IP{0, 0, 0, 0}}, raddr)
	default:
		return nil, fmt.Errorf("unsupported network: %q", network)
	}
}

// Listen announces on the local network address laddr. The network net must be
// a stream-oriented network: "tcp", "tcp4", "tcp6". See Dial for the syntax of
// laddr.
func (inst Instance) Listen(network, address string) (net.Listener, error) {
	switch network {
	case "tcp", "tcp4", "tcp6":
		addr, err := inst.ResolveTCPAddr(network, address)
		if err != nil {
			return nil, err
		}
		return inst.ListenTCP(network, addr)
	default:
		return nil, fmt.Errorf("unsupported network: %q", network)
	}
}
