// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "../runtime.h"
#include "../cgocall.h"
#include "../../cmd/ld/textflag.h"
#include "ppapi_GOOS.h"

typedef void *(*PPB_GetInterface)(int8 *interface_name);

// PPBInterface contains the PPAPI interface name and a pointer to the PPAPI
// struct containing the function pointers for that interface.
struct PPBInterface {
	int8 *name;
	void *ppb;
};

typedef int32 (*PPBFunction)(uintptr arg1, uintptr arg2, uintptr arg3);

// PPPInterface specifies an interface for callbacks.
struct PPPInterface {
	int8 *name;
	PPBFunction *functions;
};

// ppapi·module_id is the module identifier for the process instance.
int32 ppapi·module_id;

#pragma dataflag NOPTR
struct PPBInterface ppapi·ppb_interfaces[] = {
	{ "PPB_Audio;1.1" },
	{ "PPB_AudioBuffer;0.1" },
	{ "PPB_AudioConfig;1.1" },
	{ "PPB_Console;1.0" },
	{ "PPB_Core;1.0" },
	{ "PPB_FileIO;1.1" },
	{ "PPB_FileMapping;0.1" },
	{ "PPB_FileRef;1.2" },
	{ "PPB_FileSystem;1.0" },
	{ "PPB_Fullscreen;1.0" },
	{ "PPB_Gamepad;1.0" },
	{ "PPB_Graphics2D;1.1" },
	{ "PPB_Graphics3D;1.0" },
	{ "PPB_HostResolver;1.0" },
	{ "PPB_ImageData;1.0" },
	{ "PPB_InputEvent;1.0" },
	{ "PPB_MouseInputEvent;1.1" },
	{ "PPB_WheelInputEvent;1.0" },
	{ "PPB_KeyboardInputEvent;1.2" },
	{ "PPB_TouchInputEvent;1.0" },
	{ "PPB_IMEInputEvent;1.0" },
	{ "PPB_Instance;1.0" },
	{ "PPB_MediaStreamAudioTrack;0.1" },
	{ "PPB_MediaStreamVideoTrack;0.1" },
	{ "PPB_MessageLoop;1.0" },
	{ "PPB_Messaging;1.0" },
	{ "PPB_MouseCursor;1.0" },
	{ "PPB_MouseLock;1.0" },
	{ "PPB_NetAddress;1.0" },
	{ "PPB_NetworkList;1.0" },
	{ "PPB_NetworkMonitor;1.0" },
	{ "PPB_NetworkProxy;1.0" },
	{ "PPB_OpenGLES2;1.0" },
	{ "PPB_TCPSocket;1.1" },
	{ "PPB_TextInputController;1.0" },
	{ "PPB_UDPSocket;1.0" },
	{ "PPB_URLLoader;1.0" },
	{ "PPB_URLRequestInfo;1.0" },
	{ "PPB_URLResponseInfo;1.0" },
	{ "PPB_Var;1.2" },
	{ "PPB_VarArray;1.0" },
	{ "PPB_VarArrayBuffer;1.0" },
	{ "PPB_VarDictionary;1.0" },
	{ "PPB_VideoFrame;0.1" },
	{ "PPB_View;1.1" },
	{ "PPB_WebSocket;1.0" },
	{ 0 },
};

struct pp_Var {
	int32 ty;
	int32 pad;
	int64 value;
};

void ppapi·start(void *arg);
void ppapi·breakpoint(void);

int32 ppapi·ppp_initialize_module_handler(int32 module_id, PPBFunction get_interface);
void ppapi·ppp_shutdown_module_handler(void);
void ppapi·ppp_get_interface_handler(int8 *interface_name);

void ppapi·ppp_graphics3d_context_lost(int32 instance);
int32 ppapi·ppp_handle_input_event(int32 instance, int32 event);
int32 ppapi·ppp_did_create(int32 instance, int32 argc, int8 **argn, int8 **argv);
void ppapi·ppp_did_destroy(int32 instance);
void ppapi·ppp_did_change_view(int32 instance, int32 view);
void ppapi·ppp_did_change_focus(int32 instance, int32 has_focus);
int32 ppapi·ppp_handle_document_load(int32 instance, int32 url_loader);
void ppapi·ppp_handle_message(int32 instance, struct pp_Var msg);
void ppapi·ppp_mouse_lock_lost(int32 instance);
void *ppapi·get_array_output_buffer(void *data, uint32 count, uint32 size);

// PPP_Graphics3D callbacks.
#pragma dataflag NOPTR
static PPBFunction ppapi·ppp_graphics_3d[] = {
	(PPBFunction) ppapi·ppp_graphics3d_context_lost,
};

// PPP_InputEvent callbacks.
#pragma dataflag NOPTR
static PPBFunction ppapi·ppp_input_event[] = {
	(PPBFunction) ppapi·ppp_handle_input_event,
};

// PPP_Instance callbacks.
#pragma dataflag NOPTR
static PPBFunction ppapi·ppp_instance[] = {
	(PPBFunction) ppapi·ppp_did_create,
	(PPBFunction) ppapi·ppp_did_destroy,
	(PPBFunction) ppapi·ppp_did_change_view,
	(PPBFunction) ppapi·ppp_did_change_focus,
	(PPBFunction) ppapi·ppp_handle_document_load,
};

// PPP_Messaging callbacks.
#pragma dataflag NOPTR
static PPBFunction ppapi·ppp_messaging[] = {
	(PPBFunction) ppapi·ppp_handle_message,
};

// PPP_MouseLock callbacks.
#pragma dataflag NOPTR
static PPBFunction ppapi·ppp_mouse_lock[] = {
	(PPBFunction) ppapi·ppp_mouse_lock_lost,
};

#pragma dataflag NOPTR
static struct PPPInterface ppapi·ppp_interfaces[] = {
	{ "PPP_Graphics_3D;1.0", ppapi·ppp_graphics_3d },
	{ "PPP_InputEvent;0.1",  ppapi·ppp_input_event },
	{ "PPP_Instance;1.1",    ppapi·ppp_instance },
	{ "PPP_Messaging;1.0",   ppapi·ppp_messaging },
	{ "PPP_MouseLock;1.0",   ppapi·ppp_mouse_lock },
	{ 0 },
};

#pragma dataflag NOPTR
PPBFunction ppapi·pp_start_functions[] = {
	(PPBFunction) ppapi·ppp_initialize_module_handler,
	(PPBFunction) ppapi·ppp_shutdown_module_handler,
	(PPBFunction) ppapi·ppp_get_interface_handler,
};

#pragma textflag NOSPLIT
void *ppapi·ppp_get_interface(int8 *interface_name) {
	struct PPPInterface *intf;
	for (intf = ppapi·ppp_interfaces; intf->name != 0; intf++) {
	  if (runtime·strcmp((byte *) intf->name, (byte *) interface_name) == 0)
	    return intf->functions;
	}
	return 0;
}

// C array allocator.
struct ppapi·ArrayOutput {
	void *(*get_data_buffer)(void *user_data, uint32 count, uint32 size);
	void *user_data;
};

#pragma textflag NOSPLIT
void ·init_array_output(struct ppapi·ArrayOutput *aout, void *alloc) {
	aout->get_data_buffer = ppapi·get_array_output_buffer;
	aout->user_data = alloc;
}
