// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppapi

// Resource represents a generic NaCl reference of some kind.  You won't normally
// use Resource directly, you will use one of the specific resources, like FileIO,
// InputEvent, or others.
//
// The Resource type is defined to support generic operations on all Resources,
// including manipulating the reference counts.
type Resource struct {
	id pp_Resource
}

func makeResource(id pp_Resource) Resource {
	return Resource{id: id}
}

// IsNull returns true iff the resource is NULL.
func (r Resource) IsNull() bool {
	return r.id == 0
}

// IsValid returns true iff the resource is not NULL.
func (r Resource) IsValid() bool {
	return r.id != 0
}

// AddRef increments the Resource's reference count.
func (r Resource) AddRef() {
	ppb_core_add_ref_resource(r.id)
}

// Release decrements the Resource's reference count.  Deletes the Resource if
// the resulting count is zero.
func (r *Resource) Release() {
	if r.id != 0 {
		ppb_core_release_resource(r.id)
		r.id = 0
	}
}
