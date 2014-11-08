// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppapi

import (
	"sync"
	"log"
	"unsafe"
)

var (
	ppNullCompletionCallback pp_CompletionCallback

	instanceFactory func(inst Instance) InstanceHandlers

	// Instance table.
	instanceLock sync.Mutex
	instances    = make(map[pp_Instance]InstanceHandlers)
)

// InstanceHandlers contains handlers that you must implement in your application.
type InstanceHandlers interface {
	// DidCreate() is a creation handler that is called when a new instance is created.
	//
	// This function is called for each instantiation on the page, corresponding
	// to one <embed> tag on the page.
	//
	// Generally you would handle this call by initializing the information your
	// module associates with an instance and creating a mapping from the given
	// PP_Instance handle to this data. The PP_Instance handle will be used in
	// subsequent calls to identify which instance the call pertains to.
	//
	// It's possible for more than one instance to be created in a single
	// module. This means that you may get more than one OnCreate without an
	// OnDestroy in between, and should be prepared to maintain multiple states
	// associated with each instance.
	//
	// If this function reports a failure (by returning false), the instance
	// will be deleted.
	DidCreate(args map[string]string) bool

	// DidDestroy is an instance destruction handler.
	//
	// This function is called in many cases (see below) when a module instance
	// is destroyed. It will be called even if DidCreate() returned failure.
	//
	// Generally you will handle this call by deallocating the tracking
	// information and the Instance you created in the DidCreate
	// call. You can also free resources associated with this instance but this
	// isn't required; all resources associated with the deleted instance will
	// be automatically freed when this function returns.
	//
	// The instance identifier will still be valid during this call, so the
	// module can perform cleanup-related tasks. Once this function returns, the
	// Instance handle will be invalid. This means that you can't do any
	// asynchronous operations like network requests, file writes or messaging
	// from this function since they will be immediately canceled.
	//
	// Note: This function will always be skipped on untrusted (Native Client)
	// implementations. This function may be skipped on trusted implementations
	// in certain circumstances when Chrome does "fast shutdown" of a web
	// page. Fast shutdown will happen in some cases when all module instances
	// are being deleted, and no cleanup functions will be called. The module
	// will just be unloaded and the process terminated.
	DidDestroy()

	// DidChangeView is called when the position, size, or other view attributes
	// of the instance has changed.
	DidChangeView(view View)

	// DidChangeFocus is called when an instance has gained or lost focus.
	//
	// Having focus means that keyboard events will be sent to the instance. An
	// instance's default condition is that it will not have focus.
	//
	// The focus flag takes into account both browser tab and window focus as well
	// as focus of the plugin element on the page. In order to be deemed to have
	// focus, the browser window must be topmost, the tab must be selected in the
	// window, and the instance must be the focused element on the page.
	//
	// Note:Clicks on instances will give focus only if you handle the click
	// event. Return true from HandleInputEvent in PPP_InputEvent (or use unfiltered
	// events) to signal that the click event was handled. Otherwise, the browser
	// will bubble the event and give focus to the element on the page that actually
	// did end up consuming it. If you're not getting focus, check to make sure
	// you're either requesting them via RequestInputEvents() (which implicitly
	// marks all input events as consumed) or via RequestFilteringInputEvents() and
	// returning true from your event handler.
	DidChangeFocus(has_focus bool)

	// HandleDocumentLoad is called after initialize for a full-frame instance
	// that was instantiated based on the MIME type of a DOMWindow navigation.
	//
	// This situation only applies to modules that are pre-registered to handle
	// certain MIME types. If you haven't specifically registered to handle a
	// MIME type or aren't positive this applies to you, your implementation of
	// this function can just return PP_FALSE.
	//
	// The given url_loader corresponds to a PPB_URLLoader instance that is
	// already opened. Its response headers may be queried using
	// PPB_URLLoader::GetResponseInfo. The reference count for the URL loader is
	// not incremented automatically on behalf of the module. You need to
	// increment the reference count yourself if you are going to keep a
	// reference to it.
	//
	// This method returns PP_FALSE if the module cannot handle the data. In
	// response to this method, the module should call ReadResponseBody() to
	// read the incoming data.	HandleDocumentLoad(url_loader Resource) bool
	HandleDocumentLoad(url_loader Resource) bool

	// HandleInputEvent is the function for receiving input events from the browser.
	//
	// In order to receive input events, you must register for them by calling
	// PPB_InputEvent.RequestInputEvents() or RequestFilteringInputEvents(). By
	// default, no events are delivered.
	//
	// If the event was handled, it will not be forwarded to the default
	// handlers in the web page. If it was not handled, it may be dispatched to
	// a default handler. So it is important that an instance respond accurately
	// with whether event propagation should continue.
	//
	// Event propagation also controls focus. If you handle an event like a
	// mouse event, typically the instance will be given focus. Returning false
	// from a filtered event handler or not registering for an event type means
	// that the click will be given to a lower part of the page and your
	// instance will not receive focus. This allows an instance to be partially
	// transparent, where clicks on the transparent areas will behave like
	// clicks to the underlying page.
	//
	// In general, you should try to keep input event handling short. Especially
	// for filtered input events, the browser or page may be blocked waiting for
	// you to respond.
	//
	// The caller of this function will maintain a reference to the input event
	// resource during this call. Unless you take a reference to the resource to
	// hold it for later, you don't need to release it.
	//
	// Note: If you're not receiving input events, make sure you register for
	// the event classes you want by calling RequestInputEvents or
	// RequestFilteringInputEvents. If you're still not receiving keyboard input
	// events, make sure you're returning true (or using a non-filtered event
	// handler) for mouse events. Otherwise, the instance will not receive focus
	// and keyboard events will not be sent.
	HandleInputEvent(event InputEvent) bool

	// Graphics3DContextLost is called when the OpenGL ES window is invalidated
	// and needs to be repainted.
	Graphics3DContextLost()

	// HandleMessage is a function that the browser calls when PostMessage() is
	// invoked on the DOM element for the module instance in JavaScript.
	//
	// Note that PostMessage() in the JavaScript interface is asynchronous,
	// meaning JavaScript execution will not be blocked while HandleMessage() is
	// processing the message.
	//
	// When converting JavaScript arrays, any object properties whose name is
	// not an array index are ignored. When passing arrays and objects, the
	// entire reference graph will be converted and transferred. If the
	// reference graph has cycles, the message will not be sent and an error
	// will be logged to the console.
	//
	// The following JavaScript code invokes HandleMessage, passing the module
	// instance on which it was invoked, with message being a string PP_Var
	// containing "Hello world!"
	//
	// Example:
	//
	//  <body>
	//    <object id="plugin"
	//            type="application/x-ppapi-postMessage-example"/>
	//    <script type="text/javascript">
	//      document.getElementById('plugin').postMessage("Hello world!");
	//    </script>
	//  </body>
	HandleMessage(message Var)

	// MouseLockLost is called when the instance loses the mouse lock, such as
	// when the user presses the ESC key.
	MouseLockLost()
}

