package raw

import (
	"maps"
	"slices"

	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"
)

// IndexHandlers is encoded by hand: msgp ranges over a map in random order, which would make the same
// translations encode to different bytes on every run.
//
//msgp:ignore IndexHandlers
type IndexHandlers map[string]string

// EncodeMsg implements msgp.Encodable, writing keys in sorted order.
func (z IndexHandlers) EncodeMsg(en *msgp.Writer) error {
	if err := en.WriteMapHeader(uint32(len(z))); err != nil {
		return errors.Wrap(err, "failed to write map header")
	}
	for _, k := range slices.Sorted(maps.Keys(z)) {
		if err := en.WriteString(k); err != nil {
			return errors.Wrap(err, "failed to write key")
		}
		if err := en.WriteString(z[k]); err != nil {
			return errors.Wrap(err, "failed to write value of "+k)
		}
	}
	return nil
}

// MarshalMsg implements msgp.Marshaler, writing keys in sorted order.
func (z IndexHandlers) MarshalMsg(b []byte) ([]byte, error) {
	o := msgp.Require(b, z.Msgsize())
	o = msgp.AppendMapHeader(o, uint32(len(z)))
	for _, k := range slices.Sorted(maps.Keys(z)) {
		o = msgp.AppendString(o, k)
		o = msgp.AppendString(o, z[k])
	}
	return o, nil
}

// DecodeMsg implements msgp.Decodable.
func (z *IndexHandlers) DecodeMsg(dc *msgp.Reader) error {
	sz, err := dc.ReadMapHeader()
	if err != nil {
		return errors.Wrap(err, "failed to read map header")
	}
	if *z == nil {
		*z = make(IndexHandlers, sz)
	} else {
		clear(*z)
	}
	for range sz {
		k, err := dc.ReadString()
		if err != nil {
			return errors.Wrap(err, "failed to read key")
		}
		v, err := dc.ReadString()
		if err != nil {
			return errors.Wrap(err, "failed to read value of "+k)
		}
		(*z)[k] = v
	}
	return nil
}

// UnmarshalMsg implements msgp.Unmarshaler.
func (z *IndexHandlers) UnmarshalMsg(bts []byte) ([]byte, error) {
	sz, bts, err := msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return bts, errors.Wrap(err, "failed to read map header")
	}
	if *z == nil {
		*z = make(IndexHandlers, sz)
	} else {
		clear(*z)
	}
	for range sz {
		var k, v string
		if k, bts, err = msgp.ReadStringBytes(bts); err != nil {
			return bts, errors.Wrap(err, "failed to read key")
		}
		if v, bts, err = msgp.ReadStringBytes(bts); err != nil {
			return bts, errors.Wrap(err, "failed to read value of "+k)
		}
		(*z)[k] = v
	}
	return bts, nil
}

// Msgsize implements msgp.Sizer.
func (z IndexHandlers) Msgsize() int {
	s := msgp.MapHeaderSize
	for k, v := range z {
		s += msgp.StringPrefixSize + len(k) + msgp.StringPrefixSize + len(v)
	}
	return s
}
