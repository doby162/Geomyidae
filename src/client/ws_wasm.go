//go:build js && wasm
// +build js,wasm

package main

import (
	"errors"
	"syscall/js"
	"time"
)

// wasmConn is a minimal WebSocket wrapper using the browser WebSocket API.
type wasmConn struct {
	ws      js.Value
	recv    chan []byte
	closeCh chan struct{}
	closed  bool
}

func dialWasmWebSocket(url string) (*wasmConn, error) {
	wsCtor := js.Global().Get("WebSocket")
	if wsCtor.IsUndefined() {
		return nil, errors.New("WebSocket not available in JS global")
	}
	ws := wsCtor.New(url)
	c := &wasmConn{
		ws:      ws,
		recv:    make(chan []byte, 256),
		closeCh: make(chan struct{}),
	}

	ws.Set("binaryType", "arraybuffer")

	openCh := make(chan struct{}, 1)
	errCh := make(chan error, 1)

	msgCb := js.FuncOf(func(this js.Value, args []js.Value) any {
		ev := args[0]
		data := ev.Get("data")
		switch data.Type() {
		case js.TypeString:
			c.recv <- []byte(data.String())
		default:
			uint8Arr := js.Global().Get("Uint8Array").New(data)
			b := make([]byte, uint8Arr.Get("length").Int())
			js.CopyBytesToGo(b, uint8Arr)
			c.recv <- b
		}
		return nil
	})
	openCb := js.FuncOf(func(this js.Value, args []js.Value) any {
		openCh <- struct{}{}
		return nil
	})
	errCb := js.FuncOf(func(this js.Value, args []js.Value) any {
		errCh <- errors.New("websocket error")
		return nil
	})
	closeCb := js.FuncOf(func(this js.Value, args []js.Value) any {
		if !c.closed {
			c.closed = true
			close(c.closeCh)
		}
		return nil
	})

	ws.Call("addEventListener", "message", msgCb)
	ws.Call("addEventListener", "open", openCb)
	ws.Call("addEventListener", "error", errCb)
	ws.Call("addEventListener", "close", closeCb)

	select {
	case <-openCh:
	case <-errCh:
		return nil, errors.New("websocket open error")
	case <-time.After(5 * time.Second):
		return nil, errors.New("websocket open timeout")
	}

	// Note: In a full implementation we'd release the js.Func callbacks on Close().
	return c, nil
}

func (c *wasmConn) ReadMessage() (int, []byte, error) {
	select {
	case b := <-c.recv:
		return 1, b, nil
	case <-c.closeCh:
		return 0, nil, errors.New("closed")
	}
}

func (c *wasmConn) WriteMessage(messageType int, data []byte) error {
	if c.closed {
		return errors.New("closed")
	}
	if messageType == 1 { // TextMessage
		c.ws.Call("send", string(data))
		return nil
	}
	uint8Arr := js.Global().Get("Uint8Array").New(len(data))
	js.CopyBytesToJS(uint8Arr, data)
	c.ws.Call("send", uint8Arr)
	return nil
}

func (c *wasmConn) Close() error {
	if c.closed {
		return nil
	}
	c.ws.Call("close")
	c.closed = true
	return nil
}

// DialWS for wasm uses the browser WebSocket API and returns WSConn.
func DialWS(u string) (WSConn, error) {
	return dialWasmWebSocket(u)
}
