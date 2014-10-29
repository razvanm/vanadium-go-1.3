// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppapi

// Instance represents one instance of the module on a web page.
// This corresponds to one <embed ...> occurrence.
type Instance struct {
	id pp_Instance
}

func makeInstance(id pp_Instance) (inst Instance) {
	inst.id = id
	return
}

// Log writes a message to the console.
func (inst Instance) Log(level LogLevel, v Var) {
	ppb_console_log(inst.id, level, v.toPPVar())
}

// Log writes a message to the console.
func (inst Instance) LogString(level LogLevel, msg string) {
	v := VarFromString(msg)
	inst.Log(level, v)
	v.Release()
}

// LogWithSource writes a message to the console, using the source information rather than the plugin name.
func (inst Instance) LogWithSource(level LogLevel, src Var, v Var) {
	ppb_console_log_with_source(inst.id, level, src.toPPVar(), v.toPPVar())
}
