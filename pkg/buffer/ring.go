package buffer

import (
	"golang.org/x/sys/unix"
	"io"
	"syscall"
)

// RingBuffer 环形队列
type RingBuffer struct {
	bs      [][]byte
	buf     []byte // 数据切片
	size    int    // 总长度
	r       int    // next position to read
	w       int    // next position to write
	isEmpty bool
}

func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		bs:      make([][]byte, 2),
		buf:     make([]byte, size),
		size:    size,
		isEmpty: true,
	}
}

func (rb *RingBuffer) Peek(n int) (head []byte, tail []byte) {
	if rb.isEmpty {
		return
	}

	if n <= 0 { // 返回全部
		return rb.peekAll()
	}

	if rb.w > rb.r {
		m := rb.w - rb.r // 获取未读数据的长度
		if m > n {
			m = n
		}
		head = rb.buf[rb.r : rb.r+m]
		return
	}

	// 当读指针<写指针时，也就是说数据已经写至数组末端并且开始从数组头部继续写数据
	m := rb.size - rb.r + rb.w // 获取未读数据的长度
	if m > n {
		m = n
	}
	if rb.r+m <= rb.size {
		head = rb.buf[rb.r : rb.r+m]
	} else {
		c1 := rb.size - rb.r // 从读指针开始读取直到数组末端
		head = rb.buf[rb.r:]
		c2 := m - c1 // 计算差额，从头部开始读取
		tail = rb.buf[:c2]
	}
	return
}

// peekAll 读取全部
func (rb *RingBuffer) peekAll() (head []byte, tail []byte) {
	if rb.isEmpty {
		return
	}

	if rb.w > rb.r {
		head = rb.buf[rb.r:rb.w]
		return
	}

	head = rb.buf[rb.r:]
	if rb.w != 0 {
		tail = rb.buf[:rb.w]
	}
	return

}

func (rb *RingBuffer) Discard(n int) (discarded int, err error) {
	if n <= 0 {
		return 0, nil
	}
	discarded = rb.Buffered()
	if n < discarded {
		// 更新读指针位置
		rb.r = (rb.r + n) % rb.size
		return n, nil
	}

	rb.Reset()
	return
}

func (rb *RingBuffer) IsEmpty() bool {
	return rb.isEmpty
}

func (rb *RingBuffer) IsFull() bool {
	return rb.r == rb.w && !rb.isEmpty
}

func (rb *RingBuffer) Reset() {
	rb.isEmpty = true
	rb.r, rb.w = 0, 0
}

// Buffered 获取已缓存的数据长度
func (rb *RingBuffer) Buffered() int {
	if rb.w > rb.r {
		return rb.w - rb.r
	}

	// 环形
	return rb.size - rb.r + rb.w
}

func (rb *RingBuffer) Available() int {
	if rb.r == rb.w {
		if rb.isEmpty {
			return rb.size
		}
		return 0
	}

	if rb.w < rb.r {
		return rb.r - rb.w
	}

	return rb.size - rb.r + rb.w
}

func (rb *RingBuffer) Write(p []byte) (n int, err error) {
	n = len(p)
	if n == 0 {
		return
	}

	free := rb.Available()
	if n > free { // 判断是否需要扩容
		rb.grow(rb.size + n - free)
	}

	if rb.w >= rb.r {
		c1 := rb.size - rb.w // 计算直至数组末端还剩余的空间
		if c1 >= n {         // 空间足够，直接写入
			copy(rb.buf[rb.w:], p)
			rb.w += n
		} else { // 分两段拷贝
			copy(rb.buf[rb.w:], p[:c1])
			copy(rb.buf, p[c1:])
			rb.w = n - c1
		}
	} else {
		copy(rb.buf[rb.w:], p)
		rb.w += n
	}
	if rb.w == rb.size {
		rb.w = 0
	}
	rb.isEmpty = false
	return

}

// CeilToPowerOfTwo returns the least power of two integer value greater than
// or equal to n.
func CeilToPowerOfTwo(n int) int {

	if n <= 2 {
		return 2
	}
	n--
	n = fillBits(n)
	n++
	return n
}

// FloorToPowerOfTwo returns the greatest power of two integer value less than
// or equal to n.
func FloorToPowerOfTwo(n int) int {
	if n <= 2 {
		return 2
	}
	n = fillBits(n)
	n >>= 1
	n++
	return n
}

func fillBits(n int) int {
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	return n
}

