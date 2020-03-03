// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package w3g

// RecordFactory returns a struct of the appropiate type for a record ID
type RecordFactory interface {
	NewRecord(rid uint8, enc *Encoding) Record
}

// FactoryFunc creates new Record
type FactoryFunc func(enc *Encoding) Record

// MapFactory implements RecordFactory using a map
type MapFactory map[uint8]FactoryFunc

// NewRecord implements RecordFactory interface
func (f MapFactory) NewRecord(rid uint8, enc *Encoding) Record {
	fun, ok := f[rid]
	if !ok {
		return nil
	}
	return fun(enc)
}

type cacheKey struct {
	enc Encoding
	rid uint8
}

// CacheFactory implements a RecordFactory that will only create a type once
type CacheFactory struct {
	factory RecordFactory
	cache   map[cacheKey]Record
}

// NewFactoryCache initializes CacheFactory
func NewFactoryCache(factory RecordFactory) RecordFactory {
	return &CacheFactory{
		factory: factory,
		cache:   map[cacheKey]Record{},
	}
}

// NewRecord implements RecordFactory interface
func (f CacheFactory) NewRecord(rid uint8, enc *Encoding) Record {
	var key = cacheKey{
		enc: *enc,
		rid: rid,
	}

	if p, ok := f.cache[key]; ok {
		return p
	}

	pkt := f.factory.NewRecord(rid, enc)
	f.cache[key] = pkt
	return pkt
}
