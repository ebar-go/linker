package buffer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRing(t *testing.T) {
	rb := NewRingBuffer(16)
	rb.Write([]byte("hello"))

	s := make([]byte, 5)
	n, err := rb.Read(s)
	assert.Equal(t, 5, n)
	assert.Nil(t, err)
	assert.Equal(t, []byte("hello"), s)
}
