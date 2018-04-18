// Package mpq provides golang bindings to the StormLib library to read MPQ archives.
package mpq

// #cgo CFLAGS: -I${SRCDIR}/../../vendor/StormLib/src
// #cgo LDFLAGS: -lstorm -lz -lbz2 -lstdc++ -L${SRCDIR}/../../vendor/StormLib
// #include <StormLib.h>
import "C"
import (
	"errors"
	"io"
	"math"
	"os"
	"unsafe"
)

// Errors
var (
	ErrBadFormat    = errors.New("mpq: Bad format")
	ErrArchiveOpen  = errors.New("mpq: Could not open archive")
	ErrArchiveClose = errors.New("mpq: Could not close archive")
	ErrFileOpen     = errors.New("mpq: Could not open subfile")
	ErrFileClose    = errors.New("mpq: Could not close subfile")
	ErrFileRead     = errors.New("mpq: Could not read subfile")
)

func getLastError(def error) error {
	switch C.GetLastError() {
	case C.ERROR_SUCCESS:
		return nil
	case C.ERROR_FILE_NOT_FOUND:
		return os.ErrNotExist
	case C.ERROR_ACCESS_DENIED:
		return os.ErrPermission
	case C.ERROR_INVALID_HANDLE:
		return os.ErrInvalid
	case C.ERROR_INVALID_PARAMETER:
		return os.ErrInvalid
	case C.ERROR_ALREADY_EXISTS:
		return os.ErrExist
	case C.ERROR_INSUFFICIENT_BUFFER:
		return io.ErrShortBuffer
	case C.ERROR_BAD_FORMAT:
		return ErrBadFormat
	case C.ERROR_HANDLE_EOF:
		return io.EOF
	default:
		return def
	}
}

// Archive stores a handle to an opened MPQ archive
type Archive struct {
	h C.HANDLE
}

// File stores a handle to an opened subfile in an MPQ archive
type File struct {
	h C.HANDLE
}

// OpenArchive opens fileName as MPQ archive
func OpenArchive(fileName string) (*Archive, error) {
	var res Archive

	var cstr = C.CString(fileName)
	defer C.free(unsafe.Pointer(cstr))

	//bool SFileOpenArchive(const TCHAR * szMpqName, DWORD dwPriority, DWORD dwFlags, HANDLE * phMpq)
	if C.SFileOpenArchive(cstr, 0, C.MPQ_OPEN_READ_ONLY|C.MPQ_OPEN_NO_LISTFILE|C.MPQ_OPEN_NO_ATTRIBUTES, &res.h) == 0 {
		return nil, getLastError(ErrArchiveOpen)
	}

	return &res, nil
}

// Close an MPQ archive
func (a *Archive) Close() error {
	if a.h != nil {
		if C.SFileCloseArchive(a.h) == 0 {
			return getLastError(ErrArchiveClose)
		}
		a.h = nil
	}
	return nil
}

// WeakSigned checks and verifies the archive against its weak signature if present
func (a *Archive) WeakSigned() bool {
	return C.SFileVerifyArchive(a.h) == C.ERROR_WEAK_SIGNATURE_OK
}

// StrongSigned checks and verifies the archive against its strong signature if present
func (a *Archive) StrongSigned() bool {
	return C.SFileVerifyArchive(a.h) == C.ERROR_STRONG_SIGNATURE_OK
}

// Open a subfile inside an opened MPQ archive
func (a *Archive) Open(subFileName string) (*File, error) {
	var res File

	var cstr = C.CString(subFileName)
	defer C.free(unsafe.Pointer(cstr))

	//bool SFileOpenFileEx(HANDLE hMpq, const char * szFileName, DWORD dwSearchScope, HANDLE * phFile)
	if C.SFileOpenFileEx(a.h, cstr, 0, &res.h) == 0 {
		return nil, getLastError(ErrFileOpen)
	}

	return &res, nil
}

// Close an MPQ subfile
func (f *File) Close() error {
	if f.h != nil {
		if C.SFileCloseFile(f.h) == 0 {
			return getLastError(ErrFileClose)
		}
		f.h = nil
	}
	return nil
}

// Size reports the total file size in bytes
func (f *File) Size() int64 {
	var sizeHigh C.DWORD
	var sizeLow = C.SFileGetFileSize(f.h, &sizeHigh)
	if sizeLow == math.MaxUint32 {
		return -1
	}

	var size = uint64(sizeLow) | uint64(sizeHigh)<<32
	if size > math.MaxInt64 {
		return -1
	}

	return int64(size)
}

// Read implements the io.Reader interface
func (f *File) Read(b []byte) (int, error) {
	var bytesRead C.DWORD

	//bool SFileReadFile(HANDLE hFile, void * lpBuffer, DWORD dwToRead, LPDWORD pdwRead, LPOVERLAPPED lpOverlapped)
	if C.SFileReadFile(f.h, unsafe.Pointer(&b[0]), C.DWORD(len(b)), &bytesRead, nil) == 0 {
		return int(bytesRead), getLastError(ErrFileRead)
	}

	return int(bytesRead), nil
}
