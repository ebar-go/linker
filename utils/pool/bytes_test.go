package pool

import (
	"fmt"
	"testing"
)

func BenchmarkByteSlice(b *testing.B) {
	b.Run("Run.N", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			bs := Get(1024)
			Put(bs)
		}
	})
	b.Run("Run.Parallel", func(b *testing.B) {
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				bs := Get(1024)
				Put(bs)
			}
		})
	})

	bp := NewBufferPool()
	b.Run("Run.N.BufferPool", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			bs := bp.Get()
			bp.Put(bs)
		}
	})
	b.Run("Run.Parallel.BufferPool", func(b *testing.B) {
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				bs := bp.Get()
				bp.Put(bs)
			}
		})
	})
}

func TestByteSlice(t *testing.T) {
	s := make([]byte, 1024)
	copy(s, "hello")
	a := s[:0]
	fmt.Printf("s = %p, %v\n", s, s)
	fmt.Printf("a = %p\n", a)

	s = s[2:4]
	fmt.Printf("s = %p, %v\n", s, s)
}
