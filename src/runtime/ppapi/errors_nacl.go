// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppapi

import (
	"errors"
	"fmt"
)

var (
	ppErrors = map[Error]error{
		/**
		 * This value is returned by a function on successful synchronous completion
		 * or is passed as a result to a PP_CompletionCallback_Func on successful
		 * asynchronous completion.
		 */
		PP_OK: nil,
		/**
		 * This value is returned by a function that accepts a PP_CompletionCallback
		 * and cannot complete synchronously. This code indicates that the given
		 * callback will be asynchronously notified of the final result once it is
		 * available.
		 */
		PP_OK_COMPLETIONPENDING: errors.New("ok: completion pending"),
		/**This value indicates failure for unspecified reasons. */
		PP_ERROR_FAILED: errors.New("failed"),
		/**
		 * This value indicates failure due to an asynchronous operation being
		 * interrupted. The most common cause of this error code is destroying a
		 * resource that still has a callback pending. All callbacks are guaranteed
		 * to execute, so any callbacks pending on a destroyed resource will be
		 * issued with PP_ERROR_ABORTED.
		 *
		 * If you get an aborted notification that you aren't expecting, check to
		 * make sure that the resource you're using is still in scope. A common
		 * mistake is to create a resource on the stack, which will destroy the
		 * resource as soon as the function returns.
		 */
		PP_ERROR_ABORTED: errors.New("aborted"),
		/** This value indicates failure due to an invalid argument. */
		PP_ERROR_BADARGUMENT: errors.New("bad argument"),
		/** This value indicates failure due to an invalid PP_Resource. */
		PP_ERROR_BADRESOURCE: errors.New("bad resource"),
		/** This value indicates failure due to an unavailable PPAPI interface. */
		PP_ERROR_NOINTERFACE: errors.New("no interface"),
		/** This value indicates failure due to insufficient privileges. */
		PP_ERROR_NOACCESS: errors.New("access denied"),
		/** This value indicates failure due to insufficient memory. */
		PP_ERROR_NOMEMORY: errors.New("out of memory"),
		/** This value indicates failure due to insufficient storage space. */
		PP_ERROR_NOSPACE: errors.New("out of space"),
		/** This value indicates failure due to insufficient storage quota. */
		PP_ERROR_NOQUOTA: errors.New("quota exhausted"),
		/**
		 * This value indicates failure due to an action already being in
		 * progress.
		 */
		PP_ERROR_INPROGRESS: errors.New("operation already in progress"),
		/**
		 * The requested command is not supported by the browser.
		 */
		PP_ERROR_NOTSUPPORTED: errors.New("operation not supported"),
		/**
		 * Returned if you try to use a null completion callback to "block until
		 * complete" on the main thread. Blocking the main thread is not permitted
		 * to keep the browser responsive (otherwise, you may not be able to handle
		 * input events, and there are reentrancy and deadlock issues).
		 */
		PP_ERROR_BLOCKS_MAIN_THREAD: errors.New("operation would block the main thread"),
		/** This value indicates failure due to a file that does not exist. */
		PP_ERROR_FILENOTFOUND: errors.New("file not found"),
		/** This value indicates failure due to a file that already exists. */
		PP_ERROR_FILEEXISTS: errors.New("file exists"),
		/** This value indicates failure due to a file that is too big. */
		PP_ERROR_FILETOOBIG: errors.New("file too big"),
		/**
		 * This value indicates failure due to a file having been modified
		 * unexpectedly.
		 */
		PP_ERROR_FILECHANGED: errors.New("file changed"),
		/** This value indicates that the pathname does not reference a file. */
		PP_ERROR_NOTAFILE: errors.New("not a file"),
		/** This value indicates failure due to a time limit being exceeded. */
		PP_ERROR_TIMEDOUT: errors.New("operation timed out"),
		/**
		 * This value indicates that the user cancelled rather than providing
		 * expected input.
		 */
		PP_ERROR_USERCANCEL: errors.New("operation was canceled by the user"),
		/**
		 * This value indicates failure due to lack of a user gesture such as a
		 * mouse click or key input event. Examples of actions requiring a user
		 * gesture are showing the file chooser dialog and going into fullscreen
		 * mode.
		 */
		PP_ERROR_NO_USER_GESTURE: errors.New("no user gesture"),
		/**
		 * This value indicates that the graphics context was lost due to a
		 * power management event.
		 */
		PP_ERROR_CONTEXT_LOST: errors.New("graphics context was lost"),
		/**
		 * Indicates an attempt to make a PPAPI call on a thread without previously
		 * registering a message loop via PPB_MessageLoop.AttachToCurrentThread.
		 * Without this registration step, no PPAPI calls are supported.
		 */
		PP_ERROR_NO_MESSAGE_LOOP: errors.New("no message loop"),
		/**
		 * Indicates that the requested operation is not permitted on the current
		 * thread.
		 */
		PP_ERROR_WRONG_THREAD: errors.New("operation is not permitted on the current thread"),
		/**
		 * This value indicates that the connection was closed. For TCP sockets, it
		 * corresponds to a TCP FIN.
		 */
		PP_ERROR_CONNECTION_CLOSED: errors.New("connection closed"),
		/**
		 * This value indicates that the connection was reset. For TCP sockets, it
		 * corresponds to a TCP RST.
		 */
		PP_ERROR_CONNECTION_RESET: errors.New("connection reset"),
		/**
		 * This value indicates that the connection attempt was refused.
		 */
		PP_ERROR_CONNECTION_REFUSED: errors.New("connection refused"),
		/**
		 * This value indicates that the connection was aborted. For TCP sockets, it
		 * means the connection timed out as a result of not receiving an ACK for data
		 * sent. This can include a FIN packet that did not get ACK'd.
		 */
		PP_ERROR_CONNECTION_ABORTED: errors.New("connection aborted"),
		/**
		 * This value indicates that the connection attempt failed.
		 */
		PP_ERROR_CONNECTION_FAILED: errors.New("connection failed"),
		/**
		 * This value indicates that the connection attempt timed out.
		 */
		PP_ERROR_CONNECTION_TIMEDOUT: errors.New("connection timed out"),
		/**
		 * This value indicates that the IP address or port number is invalid.
		 */
		PP_ERROR_ADDRESS_INVALID: errors.New("invalid address"),
		/**
		 * This value indicates that the IP address is unreachable. This usually means
		 * that there is no route to the specified host or network.
		 */
		PP_ERROR_ADDRESS_UNREACHABLE: errors.New("address unreachable"),
		/**
		 * This value is returned when attempting to bind an address that is already
		 * in use.
		 */
		PP_ERROR_ADDRESS_IN_USE: errors.New("address in use"),
		/**
		 * This value indicates that the message was too large for the transport.
		 */
		PP_ERROR_MESSAGE_TOO_BIG: errors.New("message too big"),
		/**
		 * This value indicates that the host name could not be resolved.
		 */
		PP_ERROR_NAME_NOT_RESOLVED: errors.New("name can't be resolved"),
	}
)

func decodeError(code Error) error {
	err, ok := ppErrors[code]
	if ok {
		return err
	}
	return fmt.Errorf("unknown ppapi error %d", code)
}
