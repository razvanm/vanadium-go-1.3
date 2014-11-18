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

func (inst testInstance) HandleMessage(msg ppapi.Var) {
	switch ty := msg.Type(); ty {
	case ppapi.PP_VARTYPE_INT32:
		if i, err := msg.AsInt(); err != nil {
			inst.Errorf("Bad message: %s", err)
		} else {
			inst.Printf("Int: 0x%x", i)
		}
	case ppapi.PP_VARTYPE_DOUBLE:
		if i, err := msg.AsDouble(); err != nil {
			inst.Errorf("Bad message: %s", err)
		} else {
			inst.Printf("Double: %v", i)
		}
	case ppapi.PP_VARTYPE_STRING:
		if s, err := msg.AsString(); err != nil {
			inst.Errorf("Bad message: %s", err)
		} else {
			inst.Printf("String: %q", s)
		}
	case ppapi.PP_VARTYPE_ARRAY:
		inst.Printf("Array")
	case ppapi.PP_VARTYPE_DICTIONARY:
		keys, err := msg.GetKeys()
		if err != nil {
			inst.Errorf("Bad message: %s", err)
			break
		}
		inst.Printf("Dictionary: keys = %v", keys)
		for _, key := range keys {
			if s, err := msg.LookupStringValuedKey(key); err != nil {
				inst.Errorf("Bad key: %s", err)
			} else {
				inst.Printf("Value: %q: %q", key, s)
			}
		}
	default:
		inst.Errorf("Unknown type: %d", ty)
	}
}

func (testInstance) MouseLockLost() {
}

func newTestInstance(inst ppapi.Instance) ppapi.InstanceHandlers {
	return testInstance{Instance: inst}
}

func main() {
	ppapi.Init(newTestInstance)
}
