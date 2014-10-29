// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppapi

// InputEvent represents a generic input event, including keyboard events, mouse events,
// or some other kind of event.  Use the Type() method to get the type of event.
type InputEvent struct {
	Resource
}

func makeInputEvent(id pp_Resource) (e InputEvent) {
	e.id = id
	return
}

// Type returns the type of the input event.
func (event InputEvent) Type() InputEventType {
	return ppb_inputevent_get_type(event.id)
}

// GetTimeStamp returns the time that the event was generated.
//
// This will be before the current time since processing and dispatching the
// event has some overhead. Use this value to compare the times the user
// generated two events without being sensitive to variable processing time.
func (event InputEvent) TimeStamp() TimeTicks {
	return TimeTicks(ppb_inputevent_get_time_stamp(event.id))
}

// GetModifiers returns a bitfield indicating which modifiers were down at the
// time of the event.
//
// This is a combination of the flags in the PP_InputEvent_Modifier enum.
func (event InputEvent) Modifiers() uint32 {
	return ppb_inputevent_get_modifiers(event.id)
}

// MouseInputEvent represents all mouse events except mouse wheel events.
type MouseInputEvent struct {
	InputEvent
	Modifier   uint32
	Button     InputEventMouseButton
	Position   Point
	Movement   Point
	ClickCount int
}

// MouseInputEvent converts the generic InputEvent to a MouseInputEvent.
// Unspecified if the event is not a mouse input event.
func (event InputEvent) MouseInputEvent() (e MouseInputEvent) {
	id := event.id
	e.id = id
	e.Modifier = ppb_inputevent_get_modifiers(id)
	e.Button = ppb_mouseinputevent_get_button(id)
	ppb_mouseinputevent_get_position(&e.Position, id)
	ppb_mouseinputevent_get_movement(&e.Movement, id)
	e.ClickCount = int(ppb_mouseinputevent_get_click_count(id))
	return
}

// WheelInputEvent represents all mouse wheel events.
type WheelInputEvent struct {
	InputEvent
	Modifier     uint32
	Delta        FloatPoint
	Ticks        FloatPoint
	ScrollByPage bool
}

// WheelInputEvent converts the generic InputEvent to a WheelInputEvent.
// Unspecified if the event is not a wheel input event.
func (event InputEvent) WheelInputEvent() (e WheelInputEvent) {
	id := event.id
	e.id = id
	e.Modifier = ppb_inputevent_get_modifiers(id)
	ppb_wheelinputevent_get_delta(&e.Delta, id)
	ppb_wheelinputevent_get_ticks(&e.Ticks, id)
	e.ScrollByPage = ppb_wheelinputevent_get_scroll_by_page(id) != ppFalse
	return
}

// KeyInputEvent represents a key up or key down event.
//
// Key up and key down events correspond to physical keys on the keyboard. The
// actual character that the user typed (if any) will be delivered in a
// "character" event.
//
// If the user loses focus on the module while a key is down, a key up event
// might not occur. For example, if the module has focus and the user presses
// and holds the shift key, the module will see a "shift down" message. Then if
// the user clicks elsewhere on the web page, the module's focus will be lost
// and no more input events will be delivered.
//
// If your module depends on receiving key up events, it should also handle
// "lost focus" as the equivalent of "all keys up."
type KeyInputEvent struct {
	InputEvent
	Modifier uint32
	KeyCode  uint32
}

func (event InputEvent) KeyInputEvent() (e KeyInputEvent) {
	id := event.id
	e.id = id
	e.Modifier = ppb_inputevent_get_modifiers(id)
	e.KeyCode = ppb_keyboardinputevent_get_key_code(id)
	return
}

// GetCharacterText returns the typed character as a UTF-8 string for the given
// character event.
func (e KeyInputEvent) GetCharacterText() string {
	var ppVar pp_Var
	ppb_keyboardinputevent_get_character_text(&ppVar, e.id)
	var v Var
	v.fromPP(ppVar)
	s, _ := v.AsString()
	v.Release()
	return s
}

// GetCode returns the DOM |code| field for this keyboard event, as defined in
// the DOM3 Events spec: http://www.w3.org/TR/DOM-Level-3-Events/.
func (e KeyInputEvent) GetCode() string {
	var ppVar pp_Var
	ppb_keyboardinputevent_get_code(&ppVar, e.id)
	var v Var
	v.fromPP(ppVar)
	s, _ := v.AsString()
	v.Release()
	return s
}

// ClearInputEventRequest requests that input events corresponding to the given
// input classes no longer be delivered to the instance.
//
// By default, no input events are delivered. If you have previously requested
// input events via RequestInputEvents() or RequestFilteringInputEvents(), this
// function will unregister handling for the given instance. This will allow
// greater browser performance for those events.
//
// Note that you may still get some input events after clearing the flag if they
// were dispatched before the request was cleared. For example, if there are 3
// mouse move events waiting to be delivered, and you clear the mouse event
// class during the processing of the first one, you'll still receive the next
// two. You just won't get more events generated.
func (inst Instance) ClearInputEventRequest(eventClasses uint32) {
	ppb_inputevent_clear_input_event_request(inst.id, eventClasses)
}

// RequestFilteringInputEvents requests that input events corresponding to the
// given input events are delivered to the instance for filtering.
//
// By default, no input events are delivered. In most cases you would register
// to receive events by calling RequestInputEvents(). In some cases, however,
// you may wish to filter events such that they can be bubbled up to the default
// handlers. In this case, register for those classes of events using this
// function instead of RequestInputEvents().
//
// Filtering input events requires significantly more overhead than just
// delivering them to the instance. As such, you should only request filtering
// in those cases where it's absolutely necessary. The reason is that it
// requires the browser to stop and block for the instance to handle the input
// event, rather than sending the input event asynchronously. This can have
// significant overhead.
func (inst Instance) RequestFilteringInputEvents(eventClasses uint32) (code uint32, err error) {
	c := ppb_inputevent_request_filtering_input_events(inst.id, eventClasses)
	if c < 0 {
		err = decodeError(Error(c))
		return
	}
	code = uint32(c)
	return
}

// RequestInputEvent requests that input events corresponding to the given input
// events are delivered to the instance.
//
// It's recommended that you use RequestFilteringInputEvents() for keyboard
// events instead of this function so that you don't interfere with normal
// browser accelerators.
//
// By default, no input events are delivered. Call this function with the
// classes of events you are interested in to have them be delivered to the
// instance. Calling this function will override any previous setting for each
// specified class of input events (for example, if you previously called
// RequestFilteringInputEvents(), this function will set those events to
// non-filtering mode).
//
// Input events may have high overhead, so you should only request input events
// that your plugin will actually handle. For example, the browser may do
// optimizations for scroll or touch events that can be processed substantially
// faster if it knows there are no non-default receivers for that
// message. Requesting that such messages be delivered, even if they are
// processed very quickly, may have a noticeable effect on the performance of
// the page.
//
// Note that synthetic mouse events will be generated from touch events if (and
// only if) you do not request touch events.
//
// When requesting input events through this function, the events will be
// delivered and not bubbled to the default handlers.
func (inst Instance) RequestInputEvents(eventClasses uint32) (code uint32, err error) {
	c := ppb_inputevent_request_filtering_input_events(inst.id, eventClasses)
	if c < 0 {
		err = decodeError(Error(c))
		return
	}
	code = uint32(c)
	return
}
