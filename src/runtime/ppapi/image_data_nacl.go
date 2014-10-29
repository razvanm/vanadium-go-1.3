// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppapi

import (
	"errors"
	"unsafe"
)

// ImageData represents a 2D image.
type ImageData struct {
	Resource
}

// ImageDataDesc is a description of an ImageData resource.
type ImageDataDesc struct {
	format ImageDataFormat
	size   Size
	stride int32
}

var (
	errImageDataCreateFailed   = errors.New("ImageData creation failed")
	errImageDataDescribeFailed = errors.New("ImageData.Describe operation failed")
	errImageDataMapFailed      = errors.New("ImageData.Map operation failed")
)

// NewImageData returns a new ImageData object for 2D graphics.
func (inst Instance) NewImageData(fmt ImageDataFormat, size Size, initToZero bool) (id ImageData, err error) {
	rid := ppb_imagedata_create(inst.id, fmt, &size, toPPBool(initToZero))
	if rid == 0 {
		err = errImageDataCreateFailed
		return
	}
	id.id = rid
	return
}

// Describe returns the ImageData description.
func (data ImageData) Describe() (desc ImageDataDesc, err error) {
	ok := ppb_imagedata_describe(data.id, &desc)
	if ok == 0 {
		err = errImageDataDescribeFailed
		return
	}
	return
}

// Map returns a slice referring to the image data.
func (data ImageData) Map() ([]uint8, error) {
	// Get the data size.
	desc, err := data.Describe()
	if err != nil {
		return nil, err
	}
	size := desc.size.Width * desc.size.Height * 4

	// Map the data.
	p := ppb_imagedata_map(data.id)
	if p == nil {
		return nil, errImageDataMapFailed
	}
	b := (*[1 << 24]uint8)(unsafe.Pointer(p))[:size:size]
	return b, nil
}

// Unmap unmaps the image data.
func (data ImageData) Unmap() {
	ppb_imagedata_unmap(data.id)
}
