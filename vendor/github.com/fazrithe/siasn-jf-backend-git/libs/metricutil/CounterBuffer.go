package metricutil

import "io"

// Deprecated: use datacount package instead.
// CounterBuffer implements io.ReadWriter to store read and write count.
// Call Reset, ResetRead, or ResetWrite to reset the counter.
type CounterBuffer struct {
	io.Reader
	io.Writer
	ReadCount  int
	WriteCount int
}

func (buffer *CounterBuffer) Read(p []byte) (n int, err error) {
	defer func() {
		buffer.ReadCount += n
	}()
	return buffer.Reader.Read(p)
}

func (buffer *CounterBuffer) Write(p []byte) (n int, err error) {
	defer func() {
		buffer.WriteCount += n
	}()
	return buffer.Writer.Write(p)
}

func (buffer *CounterBuffer) ResetRead() {
	buffer.ReadCount = 0
}

func (buffer *CounterBuffer) ResetWrite() {
	buffer.WriteCount = 0
}

func (buffer *CounterBuffer) Reset() {
	buffer.ResetRead()
	buffer.ResetWrite()
}
