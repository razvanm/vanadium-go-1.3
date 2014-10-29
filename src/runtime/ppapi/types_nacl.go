// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppapi

// Point represents an integer coordinate in 2D space.
type Point struct {
	X int32
	Y int32
}

// FloatPoint represents a float32 coordinate in 2D space.
type FloatPoint struct {
	X float32
	Y float32
}

// Size represents a 2D integer size.
type Size struct {
	Width  int32
	Height int32
}

// Rect represents a 2D rectangle with integer coordinates.
type Rect struct {
	Point Point
	Size  Size
}
