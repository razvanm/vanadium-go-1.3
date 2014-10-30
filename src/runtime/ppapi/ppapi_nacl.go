// Package ppapi provides the Pepper API (PPAPI) to Native Client applications.
// Native Client (NaCl) is a sandbox for running compiled code in the browser
// efficiently and securely, independent of the user’s operating system.  See
// http://developer.chrome.com/native-client for an overview.
//
// This package is based on the Pepper C API, which is documented at
// https://developer.chrome.com/native-client/pepper_dev/c/index#pepper-dev-c-index.
// The functions and types implemented in this package mirror the C API,
// with some reorganization for a more interface-based presentation.
//
// See ${GOROOT}/test/ppapi for some examples of how to use the API.  The main
// parts include 1) a web page to be viewed in the browser, 2) a "manifest" file
// that specifies executable files, and 3) the executable itself, compiled using
// GOOS=nacl.
//
// For example, here is how to set up a basic "Hello world" executable.  The
// HTML file contains an <embed> for the application, where "hello.nmf" is the
// manifest file.
//
//    <div>
//    <embed width=640 height=480 src="hello.nmf" type="application/x-nacl"/>
//    </div>
//
// The manifest file "hello.nmf" lists the executable, using JSON syntax.
//
//     {
//       "program": {
//         "x86-32": {
//           "url": "hello_x86_32.nexe"
//         }
//       }
//     }
//
// Finally, the interesting part is the application.  Each <embed> on the page
// creates a PPAPI Instance.  You have to provide a factory for creating these
// instances (this is like the "main" function for program instance).  Your
// instance must implement the ppapi.InstanceHandlers to receive callbacks from
// the browser.  Here is an example.  For brevity, we'll elide most of the
// callbacks.
//
//     package main
//
//     type myInstance struct {
//       ppapi.Instance
//     }
//
//     // Called when an instance is created (due to an <embed ...>).
//     func (inst *myInstance) DidCreate(argv map[string]string) bool {
//       inst.LogString(ppapi.PP_LOGLEVEL_LOG, "Hello world")
//     }
//
//     // ...other InstanceHandlers methods...
//
//     // In the main function, call ppapi.Init with your instance factory.
//     func main() {
//       ppapi.Init(func (inst ppapi.Instance) ppapi.InstanceHandlers {
//         return &myInstance{Instance: inst}
//       })
//     }
//
// Compile with GOOS=nacl.  You can compile the normal way and copy the binary
// from your bin directory, or else just compile this single file.
//
//     $ GOOS=nacl GOARCH=386 go build -o hello_x86_32.nexe hello.go
//
// Copy your files (hello.html, hello.nmf, hello_x86_32.nexe) to your web server
// and you are done.  See the NaCl documentation for techniques on how to use
// the Javascript console, and how to debug NaCl applications using gdb.
package ppapi

// Init is the main entry point to start PPAPI.  Never returns.  Call this
// function at most once.
func Init(factory func(inst Instance) InstanceHandlers) {
	instanceFactory = factory
	ppapi_start()
}
