// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppapi

import (
	"errors"
)

// Graphics2D represents a 2D graphics context.  The context must be bound to an
// Instance, using BindGraphics2D(), for the drawing to be visible.
type Graphics2D struct {
	Resource
}

var (
	errBindFailed                = errors.New("graphics bind failed")
	errGraphics2DCreateFailed    = errors.New("graphics2D creation failed")
	errGraphics2DContext         = errors.New("graphics2D context error")
	errGraphics2DOperationFailed = errors.New("graphics2D operation failed")
)

// NewGraphics2D returns a new Graphics2D instance.
func (inst Instance) NewGraphics2D(size Size, isAlwaysOpaque bool) (g Graphics2D, err error) {
	rid := ppb_graphics2d_create(inst.id, &size, toPPBool(isAlwaysOpaque))
	if rid == 0 {
		err = errGraphics2DCreateFailed
		return
	}
	g.id = rid
	return
}

// BindGraphics2D binds the Graphics2D context to the Instance.
func (inst Instance) BindGraphics2D(g Graphics2D) error {
	b := ppb_instance_bind_graphics(inst.id, g.id)
	if b == 0 {
		return errBindFailed
	}
	return nil
}

// Describe returns the size and opacity of the graphics 2D context.
func (g Graphics2D) Describe() (size Size, isAlwaysOpaque bool, err error) {
	var b pp_Bool
	ok := ppb_graphics2d_describe(g.id, &size, &b)
	if ok == 0 {
		err = errGraphics2DContext
		return
	}
	isAlwaysOpaque = b != 0
	err = nil
	return
}

// Scroll enqueues a scroll of the context's backing store. This function has no
// effect until you call Flush(). The data within the provided clipping
// rectangle will be shifted by (dx, dy) pixels.
//
// This function will result in some exposed region which will have undefined
// contents. The module should call PaintImageData() on these exposed regions to
// give the correct contents.
//
// The scroll can be larger than the area of the clipping rectangle, which means
// the current image will be scrolled out of the rectangle. This scenario is not
// an error but will result in a no-op.
func (g Graphics2D) Scroll(clipRect Rect, amount Point) {
	ppb_graphics2d_scroll(g.id, &clipRect, &amount)
}

// GetScale returns the scale factor that will be applied when painting the
// graphics context onto the output device.
func (g Graphics2D) GetScale() float32 {
	return ppb_graphics2d_get_scale(g.id)
}

// SetScale sets the scale factor that will be applied when painting the
// graphics context onto the output device.
//
// Typically, if rendering at device resolution is desired, the context would be
// created with the width and height scaled up by the view's GetDeviceScale and
// SetScale called with a scale of 1.0 / GetDeviceScale(). For example, if the
// view resource passed to DidChangeView has a rectangle of (w=200, h=100) and a
// device scale of 2.0, one would call Create with a size of (w=400, h=200) and
// then call SetScale with 0.5. One would then treat each pixel in the context
// as a single device pixel.
func (g Graphics2D) SetScale(scale float32) error {
	b := ppb_graphics2d_set_scale(g.id, scale)
	if b == ppFalse {
		return errGraphics2DOperationFailed
	}
	return nil
}

// PaintImageData() enqueues a paint of the given image into the context.
//
// This function has no effect until you call Flush() As a result, what counts
// is the contents of the bitmap when you call Flush(), not when you call this
// function.
//
// The provided image will be placed at top_left from the top left of the
// context's internal backing store. Then the pixels contained in src_rect will
// be copied into the backing store. This means that the rectangle being painted
// will be at src_rect offset by top_left.
//
// The src_rect is specified in the coordinate system of the image being
// painted, not the context. For the common case of copying the entire image,
// you may specify an empty src_rect.
//
// The painted area of the source bitmap must fall entirely within the
// context. Attempting to paint outside of the context will result in an
// error. However, the source bitmap may fall outside the context, as long as
// the src_rect subset of it falls entirely within the context.
//
// There are two methods most modules will use for painting. The first method is
// to generate a new ImageData and then paint it. In this case, you'll set the
// location of your painting to top_left and set src_rect to NULL. The second is
// that you're generating small invalid regions out of a larger bitmap
// representing your entire instance. In this case, you would set the location
// of your image to (0,0) and then set src_rect to the pixels you changed.
func (g Graphics2D) PaintImageData(data ImageData, topLeft Point, src *Rect) {
	ppb_graphics2d_paint_image_data(g.id, data.id, &topLeft, src)
}

