package resp

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"strconv"
)

type Reader struct {
	conn   net.Conn
	reader *bufio.Reader
}

func NewReader(conn net.Conn) *Reader {
	return &Reader{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}
}

func (r *Reader) Read() (Value, error) {
	_type, err := r.reader.ReadByte()

	if err != nil {
		return Value{}, err
	}

	switch _type {
	case byte(RespArray):
		return r.readArray()
	case byte(RespBulk):
		return r.readBulk()
	default:
		slog.Error("Invalid type", "type", string(_type))
		return Value{}, nil
	}
}

func (r *Reader) readLine() (line []byte, n int, err error) {
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

func (r *Reader) readInt() (x int, err error) {
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
func (r *Reader) readArray() (Value, error) {
	v := Value{
		Type: RespArray,
	}

	n, err := r.readInt()
	if err != nil {
		return v, err
	}

	v.Array = make([]Value, 0)
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
func (r *Reader) readBulk() (Value, error) {
	v := Value{
		Type: RespBulk,
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

