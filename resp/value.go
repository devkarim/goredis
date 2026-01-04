package resp

import "strconv"

type RespDataType byte

const (
	RespString RespDataType = '+'
	RespArray RespDataType = '*'
	RespBulk  RespDataType = '$'
)

type Value struct {
	Type  RespDataType
	Str   string
	Array []Value
}

func (v *Value) Marshal() []byte {
	switch v.Type {
	case RespString:
		return v.marshalString()
	case RespArray:
		return v.marshalArray()
	case RespBulk:
		return v.marshalBulk()
	default:
		return []byte{}
	}
}

// +OK\r\n
func (v *Value) marshalString() []byte {
	var bytes []byte

	bytes = append(bytes, byte(RespString))
	bytes = append(bytes, v.Str...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

// *<number-of-elements>\r\n<element-1>...<element-n>
func (v *Value) marshalArray() []byte {
	length := len(v.Array)
	var bytes []byte

	bytes = append(bytes, byte(RespArray))
	bytes = append(bytes, strconv.Itoa(length)...)
	bytes = append(bytes, '\r', '\n')

	for i := range length {
		bytes = append(bytes, v.Array[i].Marshal()...)
	}

	return bytes
}

// $<length>\r\n<data>\r\n
func (v *Value) marshalBulk() []byte {
	var bytes []byte

	bytes = append(bytes, byte(RespBulk))
	bytes = append(bytes, strconv.Itoa(len(v.Str))...)
	bytes = append(bytes, '\r', '\n')
	bytes = append(bytes, v.Str...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

