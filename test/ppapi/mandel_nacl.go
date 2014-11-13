// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The Mandelbrot test displays a window containing a monochrome view of the
// Mandelbrot set.  Left mouse clicks in the window recenter the window around
// the clicked point and zoom in slightly.
//
// This tests several PPAPI interfaces, including: Graphics2D, ImageData, and
// MouseInputEvent.

package main

import (
	"runtime/ppapi"
	"sync"
)

// Mandelbrot performs the actual Mandelbrot calculation.
type mandel struct {
	size ppapi.Size
	x, y float64
	scale float64
	data []uint16
}

// Instance stores the state of a single PPAPI instance.
// It includes the mandel state, plus a 2D graphics context.
type Instance struct{
	ppapi.Instance

	// Main state.
	mutex sync.Mutex
	mandel mandel
	g2d ppapi.Graphics2D
	imageData ppapi.ImageData
	pixelData []uint8

	// Input processor.
	inputEventQueue chan ppapi.MouseInputEvent
	pending sync.WaitGroup
}

const (
	// maxIteration is the maximum number of iterations in the Mandelbrot
	// point calculation.
	maxIteration = 4096

	// maxValue is the maximum value stored in the data array.
	maxValue uint32 = 1024

	// eventQueueSize is the number of events that can be queued in the
	// inputEvent channel.
	eventQueueSize = 5
)

// reset resets and recenters the image.
func (m *mandel) reset(size ppapi.Size, x, y, scale float64) {
	m.size = size
	m.x = x
	m.y = y
	m.scale = scale
	m.data = make([]uint16, size.Width * size.Height)
}

// recenter recenters the image at a point, zooming in slightly.
func (m *mandel) recenter(p ppapi.Point) {
	xscale := m.scale
	yscale := xscale * float64(m.size.Height) / float64(m.size.Width)
	m.x += float64(p.X) * xscale / float64(m.size.Width) - xscale / 2
	m.y += float64(m.size.Height - p.Y) * yscale / float64(m.size.Height) - yscale / 2
	m.scale *= 0.75
}

// compute recomputes the Mandelbrot data.
func (m *mandel) compute() {
	var histogram [maxIteration]uint32
	xscale := m.scale
	yscale := xscale * float64(m.size.Height) / float64(m.size.Width)
	d := xscale / float64(m.size.Width)
	index := 0
	y := m.y + yscale / 2
	for i := int32(0); i != m.size.Height; i++ {
		x := m.x - xscale / 2
		for j := int32(0); j != m.size.Width; j++ {
			k := m.point(x, y)
			m.data[index] = uint16(k)
			histogram[k]++
			index++
			x += d
		}
		y -= d
	}
	m.adjust(histogram[:])
}

// point calculates a single point in the Mandelbrot set.
func (m *mandel) point(a, b float64) int {
	var x, y float64
	for i := 0; i != maxIteration; i++ {
		xsqr := x * x
		ysqr := y * y
		if xsqr + ysqr >= 4 {
			return i
		}
		y = 2 * x * y + b
		x = xsqr - ysqr + a
	}
	return 0
}

// adjust adjusts the output using the histogram.  This is to normalize the
// colors so they look sensible no matter how much dynamic range we have in the
// calculation.
func (m *mandel) adjust(histogram []uint32) {
	// Calculate the cumulative histogram.
	total := uint64(m.size.Width * m.size.Height)
	var sum uint64
	for i, k := range histogram {
		sum += uint64(k)
		histogram[i] = uint32(sum * uint64(maxValue) / total)
	}
	histogram[maxIteration - 1] = 0

	// Adjust the colors.
	for i, k := range m.data {
		m.data[i] = uint16(histogram[k])
	}
}

// release releases the Graphics2D resources associated with an instance.
func (inst *Instance) release() {
	inst.g2d.Release()
	if inst.imageData.IsValid() {
		inst.imageData.Unmap()
		inst.imageData.Release()
	}
	inst.pixelData = nil
}

// compute recomputes the Mandelbrot state then draws it to the Graphics2D context.
func (inst *Instance) compute() {
	inst.mandel.compute()

	// Copy onto pixelData.
	size := inst.mandel.size.Width * inst.mandel.size.Height
	for i := int32(0); i < size; i++ {
		color := uint8(uint32(inst.mandel.data[i]) * 255 / maxValue)
		j := i * 4
		inst.pixelData[j] = color
		inst.pixelData[j + 1] = color
		inst.pixelData[j + 2] = color
		inst.pixelData[j + 3] = 255
	}
}

