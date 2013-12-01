package marshal

import (
	"errors"
	// "log"
	"strconv"
)

type MarshalledObject struct {
	MajorVersion byte
	MinorVersion byte
	data         []byte
}

type marshalledObjectType byte

var TypeMismatch = errors.New("gorails/marshal: an attempt to implicitly typecast the marshalled object")
var IncompleteData = errors.New("gorails/marshal: incomplete data")

const (
	TYPE_UNKNOWN marshalledObjectType = 0
	TYPE_NIL     marshalledObjectType = 1
	TYPE_BOOLEAN marshalledObjectType = 2
	TYPE_INTEGER marshalledObjectType = 3
	TYPE_FLOAT   marshalledObjectType = 4
	TYPE_STRING  marshalledObjectType = 5
	TYPE_ARRAY   marshalledObjectType = 6
	TYPE_MAP     marshalledObjectType = 7
)

func CreateMarshalledObject(serialized_data []byte) *MarshalledObject {
	return &(MarshalledObject{serialized_data[0], serialized_data[1], serialized_data[2:]})
}

func assertType(obj *MarshalledObject, expected_type marshalledObjectType) (err error) {
	if obj.GetType() != expected_type {
		err = TypeMismatch
	}

	return
}

func (obj *MarshalledObject) GetType() marshalledObjectType {
	if len(obj.data) == 0 {
		return TYPE_UNKNOWN
	}

	switch obj.data[0] {
	case '0':
		return TYPE_NIL
	case 'T', 'F':
		return TYPE_BOOLEAN
	case 'i':
		return TYPE_INTEGER
	case 'f':
		return TYPE_FLOAT
	case ':':
		return TYPE_STRING
	case 'I':
		if len(obj.data) > 1 && obj.data[1] == '"' {
			return TYPE_STRING
		}
	case '[':
		return TYPE_ARRAY
	case '{':
		return TYPE_MAP
	}

	return TYPE_UNKNOWN
}

func (obj *MarshalledObject) GetAsBoolean() (value bool, err error) {
	err = assertType(obj, TYPE_BOOLEAN)
	if err == nil {
		value = obj.data[0] == 'T'
	}

	return
}

func parseInt(data []byte) int {
	if data[0] > 0x05 && data[0] < 0xfb {
		value := int(data[0])

		if value > 0x7f {
			return -(0xff ^ value + 1) + 5
		} else {
			return value - 5
		}
	} else if data[0] <= 0x05 {
		value := 0
		i := data[0]

		for ; i > 0; i-- {
			value = value<<8 + int(data[i])
		}

		return value
	} else {
		value := 0
		i := 0xff - data[0] + 1

		for ; i > 0; i-- {
			value = value<<8 + (0xff - int(data[i]))
		}

		return -(value + 1)
	}
}

func (obj *MarshalledObject) GetAsInteger() (value int, err error) {
	err = assertType(obj, TYPE_INTEGER)
	if err != nil {
		return
	}

	value = parseInt(obj.data[1:])

	return
}

func (obj *MarshalledObject) GetAsFloat() (value float64, err error) {
	err = assertType(obj, TYPE_FLOAT)
	if err != nil {
		return
	}

	value, err = strconv.ParseFloat(parseString(obj.data[1:]), 64)

	return
}

func parseString(data []byte) string {
	if data[0] > 0x05 {
		length := parseInt(data[0:1])
		return string(data[1 : length+1])
	} else {
		length := parseInt(data[0 : data[0]+1])
		return string(data[data[0]+1 : length+int(data[0])+1])
	}
}

func (obj *MarshalledObject) GetAsString() (value string, err error) {
	err = assertType(obj, TYPE_STRING)
	if err != nil {
		return
	}

	if obj.data[0] == ':' {
		value = parseString(obj.data[1:])
	} else {
		value = parseString(obj.data[2:])
	}

	return
}