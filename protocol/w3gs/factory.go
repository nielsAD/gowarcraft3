// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package w3gs

// PacketFactory returns a struct of the appropiate type for a packet ID
type PacketFactory interface {
	NewPacket(pid uint8, enc *Encoding) Packet
}

// FactoryFunc creates new Packet
type FactoryFunc func(enc *Encoding) Packet

// MapFactory implements PacketFactory using a map
type MapFactory map[uint8]FactoryFunc

// NewPacket implements PacketFactory interface
func (f MapFactory) NewPacket(pid uint8, enc *Encoding) Packet {
	fun, ok := f[pid]
	if !ok {
		return &UnknownPacket{}
	}
	return fun(enc)
}

type cacheKey struct {
	enc Encoding
	pid uint8
}

// CacheFactory implements a PacketFactory that will only create a type once
type CacheFactory struct {
	factory PacketFactory
	cache   map[cacheKey]Packet
}

// NewFactoryCache initializes CacheFactory
func NewFactoryCache(factory PacketFactory) PacketFactory {
	return &CacheFactory{
		factory: factory,
		cache:   map[cacheKey]Packet{},
	}
}

// NewPacket implements PacketFactory interface
func (f CacheFactory) NewPacket(pid uint8, enc *Encoding) Packet {
	var key = cacheKey{
		enc: *enc,
		pid: pid,
	}

	if p, ok := f.cache[key]; ok {
		return p
	}

	pkt := f.factory.NewPacket(pid, enc)
	f.cache[key] = pkt
	return pkt
}
