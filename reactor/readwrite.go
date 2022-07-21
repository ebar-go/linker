package reactor

import "linker/reactor/buffer"

type ReadWriteLoop struct {
	inboundBuffer *buffer.Buffer
}

func (loop *ReadWriteLoop) Read(conn Conn) ([]byte, error) {
	return nil, nil
}

func (loop *ReadWriteLoop) Write(msg []byte) error {
	return nil
}
