// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package bnet

// #cgo CFLAGS: -I${SRCDIR}/../../vendor/bncsutil/src
// #cgo !windows LDFLAGS: -lbncsutil_static -lgmp           -L${SRCDIR}/../../vendor/bncsutil/build
// #cgo  windows LDFLAGS: -lbncsutil_static -lgmp -lversion -L${SRCDIR}/../../vendor/bncsutil/build
// #undef __declspec
// #define __declspec(...)
// #include <bncsutil/bncsutil.h>
import (
	"C"
)
import (
	"encoding/binary"
	"net"
	"unsafe"

	"github.com/nielsAD/gowarcraft3/protocol/bncs"
)

// GetExeInfo retrieves version and date/size information from executable file
func GetExeInfo(fileName string) (uint32, string, error) {
	var cstr = C.CString(fileName)
	defer C.free(unsafe.Pointer(cstr))

	var buf [1024]byte
	var ver C.uint32_t
	var res = int(C.getExeInfo(cstr, (*C.char)(unsafe.Pointer(&buf[0])), C.size_t(len(buf)), &ver, C.BNCSUTIL_PLATFORM_X86))

	if res == 0 || res > len(buf) {
		return 0, "", ErrExeInfo
	}

	return uint32(ver), string(buf[:res]), nil
}

// ExtractMPQNumber reads an MPQ filename (e.g. IX86ver#.mpq) and returns the int value of that number
// Returns -1 on failure
func ExtractMPQNumber(mpqName string) int {
	var cstr = C.CString(mpqName)
	defer C.free(unsafe.Pointer(cstr))

	return int(C.extractMPQNumber(cstr))
}

// CheckRevision runs CheckRevision part of BNCS authentication for mpqNumber
// First fileName must be the executable file
func CheckRevision(valueString string, fileNames []string, mpqNumber int) (uint32, error) {
	var cstr = C.CString(valueString)
	defer C.free(unsafe.Pointer(cstr))

	var files = make([](*C.char), len(fileNames))
	for i := 0; i < len(fileNames); i++ {
		files[i] = C.CString(fileNames[i])
		defer C.free(unsafe.Pointer(files[i]))
	}

	var checksum C.ulong
	var res = int(C.checkRevision(cstr, &files[0], C.int(len(files)), C.int(mpqNumber), &checksum))
	if res != 1 {
		return 0, ErrCheckRevision
	}

	return uint32(checksum), nil
}

// CreateBNCSKeyInfo decodes a CD-key, retrieves its relevant values, and calculates a hash suitable for SID_AUTH_CHECK (0x51)
func CreateBNCSKeyInfo(cdkey string, clientToken uint32, serverToken uint32) (*bncs.CDKey, error) {
	var cstr = C.CString(cdkey)
	defer C.free(unsafe.Pointer(cstr))

	var publicValue C.uint32_t
	var productValue C.uint32_t
	var buf [20]byte
	var res = int(C.kd_quick(cstr, C.uint32_t(clientToken), C.uint32_t(serverToken), &publicValue, &productValue, (*C.char)(unsafe.Pointer(&buf[0])), C.size_t(len(buf))))
	if res != 1 {
		return nil, ErrKeyDecoder
	}

	return &bncs.CDKey{
		KeyLength:       uint32(len(cdkey)),
		KeyProductValue: uint32(productValue),
		KeyPublicValue:  uint32(publicValue),
		HashedKeyData:   buf,
	}, nil
}

// VerifyServerSignature received in SID_AUTH_INFO (0x50)
func VerifyServerSignature(ip net.IP, sig *[128]byte) bool {
	var aton = binary.LittleEndian.Uint32(ip.To4())
	return C.nls_check_signature(C.uint32_t(aton), (*C.char)(unsafe.Pointer(&sig[0]))) != 0
}

// SRP password helper
type SRP interface {
	AccountCreate() ([]byte, []byte, error)
	ClientKey() [32]byte
	PasswordProof(serverKey *[32]byte, salt *[32]byte) [20]byte
	VerifyPassword(proof *[20]byte) bool
	Free()
}

// NLS provider for SRP
type NLS struct {
	n *C.nls_t
}

// NewNLS initializes a new NLS provider for SRP
func NewNLS(username string, password string) (*NLS, error) {
	var cstru = C.CString(username)
	defer C.free(unsafe.Pointer(cstru))

	var cstrp = C.CString(password)
	defer C.free(unsafe.Pointer(cstrp))

	var res = NLS{
		n: C.nls_init_l(cstru, C.ulong(len(username)), cstrp, C.ulong(len(password))),
	}
	if res.n == nil {
		return nil, ErrNLS
	}

	return &res, nil
}

// Free NLS struct
func (n *NLS) Free() {
	C.nls_free(n.n)
}

// AccountCreate generates the content for an SID_AUTH_ACCOUNTCREATE packet
func (n *NLS) AccountCreate() ([]byte, []byte, error) {
	var buf [128]byte
	var res = int(C.nls_account_create(n.n, (*C.char)(unsafe.Pointer(&buf[0])), C.ulong(len(buf))))
	if res == 0 {
		return nil, nil, ErrNLS
	}

	return buf[0:32], buf[32:64], nil
}

// ClientKey for SRP exchange
func (n *NLS) ClientKey() (res [32]byte) {
	C.nls_get_A(n.n, (*C.char)(unsafe.Pointer(&res[0])))
	return res
}

// PasswordProof for SRP exchange
func (n *NLS) PasswordProof(serverKey *[32]byte, salt *[32]byte) (res [20]byte) {
	C.nls_get_M1(n.n, (*C.char)(unsafe.Pointer(&res[0])), (*C.char)(unsafe.Pointer(&serverKey[0])), (*C.char)(unsafe.Pointer(&salt[0])))
	return res
}

// VerifyPassword after SRP exchange
func (n *NLS) VerifyPassword(proof *[20]byte) bool {
	return C.nls_check_M2(n.n, (*C.char)(unsafe.Pointer(&proof[0])), nil, nil) != 0
}

// SHA1 provider for SRP
type SHA1 struct {
	password string
}

// NewSHA1 initializes a new SHA1 provider for SRP
func NewSHA1(password string) *SHA1 {
	return &SHA1{
		password: password,
	}
}

// Free SHA1 struct
func (p *SHA1) Free() {}

// AccountCreate generates the content for an SID_AUTH_ACCOUNTCREATE packet
func (p *SHA1) AccountCreate() ([]byte, []byte, error) {
	return nil, []byte(p.password), nil
}

// ClientKey for SRP exchange
func (p *SHA1) ClientKey() (res [32]byte) {
	return res
}

// PasswordProof for SRP exchange
func (p *SHA1) PasswordProof(serverKey *[32]byte, salt *[32]byte) (res [20]byte) {
	var cstrp = C.CString(p.password)
	defer C.free(unsafe.Pointer(cstrp))

	C.hashPassword(cstrp, (*C.char)(unsafe.Pointer(&res[0])))
	return res
}

// VerifyPassword after SRP exchange
func (p *SHA1) VerifyPassword(proof *[20]byte) bool {
	return true
}
