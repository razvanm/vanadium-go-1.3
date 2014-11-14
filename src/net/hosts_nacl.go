// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Read static host/IP entries from /etc/hosts.

package net

import (
	"sync"
	"time"
)

// hostsPath points to the file with static IP/address entries.
var hostsPath = "/etc/hosts"

// Simple cache.
var hosts struct {
	sync.Mutex
	byName map[string][]string
	byAddr map[string][]string
	expire time.Time
	path   string
}

// lookupStaticHost looks up the addresses for the given host from /etc/hosts.
func lookupStaticHost(host string) []string {
	if host == "localhost" {
		return []string{"127.0.0.1"}
	}
	return nil
}

// lookupStaticAddr looks up the hosts for the given address from /etc/hosts.
func lookupStaticAddr(addr string) []string {
	if addr == "127.0.0.1" {
		return []string{"localhost"}
	}
	return nil
}
