package util

import (
	"io"
	"net"
	"os"
	"syscall"
)

// ErrUseClosedConn checks if net.error is poll.ErrNetClosed
func ErrUseClosedConn(err error) bool {
	if err == nil {
		return false
	}
	if operr, ok := err.(*net.OpError); ok {
		return ErrUseClosedConn(operr.Err)
	}
	return err.Error() == "use of closed network connection"
}

// ErrSysCall checks if net.error is syscall.Errno
func ErrSysCall(err error, errno syscall.Errno) bool {
	if err == nil {
		return false
	}
	if operr, ok := err.(*net.OpError); ok {
		return ErrSysCall(operr.Err, errno)
	}
	if scerr, ok := err.(*os.SyscallError); ok {
		return ErrSysCall(scerr.Err, errno)
	}
	return err.Error() == errno.Error()
}

// ErrConnClosed checks if net.error is a "connection closed" error
func ErrConnClosed(err error) bool {
	return err == io.EOF || ErrUseClosedConn(err) || ErrSysCall(err, syscall.ECONNRESET)
}
