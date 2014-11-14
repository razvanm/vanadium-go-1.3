// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Internet protocol family sockets for POSIX

package net

import (
	"syscall"
	"time"
)

func probeIPv4Stack() bool {
	return true
}

func probeIPv6Stack() (supportsIPv6, supportsIPv4map bool) {
	// TODO Make this actually test for IPv6 somehow.
	return true, true
}

func favoriteAddrFamily(net string, laddr, raddr sockaddr, mode string) (family int, ipv6only bool) {
	switch net[len(net)-1] {
	case '4':
		return syscall.AF_INET, false
	case '6':
		return syscall.AF_INET6, true
	}

	if mode == "listen" && (laddr == nil || laddr.isWildcard()) {
		if supportsIPv4map {
			return syscall.AF_INET6, false
		}
		if laddr == nil {
			return syscall.AF_INET, false
		}
		return laddr.family(), false
	}

	if (laddr == nil || laddr.family() == syscall.AF_INET) &&
		(raddr == nil || raddr.family() == syscall.AF_INET) {
		return syscall.AF_INET, false
	}
	return syscall.AF_INET6, false
}

// Internet sockets (TCP, UDP, IP)

func internetSocket(net string, laddr, raddr sockaddr, deadline time.Time, sotype, proto int, mode string) (fd *netFD, err error) {
	family, ipv6only := favoriteAddrFamily(net, laddr, raddr, mode)
	fd, err = socket(net, family, sotype, proto, ipv6only, laddr, raddr, deadline)
	return
}

func ipToSockaddr(family int, ip IP, port int, zone string) (syscall.Sockaddr, error) {
	switch family {
	case syscall.AF_INET:
		if len(ip) == 0 {
			ip = IPv4zero
		}
		if ip = ip.To4(); ip == nil {
			panic("Non IPv4 address")
			return nil, InvalidAddrError("non-IPv4 address")
		}
		sa := new(syscall.SockaddrInet4)
		for i := 0; i < IPv4len; i++ {
			sa.Addr[i] = ip[i]
		}
		sa.Port = port
		return sa, nil
	case syscall.AF_INET6:
		if len(ip) == 0 {
			ip = IPv6zero
		}
		// IPv4 callers use 0.0.0.0 to mean "announce on any available address".
		// In IPv6 mode, Linux treats that as meaning "announce on 0.0.0.0",
		// which it refuses to do.  Rewrite to the IPv6 unspecified address.
		if ip.Equal(IPv4zero) {
			ip = IPv6zero
		}
		if ip = ip.To16(); ip == nil {
			return nil, InvalidAddrError("non-IPv6 address")
		}
		sa := new(syscall.SockaddrInet6)
		for i := 0; i < IPv6len; i++ {
			sa.Addr[i] = ip[i]
		}
		sa.Port = port
		sa.ZoneId = uint32(zoneToInt(zone))
		return sa, nil
	}
	return nil, InvalidAddrError("unexpected socket family")
}