// flush paints the Graphics2D data to the screen.
func (inst *Instance) flush() {
	inst.g2d.PaintImageData(inst.imageData, ppapi.Point{}, nil)
	if err := inst.g2d.Flush(); err != nil {
		inst.Errorf("Flush failed: %s", err)
	}
}

func (inst *Instance) refresh(rect ppapi.Rect) {
	inst.Printf("refresh: size=%v", rect.Size)
	inst.mutex.Lock()
	defer inst.mutex.Unlock()

	if rect.Size.Width == inst.mandel.size.Width && rect.Size.Height == inst.mandel.size.Height {
		// Nothing changed
		return
	}

	// Release the old Graphics2D context.
	inst.release()

	// Create the Graphics2D context.
	g2d, err := inst.NewGraphics2D(rect.Size, false)
	if err != nil {
		inst.Printf("Failed to create 2D graphics: %s", err)
		return
	}
	if err := inst.BindGraphics2D(g2d); err != nil {
		inst.Printf("Failed to bind 2D graphics: %s", err)
		g2d.Release()
		return
	}
	inst.g2d = g2d
	inst.imageData, err = inst.NewImageData(ppapi.PP_IMAGEDATAFORMAT_BGRA_PREMUL, rect.Size, true)
	if err != nil {
		inst.Printf("Failed to create ImageData: %s", err)
		return
	}
	inst.pixelData, err = inst.imageData.Map()
	if err != nil {
		inst.Printf("Failed to map ImageData: %s", err)
		return
	}

	// Reset the view.
	inst.mandel.reset(rect.Size, 0, 0, 4)

	// Draw the image.
	inst.compute()
	inst.flush()

	inst.Printf("refresh: done")
}

func (inst *Instance) inputEventLoop() {
	for e := range inst.inputEventQueue {
		inst.handleInputEvent(e)
	}
	inst.pending.Done()
}

// recenter the image around the current mouse click position and zoom in a
// little.
func (inst *Instance) handleInputEvent(e ppapi.MouseInputEvent) {
	inst.Printf("handleInputEvent: %v", e)
	inst.mutex.Lock()
	defer inst.mutex.Unlock()
	inst.Printf("MouseUpEvent: %v", e)
	inst.mandel.recenter(e.Position)
	inst.compute()
	inst.flush()
	inst.Printf("MouseUpEvent: done")
}

// DidCreate is called when the instance is created.
func (inst *Instance) DidCreate(args map[string]string) bool {
	inst.Printf("DidCreate: %v", args)
	inst.inputEventQueue = make(chan ppapi.MouseInputEvent, eventQueueSize)
	inst.pending.Add(1)
	go inst.inputEventLoop()
	inst.RequestInputEvents(uint32(ppapi.PP_INPUTEVENT_TYPE_MOUSEUP))
	inst.Printf("DidCreate: done")
	return true
}

func (inst *Instance) DidDestroy() {
	inst.Printf("DidDestroy")
	if inst.inputEventQueue == nil {
		return
	}
	close(inst.inputEventQueue)
	inst.pending.Wait()
}

func (inst *Instance) DidChangeView(view ppapi.View) {
	inst.Printf("DidChangeView")
	r, err := view.GetRect()
	if err != nil {
		inst.Errorf("Can't get view rectangle: %s", err)
		return
	}
	go inst.refresh(r)
	inst.Printf("DidChangeView: done")
}

func (*Instance) DidChangeFocus(has_focus bool) {
}

func (*Instance) HandleDocumentLoad(url_loader ppapi.Resource) bool {
	return true
}

func (inst *Instance) HandleInputEvent(event ppapi.InputEvent) bool {
	switch event.Type() {
	case ppapi.PP_INPUTEVENT_TYPE_MOUSEUP:
		e := event.MouseInputEvent()
		inst.Printf("HandleInputEvent: %v", e)
		inst.inputEventQueue <- e
		return true
	}
	return false
}

func (*Instance) Graphics3DContextLost() {
}

func (*Instance) HandleMessage(message ppapi.Var) {
}

func (*Instance) MouseLockLost() {
}

func newTestInstance(inst ppapi.Instance) ppapi.InstanceHandlers {
	return &Instance{Instance: inst}
}

func main() {
	ppapi.Init(newTestInstance)
}