// grow 扩容
func (rb *RingBuffer) grow(newCap int) {
	n := rb.size
	if n == 0 {
		if newCap <= 1024 {
			newCap = 1024
		} else {
			newCap = CeilToPowerOfTwo(newCap)
		}
	} else {
		doubleCap := n + n
		if newCap <= doubleCap {
			if n < 4096 {
				newCap = doubleCap
			} else {
				for 0 < n && n < newCap {
					n += n / 4
				}

				if n > 0 {
					newCap = n
				}
			}
		}
	}

	newBuf := bsPool.Get(newCap) // 取一个指定容量的数组
	oldLen := rb.Buffered()      // 获取旧数组的长度
	_, _ = rb.Read(newBuf)       // 将数据写入新数组
	bsPool.Put(rb.buf)           // 回收旧数组
	rb.buf = newBuf              // 赋值
	rb.r = 0                     // 从0开始读取
	rb.w = oldLen                // 标记读取位置
	rb.size = newCap             // 更新
	if rb.w > 0 {                // 判断是否为空
		rb.isEmpty = true
	}
}

// Read 读取数据
func (rb *RingBuffer) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return
	}

	if rb.isEmpty {
		return
	}

	if rb.w > rb.r { // 数据未写至数组末端
		n = rb.w - rb.r
		if n > len(p) {
			n = len(p)
		}
		copy(p, rb.buf[rb.r:rb.r+n])
		rb.r += n         // 更新读指针
		if rb.r == rb.w { // 判断是否满足重置的状态
			rb.Reset()
		}
		return
	}

	n = rb.size - rb.r + rb.w
	if n > len(p) {
		n = len(p)
	}

	if rb.r+n <= rb.size { // 需要的数据量不会超过数组末端
		copy(p, rb.buf[rb.r:rb.r+n])
	} else { // 先读取到数组末端，再从0开始读取
		c1 := rb.size - rb.r
		copy(p, rb.buf[rb.r:])
		c2 := n - c1
		copy(p[c1:], rb.buf[:c2])
	}
	rb.r = (rb.r + n) % rb.size
	if rb.r == rb.w {
		rb.Reset()
	}

	return
}

func (rb *RingBuffer) CopyFromSocket(fd int) (n int, err error) {
	if rb.r == rb.w {
		if !rb.isEmpty {
			rb.grow(rb.size + rb.size/2)
			n, err = syscall.Read(fd, rb.buf[rb.w:])
			if n > 0 {
				rb.w = (rb.w + n) % rb.size
			}
			return
		}
		rb.r, rb.w = 0, 0
		n, err = syscall.Read(fd, rb.buf)
		if n > 0 {
			rb.w = (rb.w + n) % rb.size
			rb.isEmpty = false
		}
		return
	}
	if rb.w < rb.r {
		n, err = syscall.Read(fd, rb.buf[rb.w:rb.r])
		if n > 0 {
			rb.w = (rb.w + n) % rb.size
		}
		return
	}

	rb.bs[0] = rb.buf[rb.w:]
	rb.bs[1] = rb.buf[:rb.r]
	n, err = unix.Readv(fd, rb.bs)
	if n > 0 {
		rb.w = (rb.w + n) % rb.size
	}

	return
}

func (rb *RingBuffer) ReadFrom(r io.Reader) (n int64, err error) {
	var m int
	for {
		if rb.Available() < 512 {
			rb.grow(rb.Buffered() + 512)
		}

		if rb.w >= rb.r {
			m, err = r.Read(rb.buf[rb.w:])
			if m < 0 {
				panic("RingBuffer.ReadFrom: reader returned negative count from Read")
			}
			rb.isEmpty = false
			rb.w = (rb.w + m) % rb.size
			n += int64(m)
			if err == io.EOF {
				return n, nil
			}
			if err != nil {
				return
			}
			m, err = r.Read(rb.buf[:rb.r])
			if m < 0 {
				panic("RingBuffer.ReadFrom: reader returned negative count from Read")
			}
			rb.w = (rb.w + m) % rb.size
			n += int64(m)
			if err == io.EOF {
				return n, nil
			}
			if err != nil {
				return
			}
		} else {
			m, err = r.Read(rb.buf[rb.w:rb.r])
			if m < 0 {
				panic("RingBuffer.ReadFrom: reader returned negative count from Read")
			}
			rb.isEmpty = false
			rb.w = (rb.w + m) % rb.size
			n += int64(m)
			if err == io.EOF {
				return n, nil
			}
			if err != nil {
				return
			}
		}
	}
}