// ReplaceContents provides a slightly more efficient way to paint the entire
// module's image.
//
// Normally, calling PaintImageData() requires that the browser copy the pixels
// out of the image and into the graphics context's backing store. This function
// replaces the graphics context's backing store with the given image, avoiding
// the copy.
//
// The new image must be the exact same size as this graphics context. If the
// new image uses a different image format than the browser's native bitmap
// format (use PPB_ImageData.GetNativeImageDataFormat() to retrieve the format),
// then a conversion will be done inside the browser which may slow the
// performance a little bit.
//
// Note: The new image will not be painted until you call Flush().
//
// After this call, you should take care to release your references to the
// image. If you paint to the image after ReplaceContents(), there is the
// possibility of significant painting artifacts because the page might use
// partially-rendered data when copying out of the backing store.
//
// In the case of an animation, you will want to allocate a new image for the
// next frame. It is best if you wait until the flush callback has executed
// before allocating this bitmap. This gives the browser the option of caching
// the previous backing store and handing it back to you (assuming the sizes
// match). In the optimal case, this means no bitmaps are allocated during the
// animation, and the backing store and "front buffer" (which the plugin is
// painting into) are just being swapped back and forth.
func (g Graphics2D) ReplaceContents(data ImageData) {
	ppb_graphics2d_replace_contents(g.id, data.id)
}

// Flush flushes any enqueued paint, scroll, and replace commands to the
// backing store.
//
// This function actually executes the updates, and causes a repaint of the
// webpage, assuming this graphics context is bound to a module instance.
//
// Flush() runs in asynchronous mode. Specify a callback function and the
// argument for that callback function. The callback function will be executed
// on the calling thread when the image has been painted to the screen. While
// you are waiting for a flush callback, additional calls to Flush() will fail.
//
// Because the callback is executed (or thread unblocked) only when the
// instance's image is actually on the screen, this function provides a way to
// rate limit animations. By waiting until the image is on the screen before
// painting the next frame, you can ensure you're not flushing 2D graphics
// faster than the screen can be updated.
//
// Unbound contexts If the context is not bound to a module instance, you will
// still get a callback. The callback will execute after Flush() returns to
// avoid reentrancy. The callback will not wait until anything is painted to the
// screen because there will be nothing on the screen. The timing of this
// callback is not guaranteed and may be deprioritized by the browser because it
// is not affecting the user experience.
//
// Off-screen instances If the context is bound to an instance that is currently
// not visible (for example, scrolled out of view) it will behave like the
// "unbound context" case.
//
// Detaching a context If you detach a context from a module instance, any
// pending flush callbacks will be converted into the "unbound context" case.
//
// Released contexts A callback may or may not get called even if you have
// released all of your references to the context. This scenario can occur if
// there are internal references to the context suggesting it has not been
// internally destroyed (for example, if it is still bound to an instance) or
// due to other implementation details. As a result, you should be careful to
// check that flush callbacks are for the context you expect and that you're
// capable of handling callbacks for unreferenced contexts.
//
// Shutdown If a module instance is removed when a flush is pending, the
// callback will not be executed.
func (g Graphics2D) Flush() error {
	code := ppb_graphics2d_flush(g.id, ppNullCompletionCallback)
	return decodeError(Error(code))
}
