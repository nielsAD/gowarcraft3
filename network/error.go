// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package network

import (
	"io"
	"net"
	"os"
	"syscall"

	"github.com/gorilla/websocket"
)

// AsyncError keeps track of where a non-fatal asynchronous error orignated
type AsyncError struct {
	Src string
	Err error
}

func (e *AsyncError) Error() string {
	if e.Err == nil {
		return e.Src + ":NIL"
	}
	return e.Src + ":" + e.Err.Error()
}

// UnnestError retrieves the innermost error
func UnnestError(err error) error {
	switch e := err.(type) {
	case *AsyncError:
		return UnnestError(e.Err)
	case *net.OpError:
		return UnnestError(e.Err)
	case *os.SyscallError:
		return UnnestError(e.Err)
	case *os.PathError:
		return UnnestError(e.Err)
	case *os.LinkError:
		return UnnestError(e.Err)
	default:
		return err
	}
}

// IsUseClosedNetworkError checks if net.error is poll.ErrNetClosed
func IsUseClosedNetworkError(err error) bool {
	err = UnnestError(err)
	return err != nil && err.Error() == "use of closed network connection"
}

// IsSysCallError checks if net.error is syscall.Errno
func IsSysCallError(err error, errno syscall.Errno) bool {
	err = UnnestError(err)
	return err != nil && err.Error() == errno.Error()
}

// WSAECONNREFUSED is ECONNREFUSED on Windows
const WSAECONNREFUSED = 10061

// IsConnRefusedError checks if net.error is a "connection refused" error
func IsConnRefusedError(err error) bool {
	err = UnnestError(err)
	return IsSysCallError(err, syscall.ECONNREFUSED) ||
		IsSysCallError(err, WSAECONNREFUSED) ||
		websocket.IsCloseError(err, websocket.CloseTLSHandshake, websocket.CloseMandatoryExtension)
}

// WSAECONNABORTED is ECONNABORTED on Windows
const WSAECONNABORTED = 10053

// WSAECONNRESET is ECONNRESET on Windows
const WSAECONNRESET = 10054

// WSAENOTCONN is ENOTCONN on Windows
const WSAENOTCONN = 10057

// WSAESHUTDOWN is ESHUTDOWN on Windows
const WSAESHUTDOWN = 10058

// IsConnClosedError checks if net.error is a "connection closed" error
func IsConnClosedError(err error) bool {
	err = UnnestError(err)
	return err == io.EOF ||
		err == websocket.ErrCloseSent ||
		IsUseClosedNetworkError(err) ||
		IsSysCallError(err, syscall.ECONNABORTED) || IsSysCallError(err, WSAECONNABORTED) ||
		IsSysCallError(err, syscall.ECONNRESET) || IsSysCallError(err, WSAECONNRESET) ||
		IsSysCallError(err, syscall.ENOTCONN) || IsSysCallError(err, WSAENOTCONN) ||
		IsSysCallError(err, syscall.ESHUTDOWN) || IsSysCallError(err, WSAESHUTDOWN) ||
		IsSysCallError(err, syscall.EPIPE) ||
		websocket.IsUnexpectedCloseError(err)
}
