// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppapi

import (
	"errors"
)

var (
	errViewGetRectFailed     = errors.New("View.GetRect operation failed")
	errViewGetClipRectFailed = errors.New("View.GetClipRect operation failed")
)

// View represents the state of the view of an instance.
//
// You receive new view information through the Instance.DidChangeView callback.
type View struct {
	Resource
}

func makeView(id pp_Resource) (view View) {
	view.id = id
	return
}

// IsFullscreen returns true iff the view is currently full screen.
func (v View) IsFullscreen() bool {
	ok := ppb_view_is_fullscreen(v.id)
	return ok != ppFalse
}

// IsVisible determines whether the module instance might be visible to the
// user.
//
// For example, the Chrome window could be minimized or another window could be
// over it. In both of these cases, the module instance would not be visible to
// the user, but IsVisible() will return true.
//
// Use the result to speed up or stop updates for invisible module instances.
//
// This function performs the duties of GetRect() (determining whether the
// module instance is scrolled into view and the clip rectangle is nonempty) and
// IsPageVisible() (whether the page is visible to the user).
func (v View) IsVisible() bool {
	ok := ppb_view_is_visible(v.id)
	return ok != ppFalse
}

// IsPageVisible determines if the page that contains the module instance is
// visible.
//
// The most common cause of invisible pages is that the page is in a background
// tab in the browser.
//
// Most applications should use IsVisible() instead of this function since the
// module instance could be scrolled off of a visible page, and this function
// will still return true. However, depending on how your module interacts with
// the page, there may be certain updates that you may want to perform when the
// page is visible even if your specific module instance is not visible.
func (v View) IsPageVisible() bool {
	ok := ppb_view_is_page_visible(v.id)
	return ok != ppFalse
}

// GetRect retrieves the rectangle of the module instance associated with a view
// changed notification relative to the upper-left of the browser viewport.
//
// This position changes when the page is scrolled.
//
// The returned rectangle may not be inside the visible portion of the viewport
// if the module instance is scrolled off the page. Therefore, the position may
// be negative or larger than the size of the page. The size will always reflect
// the size of the module were it to be scrolled entirely into view.
//
// In general, most modules will not need to worry about the position of the
// module instance in the viewport, and only need to use the size.
func (v View) GetRect() (rect Rect, err error) {
	ok := ppb_view_get_rect(v.id, &rect)
	if ok != ppTrue {
		err = errViewGetRectFailed
	}
	return
}

// GetClipRect returns the clip rectangle relative to the upper-left corner of
// the module instance.
//
// This rectangle indicates the portions of the module instance that are
// scrolled into view.
//
// If the module instance is scrolled off the view, the return value will be (0,
// 0, 0, 0). This clip rectangle does not take into account page
// visibility. Therefore, if the module instance is scrolled into view, but the
// page itself is on a tab that is not visible, the return rectangle will
// contain the visible rectangle as though the page were visible. Refer to
// IsPageVisible() and IsVisible() if you want to account for page visibility.
//
// Most applications will not need to worry about the clip rectangle. The
// recommended behavior is to do full updates if the module instance is visible,
// as determined by IsVisible(), and do no updates if it is not visible.
//
// However, if the cost for computing pixels is very high for your application,
// or the pages you're targeting frequently have very large module instances
// with small visible portions, you may wish to optimize further. In this case,
// the clip rectangle will tell you which parts of the module to update.
//
// Note that painting of the page and sending of view changed updates happens
// asynchronously. This means when the user scrolls, for example, it is likely
// that the previous backing store of the module instance will be used for the
// first paint, and will be updated later when your application generates new
// content with the new clip. This may cause flickering at the boundaries when
// scrolling. If you do choose to do partial updates, you may want to think
// about what color the invisible portions of your backing store contain (be it
// transparent or some background color) or to paint a certain region outside
// the clip to reduce the visual distraction when this happens.
func (v View) GetClipRect() (rect Rect, err error) {
	ok := ppb_view_get_clip_rect(v.id, &rect)
	if ok != ppTrue {
		err = errViewGetClipRectFailed
	}
	return
}

// GetCSSScale returns the scale factor between DIPs and CSS pixels.
//
// This allows proper scaling between DIPs - as sent via the Pepper API - and
// CSS pixel coordinates used for Web content.
func (v View) GetCSSScale() float32 {
	return ppb_view_get_css_scale(v.id)
}

// GetDeviceScale returns the scale factor between device pixels and Density
// Independent Pixels (DIPs, also known as logical pixels or UI pixels on some
// platforms).
//
// This allows the developer to render their contents at device resolution, even
// as coordinates / sizes are given in DIPs through the API.
//
// Note that the coordinate system for Pepper APIs is DIPs. Also note that one
// DIP might not equal one CSS pixel - when page scale/zoom is in effect.
func (v View) GetDeviceScale() float32 {
	return ppb_view_get_device_scale(v.id)
}
