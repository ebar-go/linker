package linker

import (
	"context"
	"fmt"
	"github.com/rcrowley/go-metrics"
	"linker/pkg/system"
	"log"
	"math/rand"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"testing"
	"time"
)

var (
	addr = "0.0.0.0:8086"
)

func startPprof() {
	http.ListenAndServe("0.0.0.0:6060", nil)
}
func TestReactor(t *testing.T) {
	go startPprof()
	var connected int
	go func() {
		for {
			time.Sleep(time.Second * 5)

			fmt.Printf("current connection:%d, memory usage: %.3f MB\n", connected, float64(system.GetMem())/1024/1024)
		}

	}()
	//system.SetLimit()
	reactor := NewReactor(WithProcessor(32), WithContextPoolSize(32))

	reactor.OnConnect(func(conn Conn) {
		connected++
		log.Println("connected")
	})

	reactor.OnDisconnect(func(conn Conn) {
		connected--
		log.Println("disconnected", conn.FD())
	})
	reactor.OnRequest(func(ctx *Context) {
		//log.Println("receive:", string(ctx.Body()))
		ctx.Conn().Push([]byte("hello"))
	})

	if err := reactor.Run("tcp", addr); err != nil {
		t.Fatal(err)
	}
}

func TestClient(t *testing.T) {
	c, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			bytes := make([]byte, 100)
			n, err := c.Read(bytes)
			if err != nil {
				panic(err)

			}
			if n > 0 {
				log.Println("receive:", string(bytes[:n]))
			}

		}

	}()

	for i := 0; i < 10; i++ {

		if _, err := c.Write([]byte("hello,world\n")); err != nil {
			panic(err)
		}
		time.Sleep(time.Second * 1)
	}
}

func BenchmarkClient(b *testing.B) {
	system.SetLimit()
	opsRate := metrics.NewRegisteredTimer("ops", nil)

	ch := make(chan net.Conn, 50)
	n := 10000
	ctx, cancel := context.WithCancel(context.Background())
	for i := 0; i < 10; i++ {
		go func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					c, err := net.DialTimeout("tcp", addr, 10*time.Second)
					if err == nil {
						ch <- c
					}
				}
			}

		}(ctx)
	}
	connections := make([]net.Conn, 0, 1024)
	for len(connections) < n {
		connections = append(connections, <-ch)
	}
	cancel()

	go func() {
		metrics.Log(metrics.DefaultRegistry, 5*time.Second, log.New(os.Stderr, "metrics: ", log.Lmicroseconds))
	}()

	b.ResetTimer()
	for i := 0; i < 100; i++ {
		go func() {
			for {
				n := rand.Intn(len(connections) - 1)
				c := connections[n]
				before := time.Now()
				if _, err := c.Write([]byte("hello\n")); err != nil {
					_ = c.Close()
					log.Println(err)
				} else {
					opsRate.Update(time.Now().Sub(before))
				}
			}

		}()
	}
	select {}
}
