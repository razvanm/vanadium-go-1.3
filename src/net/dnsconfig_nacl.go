// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux nacl netbsd openbsd

// Read system DNS config from /etc/resolv.conf

package net

type dnsConfig struct {
	servers  []string // servers to use
	search   []string // suffixes to append to local name
	ndots    int      // number of dots in name to trigger absolute lookup
	timeout  int      // seconds before giving up on packet
	attempts int      // lost packets before giving up on server
	rotate   bool     // round robin among servers
}

// See resolv.conf(5) on a Linux machine.
// TODO(rsc): Supposed to call uname() and chop the beginning
// of the host name to get the default search domain.
// We assume it's in resolv.conf anyway.
func dnsReadConfig() (*dnsConfig, error) {
	conf := new(dnsConfig)
	conf.servers = make([]string, 1)
	conf.search = make([]string, 0)
	conf.ndots = 1
	conf.timeout = 5
	conf.attempts = 2
	conf.rotate = false

	// add a standard dns server
	name := "4.2.2.1"
	switch len(ParseIP(name)) {
	case 16:
		name = "[" + name + "]"
		fallthrough
	case 4:
		conf.servers[0] = name
	}

	return conf, nil
}
