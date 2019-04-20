// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package protocol_test

import (
	"fmt"

	"github.com/nielsAD/gowarcraft3/protocol"
)

func Example() {
	var pbuf protocol.Buffer
	pbuf.WriteUInt32(0x01)
	pbuf.WriteUInt16(0x02)
	pbuf.WriteUInt8(0x03)
	pbuf.WriteCString("4")

	fmt.Println(pbuf.ReadUInt32())
	fmt.Println(pbuf.ReadUInt16())
	fmt.Println(pbuf.ReadUInt8())

	str, err := pbuf.ReadCString()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(str)

	pbuf = protocol.Buffer{Bytes: []byte{5, 0, 0, 0}}
	fmt.Println(pbuf.ReadUInt32())

	// output:
	// 1
	// 2
	// 3
	// 4
	// 5
}
