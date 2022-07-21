package linker

import (
	"bufio"
	uuid "github.com/satori/go.uuid"
	"linker/utils/binary"
	"linker/utils/pool"
	"net"
	"sync"
)

// Conn 客户端连接实例
type Conn struct {
	instance net.Conn

	uuid     string // 唯一ID
	remoteIP string // 客户端IP

	sendQueue  chan []byte // 发送队列
	once       *sync.Once
	done       chan struct{} // 关闭标识
	workerPool *pool.WorkerPool
}

// NewConn 初始化连接
func NewConn(netConn net.Conn, queueSize int) *Conn {
	conn := &Conn{instance: netConn}
	conn.init(queueSize)
	return conn
}

// UUID 返回连接的唯一ID
func (conn *Conn) UUID() string {
	return conn.uuid
}

func (conn *Conn) init(queueSize int) {
	conn.uuid = uuid.NewV4().String()
	conn.sendQueue = make(chan []byte, queueSize)
	conn.workerPool = pool.NewWorkerPool(queueSize)
	conn.once = new(sync.Once)
	conn.done = make(chan struct{})
	conn.remoteIP, _, _ = net.SplitHostPort(conn.instance.RemoteAddr().String())
}

// IP 返回客户端IP
func (conn *Conn) IP() string {
	return conn.remoteIP
}

// Push 推送消息
func (conn *Conn) Push(msg []byte) {
	// 当入列速度大于出列速度，消息将被抛弃，需要合理设置队列长度
	// 如果队列已被关闭，则会panic
	select {
	case conn.sendQueue <- msg:
	default:
	}
}

// NetConn 连接
func (conn *Conn) NetConn() net.Conn {
	return conn.instance
}

func (conn *Conn) Write(msg []byte) error {
	_, err := conn.instance.Write(msg)
	return err
}

// Close 关闭请求
func (conn *Conn) Close() {
	conn.once.Do(func() {
		close(conn.done)
		//close(conn.sendQueue)
		_ = conn.instance.Close()
	})
}

// dispatchResponse 分发数据
func (conn *Conn) dispatchResponse() {
	defer conn.Close()

	var err error
	for {
		select {
		case <-conn.done:
			return
		default:
			msg, ok := <-conn.sendQueue
			if !ok { // 队列已关闭
				return
			}

			// 写数据
			if _, err = conn.instance.Write(msg); err != nil {
				return
			}
		}
	}
}

// handleRequest 处理客户端的请求
func (conn *Conn) handleRequest(contextPool *sync.Pool, reader func() ([]byte, error)) {
	var (
		err error
		msg []byte
	)

	for {
		msg, err = reader()
		if err != nil {
			return
		}
		ctx := contextPool.Get().(*Context)
		ctx.connection = conn
		ctx.request.body = msg
		ctx.Keys = nil

		// 利用协程池控制并发上限，在提高并发的同时，防止协程泄露
		conn.workerPool.Submit(ctx.Run)

		contextPool.Put(ctx)

	}

}

func (conn *Conn) newScanner(dataLength int) (scanner *bufio.Scanner) {
	scanner = bufio.NewScanner(conn.instance)
	if dataLength > 0 {
		scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			if !atEOF && len(data) > dataLength {
				length := int(binary.BigEndian.Int32(data[:dataLength]))
				if length <= len(data) {
					return length, data[:length], nil
				}
			}
			return
		})
	}
	return
}
