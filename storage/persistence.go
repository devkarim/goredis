package storage

import (
	"io"
	"os"
	"sync"

	"github.com/devkarim/goredis/resp"
)

type Aof struct {
	mu   sync.Mutex
	file *os.File
}

func NewAof(path string) (*Aof, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)

	if err != nil {
		return nil, err
	}

	return &Aof{file: file}, nil
}

func (aof *Aof) Write(v resp.Value) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	_, err := aof.file.Write(v.Marshal())
	if err != nil {
		return err
	}

	return nil
}

func (aof *Aof) Read(callback func(val resp.Value)) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	// start at the beginning of the file
	_, err := aof.file.Seek(0, 0)
	if err != nil {
		return err
	}

	rd := resp.NewReader(aof.file)

	for {
		val, err := rd.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		callback(val)
	}

	return nil
}

func (aof *Aof) Close() error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	err := aof.file.Close()

	if err != nil {
		return err
	}

	return nil
}

func (aof *Aof) Sync() {
	aof.mu.Lock()
	aof.file.Sync()
	defer aof.mu.Unlock()
}
