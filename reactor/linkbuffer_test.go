package reactor

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLinkedBuffer(t *testing.T) {
	var llb Buffer
	const dataLen = 5
	data := make([]byte, dataLen)
	copy(data, []byte("hello"))

	r := bytes.NewReader(data)
	n, err := llb.ReadFrom(r)
	require.Nil(t, err)
	require.EqualValues(t, dataLen, n)
	require.EqualValues(t, dataLen, llb.Buffered())

	target := make([]byte, 2)
	llb.Read(target)
	fmt.Println(string(target))
	require.EqualValues(t, dataLen-2, llb.Buffered())
}
