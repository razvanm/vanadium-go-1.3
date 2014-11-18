package ppapi

import (
	"errors"
	"fmt"
	"unsafe"
)

var (
	errCreateWebsocketFailed = errors.New("could not create websocket")
)

type WebsocketConn struct {
	Resource
	closed bool
}

func (inst Instance) createWebsocketConn() (ws *WebsocketConn, err error) {
	id := ppb_websocket_create(inst.id)
	if id == 0 {
		err = errCreateWebsocketFailed
		return
	}
	return &WebsocketConn{
		Resource: Resource{
			id: id,
		},
	}, nil
}

func (ws *WebsocketConn) connect(url string) error {
	urlVar := VarFromString(url)
	defer urlVar.Release()
	code := ppb_websocket_connect(ws.id, pp_Var(urlVar), (*pp_Var)(unsafe.Pointer(uintptr(0))), uint32(0), ppNullCompletionCallback)
	return decodeError(Error(code))
}

func (ws *WebsocketConn) Close() error {
	code := ppb_websocket_close(ws.id, uint16(PP_WEBSOCKETSTATUSCODE_NOT_SPECIFIED), pp_Var(VarUndefined), ppNullCompletionCallback)
	return decodeError(Error(code))
}

func (ws *WebsocketConn) sendMessageInternal(message Var) error {
	if ws.closed {
		return fmt.Errorf("Cannot send on closed connection")
	}
	code := ppb_websocket_send_message(ws.id, pp_Var(message))
	return decodeError(Error(code))
}

// Sends a message as a utf-8 string.
func (ws *WebsocketConn) SendMessageUtf8String(m string) error {
	v := VarFromString(m)
	defer v.Release()
	return ws.sendMessageInternal(v)
}

// Sends a byte slice as a message (treating the contents as an array buffer).
func (ws *WebsocketConn) SendMessage(m []byte) error {
	v := VarFromByteSlice(m)
	defer v.Release()
	return ws.sendMessageInternal(v)
}

func (ws *WebsocketConn) receiveMessageInternal() (Var, error) {
	if ws.closed {
		return Var{}, fmt.Errorf("Cannot send on closed connection")
	}
	var pv pp_Var
	code := ppb_websocket_receive_message(ws.id, &pv, ppNullCompletionCallback)
	return Var(pv), decodeError(Error(code))
}

// Reads a message as a utf-8 string (blocking).
func (ws *WebsocketConn) ReceiveMessageUtf8String() (string, error) {
	v, err := ws.receiveMessageInternal()
	if err != nil {
		return "", err
	}
	defer v.Release()
	return v.AsString()
}

// Reads a message as a byte slice (blocking).
func (ws *WebsocketConn) ReceiveMessage() ([]byte, error) {
	v, err := ws.receiveMessageInternal()
	if err != nil {
		return nil, err
	}
	defer v.Release()
	return v.AsByteSlice()
}

// Url must start with "ws:".
func (inst Instance) DialWebsocket(url string) (*WebsocketConn, error) {
	ws, err := inst.createWebsocketConn()
	if err != nil {
		return ws, err
	}

	err = ws.connect(url)
	return ws, err
}
