// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io"
	"net"
	"runtime"
	"runtime/ppapi"
)

type testInstance struct{
	ppapi.Instance
	errors int
}

// Errorf writes an error message to the console and increments the error count.
func (inst *testInstance) Errorf(format string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	loc := ppapi.VarFromString(fmt.Sprintf("%s:%d", file, line))
	v := ppapi.VarFromString(fmt.Sprintf(format, args...))
	inst.LogWithSource(ppapi.PP_LOGLEVEL_ERROR, loc, v)
	loc.Release()
	v.Release()
	inst.errors++
}

// doEcho echoes back whatever was read.
func (inst *testInstance) doEcho(c net.Conn) {
	var buf [100]byte
	for {
		n, err := c.Read(buf[:])
		if err == io.EOF {
			break
		}
		if err != nil {
			inst.Errorf("Read: %s", err)
			break
		}
		m, err := c.Write(buf[0:n])
		if err != nil {
			inst.Errorf("Write: %s", err)
			break
		}
		if m != n {
			inst.Errorf("Short write: %d < %d", m, n)
		}
	}
	c.Close()
}

// tcpEchoServer accepts connections, then starts an echoing goroutine.
func (inst *testInstance) tcpEchoServer(listen net.Listener) {
	for {
		c, err := listen.Accept()
		if err == io.EOF {
			break
		}
		if err != nil {
			inst.Errorf("Accept: %s", err)
			return
		}
		go inst.doEcho(c)
	}
}

// udpEchoServer accepts connections, then starts an echoing goroutine.
func (inst *testInstance) udpEchoServer(conn *ppapi.UDPConn) {
	defer conn.Close()

	var buf [100]byte
	n, addr, err := conn.ReadFrom(buf[:])
	if err == io.EOF {
		return
	}
	if err != nil {
		inst.Errorf("ReadFrom: %s", err)
		return
	}
	m, err := conn.WriteTo(buf[:n], addr)
	if err != nil {
		inst.Errorf("WriteTo: %s", err)
		return
	}
	if m != n {
		inst.Errorf("WriteTo: short write %d < %d", m, n)
		return
	}
}

// TestTCPEcho is a simple echo test.  Start an echo server, then one client,
// and compare the echoed text to what was written.
func (inst *testInstance) TestTCPEcho() {
	inst.Printf("TestTCPLoop")

	// Open a server socket.
	listen, err := inst.Listen("tcp", "localhost:0")
	if err != nil {
		inst.Errorf("Can't open TCP connection: %s", err)
		return
	}
	defer listen.Close()
	go inst.tcpEchoServer(listen)

	// Open a client socket.
	saddr := listen.Addr()
	sock, err := inst.Dial(saddr.Network(), saddr.String())
	if err != nil {
		inst.Errorf("Dial: %s, %v: %s", saddr.Network(), saddr, err)
		return
	}
	defer sock.Close()

	// Write a message.
	n, err := sock.Write([]byte("Hello world"))
	if err != nil {
		inst.Errorf("Write: %s", err)
	}
	if n != 11 {
		inst.Errorf("Short write: %d bytes", n)
	}

	// Read it back.
	var buf [100]byte
	n, err = sock.Read(buf[:])
	if err != nil {
		inst.Errorf("Read: %s", err)
	}
	if n != 11 {
		inst.Errorf("Short read: %d bytes", n)
	}
	s := string(buf[:n])
	if s != "Hello world" {
		inst.Errorf("Unexpected read: %q", s)
	}
}

// TestUDPEcho is similar to TestTCPEcho, but it uses UDP.
func (inst *testInstance) TestUDPEcho() {
	inst.Printf("TestUDPEcho")

	// Open a server socket.
	addr, err := inst.ResolveUDPAddr("udp", "localhost:0")
	if err != nil {
		inst.Errorf("ResolveUDPAddr: %s", err)
		return
	}
	conn, err := inst.ListenUDP("udp", addr)
	if err != nil {
		inst.Errorf("Can't open TCP connection: %s", err)
		return
	}
	go inst.udpEchoServer(conn)

	// Open a client socket.
	saddr := conn.LocalAddr().(*net.UDPAddr)
	sock, err := inst.DialUDP(saddr.Network(), addr, saddr)
	if err != nil {
		inst.Errorf("Dial: %s, %v: %s", saddr.Network(), saddr, err)
		return
	}
	defer sock.Close()

	// Write a message.
	n, err := sock.Write([]byte("Hello world"))
	if err != nil {
		inst.Errorf("Write: %s", err)
	}
	if n != 11 {
		inst.Errorf("Short write: %d bytes", n)
	}

	// Read it back.
	var buf [100]byte
	n, err = sock.Read(buf[:])
	if err != nil {
		inst.Errorf("Read: %s", err)
	}
	if n != 11 {
		inst.Errorf("Short read: %d bytes", n)
	}
	s := string(buf[:n])
	if s != "Hello world" {
		inst.Errorf("Unexpected read: %q", s)
	}
}

func (inst *testInstance) RunAllTests() {
	inst.TestTCPEcho()
	inst.TestUDPEcho()
	if inst.errors == 0 {
		inst.Printf("All tests passed")
	} else {
		inst.Errorf("Tests failed with %d errors", inst.errors)
	}
}

func (inst *testInstance) DidCreate(args map[string]string) bool {
	go inst.RunAllTests()
	return true
}

func (*testInstance) DidDestroy() {
}

func (*testInstance) DidChangeView(view ppapi.View) {
}

func (*testInstance) DidChangeFocus(has_focus bool) {
}

func (*testInstance) HandleDocumentLoad(url_loader ppapi.Resource) bool {
	return true
}

func (*testInstance) HandleInputEvent(event ppapi.InputEvent) bool {
	return true
}

func (*testInstance) Graphics3DContextLost() {
}

func (*testInstance) HandleMessage(message ppapi.Var) {
}

func (*testInstance) MouseLockLost() {
}

func newTestInstance(inst ppapi.Instance) ppapi.InstanceHandlers {
	return &testInstance{Instance: inst}
}

func main() {
	ppapi.Init(newTestInstance)
}
