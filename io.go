package p

import "io"

type Pipe struct {
	Reader       *io.PipeReader
	Writer       *io.PipeWriter
	BytesRead    int64
	BytesWritten int64
}

func NewPipe() *Pipe {
	rp, wp := io.Pipe()
	return &Pipe{
		Reader:       rp,
		Writer:       wp,
		BytesRead:    0,
		BytesWritten: 0,
	}
}

func (p *Pipe) Read(data []byte) (n int, err error) {
	b, err := p.Reader.Read(data)
	p.BytesRead += int64(b)
	return b, err
}

func (p *Pipe) Write(data []byte) (n int, err error) {
	b, err := p.Writer.Write(data)
	p.BytesWritten += int64(b)
	return b, err
}

func (p *Pipe) close() error {
	if err := p.Reader.Close(); err != nil {
		return err
	}
	if err := p.Writer.Close(); err != nil {
		return err
	}
	return nil
}
