package resp

import "io"

type Writer struct {
	writer io.Writer
}

func (w *Writer) Write(v Value) error {
	bytes := v.Marshal()

	_, err := w.writer.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w}
}

