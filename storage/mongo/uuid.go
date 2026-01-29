package mongo

import (
	"fmt"
	"reflect"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var tUUID = reflect.TypeOf(uuid.UUID{})

func encodeValueUUID(_ bson.EncodeContext, vw bson.ValueWriter, val reflect.Value) (err error) {

	if !val.IsValid() || val.Type() != tUUID {
		return bson.ValueEncoderError{Name: "encodeValueUUID", Types: []reflect.Type{tUUID}, Received: val}
	}
	data := make([]byte, 16)
	for idx := 0; idx < val.Len(); idx++ {
		data[idx] = val.Index(idx).Interface().(byte)
	}
	return vw.WriteBinaryWithSubtype(data, bson.TypeBinaryUUID)
}

func decodeValueUUID(_ bson.DecodeContext, vr bson.ValueReader, val reflect.Value) (err error) {

	if !val.CanSet() || val.Type() != tUUID {
		return bson.ValueDecoderError{Name: "decodeValueUUID", Types: []reflect.Type{tUUID}, Received: val}
	}
	if vr.Type() != bson.TypeBinary {
		return fmt.Errorf("cannot decode %v into an UUID", vr.Type())
	}
	data, subtype, err := vr.ReadBinary()
	if err != nil {
		return err
	}
	if len(data) != 16 {
		return fmt.Errorf("decodeValueUUID cannot decode binary, invalid length %v", len(data))
	}
	if subtype != bson.TypeBinaryUUID {
		return fmt.Errorf("decodeValueUUID can only be used to decode subtype 0x4 for %s, got %v", bson.TypeBinary, subtype)
	}
	id := uuid.UUID{}
	for idx := 0; idx < 16; idx++ {
		id[idx] = data[idx]
	}
	val.Set(reflect.ValueOf(id))
	return
}

// newBSONRegistryWithUUID — BSON registry с поддержкой uuid.UUID как binary subtype 0x04.
func newBSONRegistryWithUUID() (reg *bson.Registry) {

	reg = bson.NewMgoRegistry()
	reg.RegisterTypeEncoder(tUUID, bson.ValueEncoderFunc(encodeValueUUID))
	reg.RegisterTypeDecoder(tUUID, bson.ValueDecoderFunc(decodeValueUUID))
	return
}