// createInstanceHandlers creates a new handler for the instance.
func createInstanceHandlers(id pp_Instance) InstanceHandlers {
	if instanceFactory == nil {
		panic("PPAPI is not initialized")
	}
	handler := instanceFactory(makeInstance(id))

	instanceLock.Lock()
	defer instanceLock.Unlock()
	instances[id] = handler
	return handler
}

// getInstanceHandlers returns the handler for an instance.
func getInstanceHandlers(id pp_Instance) InstanceHandlers {
	instanceLock.Lock()
	defer instanceLock.Unlock()
	handler, _ := instances[id]
	return handler
}

// removeInstanceHandlers removes the handler, returning it.
func removeInstanceHandlers(id pp_Instance) InstanceHandlers {
	instanceLock.Lock()
	handler, _ := instances[id]
	delete(instances, id)
	instanceLock.Unlock()
	return handler
}

// Called from C.
func ppp_did_create(id pp_Instance, argc int32, argn, argv *[1 << 16]*byte) pp_Bool {
	inst := createInstanceHandlers(id)
	args := make(map[string]string)
	for i := int32(0); i < argc; i++ {
		log.Printf("AAAAAAAA2a: %d %p %p %p %p", i, argn, argv, argn[i], argv[i])
		log.Printf("AAAAAAAA2b: %v", (*[16]byte)(unsafe.Pointer(argn[i])))
		key := gostring(argn[i])
		value := gostring(argv[i])
		log.Printf("AAAAAAAA6: %p %p %p", args, key, value)
		args[key] = value
	}
	ok := inst.DidCreate(args)
	return toPPBool(ok)
}

// Called from C.
func ppp_did_destroy(id pp_Instance) {
	inst := removeInstanceHandlers(id)
	if inst == nil {
		return
	}
	inst.DidDestroy()
}

// Called from C.
func ppp_did_change_view(id pp_Instance, view pp_Resource) {
	inst := getInstanceHandlers(id)
	if inst == nil {
		return
	}
	inst.DidChangeView(makeView(view))
}

// Called from C.
func ppp_did_change_focus(id pp_Instance, hasFocus pp_Bool) {
	inst := getInstanceHandlers(id)
	if inst == nil {
		return
	}
	inst.DidChangeFocus(fromPPBool(hasFocus))
}

// Called from C.
func ppp_handle_document_load(id pp_Instance, urlLoader pp_Resource) pp_Bool {
	inst := getInstanceHandlers(id)
	if inst == nil {
		return ppFalse
	}
	ok := inst.HandleDocumentLoad(makeResource(urlLoader))
	return toPPBool(ok)
}

// Called from C.
func ppp_handle_input_event(id pp_Instance, event pp_Resource) pp_Bool {
	inst := getInstanceHandlers(id)
	if inst == nil {
		return ppFalse
	}
	ok := inst.HandleInputEvent(makeInputEvent(event))
	return toPPBool(ok)
}

// Called from C.
func ppp_graphics3d_context_lost(id pp_Instance) {
	inst := getInstanceHandlers(id)
	if inst == nil {
		return
	}
	inst.Graphics3DContextLost()
}

// Called from C.
func ppp_handle_message(id pp_Instance, msg pp_Var) {
	inst := getInstanceHandlers(id)
	if inst == nil {
		return
	}
	inst.HandleMessage(makeVar(msg))
}

// Called from C.
func ppp_mouse_lock_lost(id pp_Instance) {
	inst := getInstanceHandlers(id)
	if inst == nil {
		return
	}
	inst.MouseLockLost()
}
