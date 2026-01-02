package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"strconv"
)

type RespDataType byte

const (
	Array RespDataType = '*'
	Bulk  RespDataType = '$'
)

type Resp struct {
	conn   net.Conn
	reader *bufio.Reader
}

type RespValue struct {
	Type  RespDataType
	Str   string
	Array []RespValue
}

func NewResp(conn net.Conn) *Resp {
	return &Resp{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}
}

func (r *Resp) Read() (RespValue, error) {
	_type, err := r.reader.ReadByte()

	if err != nil {
		return RespValue{}, err
	}

	switch _type {
	case byte(Array):
		return r.readArray()
	case byte(Bulk):
		return r.readBulk()
	default:
		slog.Error("Invalid type", "type", string(_type))
		return RespValue{}, nil
	}
}

func (r *Resp) readLine() (line []byte, n int, err error) {
	line, err = r.reader.ReadBytes('\n')
	if err != nil {
		return nil, 0, err
	}
	n = len(line)
	if n < 2 || line[n-2] != '\r' {
		return nil, 0, fmt.Errorf("protocol error: line must end with \\r\\n")
	}
	return line[:n-2], n, nil
}

func (r *Resp) readInt() (x int, err error) {
	line, _, err := r.readLine()
	if err != nil {
		return 0, err
	}
	x, err = strconv.Atoi(string(line))
	if err != nil {
		return 0, err
	}
	return x, nil
}

// *<number-of-elements>\r\n<element-1>...<element-n>
func (r *Resp) readArray() (RespValue, error) {
	v := RespValue{
		Type: Array,
	}

	n, err := r.readInt()
	if err != nil {
		return v, err
	}

	v.Array = make([]RespValue, 0)
	for range n {
		val, err := r.Read()
		if err != nil {
			return v, err
		}

		v.Array = append(v.Array, val)
	}

	return v, nil
}

// $<length>\r\n<data>\r\n
func (r *Resp) readBulk() (RespValue, error) {
	v := RespValue{
		Type: Bulk,
	}

	n, err := r.readInt()
	if err != nil {
		return v, err
	}

	bulk := make([]byte, n)
	r.reader.Read(bulk)

	v.Str = string(bulk)

	r.readLine()

	return v, nil
}
