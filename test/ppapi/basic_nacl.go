// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This is just a crash test.  If the test is successful, it prints a log
// message and quits successfully.
package main

import (
	"runtime/ppapi"
)

type testInstance struct{
	ppapi.Instance
}

func (inst testInstance) DidCreate(args map[string]string) bool {
	inst.Printf("DidCreate: %v", args)
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
