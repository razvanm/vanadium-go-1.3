// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This is just a crash test.  If the test is successful, it prints a log
// message and quits successfully.
package main

import (
	"fmt"
	"runtime/ppapi"
)

type junk struct {
	a int
	b, c interface{}
}

func (j *junk) GoString() string {
	return fmt.Sprintf("junk{a: %d, b: %#v, c:%#v}", j.a, j.b, j.c)
}

func (j *junk) String() string {
	return fmt.Sprintf("{a: %d, b: %v, c:%v}", j.a, j.b, j.c)
}

func make_junk(n int) *junk {
	if n == 0 {
		return nil
	}
	var value junk
	value.a = n
	value.b = make_junk(n - 1)
	return &value
}

func (inst testInstance) run() {
	v := make_junk(4)
	inst.Printf("%#v\n", v)
	inst.Printf("%v\n", v)
}

type testInstance struct{
	ppapi.Instance
}

func (inst testInstance) DidCreate(args map[string]string) bool {
	inst.Printf("xDidCreate: %v", args)
	inst.run()
	return true
}

func (testInstance) DidDestroy() {
}

func (testInstance) DidChangeView(view ppapi.View) {
}

func (testInstance) DidChangeFocus(has_focus bool) {
}

func (testInstance) HandleDocumentLoad(url_loader ppapi.Resource) bool {
	return true
}

func (testInstance) HandleInputEvent(event ppapi.InputEvent) bool {
	return true
}

func (testInstance) Graphics3DContextLost() {
}

func (testInstance) HandleMessage(message ppapi.Var) {
}

func (testInstance) MouseLockLost() {
}

func newTestInstance(inst ppapi.Instance) ppapi.InstanceHandlers {
	return testInstance{Instance: inst}
}

func main() {
	ppapi.Init(newTestInstance)
}
