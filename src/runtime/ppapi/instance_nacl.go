// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppapi

import (
	"fmt"
	"runtime"
)

// Instance represents one instance of the module on a web page.
// This corresponds to one <embed ...> occurrence.
type Instance struct {
	id pp_Instance
}

func makeInstance(id pp_Instance) (inst Instance) {
	inst.id = id
	return
}

// IsFullFrame returns true if the instance is full-frame.
func (inst Instance) IsFullFrame() bool {
	return ppb_instance_is_full_frame(inst.id) != 0
}

// PostMessage sends a message to the browser.
func (inst Instance) PostMessage(v Var) {
	ppb_messaging_post_message(inst.id, v.toPPVar())
}

// Log writes a message to the console.
func (inst Instance) Log(level LogLevel, v Var) {
	ppb_console_log(inst.id, level, v.toPPVar())
}

// LogWithSource writes a message to the console, using the source information rather than the plugin name.
func (inst Instance) LogWithSource(level LogLevel, src, v Var) {
	ppb_console_log_with_source(inst.id, level, src.toPPVar(), v.toPPVar())
}

func (inst Instance) LogWithSourceString(level LogLevel, src, msg string) {
	v1 := VarFromString(src)
	v2 := VarFromString(msg)
	inst.LogWithSource(level, v1, v2)
	v1.Release()
	v2.Release()
}

// LogString writes a message to the console.
func (inst Instance) LogString(level LogLevel, msg string) {
	_, file, line, _ := runtime.Caller(2)
	loc := fmt.Sprintf("%s:%d", file, line)
	inst.LogWithSourceString(level, loc, msg)
}

// Logf writes a formatted message to the console.
func (inst Instance) Logf(level LogLevel, format string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(2)
	loc := fmt.Sprintf("%s:%d", file, line)
	msg := fmt.Sprintf(format, args...)
	inst.LogWithSourceString(level, loc, msg)
}

// Printf writes a formatted message to the console.
func (inst Instance) Printf(format string, args ...interface{}) {
	inst.Logf(PP_LOGLEVEL_LOG, format, args...)
}

// Warningf writes a formatted message to the console.
func (inst Instance) Warningf(format string, args ...interface{}) {
	inst.Logf(PP_LOGLEVEL_WARNING, format, args...)
}

// Errorf writes a formatted message to the console.
func (inst Instance) Errorf(format string, args ...interface{}) {
	inst.Logf(PP_LOGLEVEL_ERROR, format, args...)
}
