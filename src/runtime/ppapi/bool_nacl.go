// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppapi

const (
	ppFalse pp_Bool = 0
	ppTrue  pp_Bool = 1
)

func fromPPBool(b pp_Bool) bool {
	return b != ppFalse
}

func toPPBool(b bool) pp_Bool {
	if b {
		return ppTrue
	}
	return ppFalse
}
