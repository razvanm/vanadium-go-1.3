// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppapi

import (
	"time"
)

func fromPPTime(t pp_Time) time.Time {
	sec := int64(t)
	nsec := int64(float64(t) - float64(sec))
	return time.Unix(sec, nsec)
}

func toPPTime(t time.Time) pp_Time {
	sec := t.Unix()
	nsec := t.Nanosecond()
	return pp_Time(sec) + pp_Time(nsec)*1000000000
}
