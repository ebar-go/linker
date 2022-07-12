package linker

import (
	"errors"
	"github.com/gorilla/websocket"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type WebsocketServer struct {
	components

	conf    *Config
	upgrade *websocket.Upgrader

	requestURI string // websocket需要一个路由地址
}

func (s *WebsocketServer) Start() error {
	s.upgrade = &websocket.Upgrader{
		ReadBufferSize:  s.conf.Rcvbuf,
		WriteBufferSize: s.conf.Sndbuf,
	}
	s.engine.Use(s.OnRequest)
	http.HandleFunc(s.requestURI, s.handle)
	for _, bind := range s.conf.Bind {
		log.Printf("start websocket listen: %s\n", bind)
		go func(addr string) {
			if err := http.ListenAndServe(addr, nil); err != nil {
				panic(err)
			}
		}(bind)
	}
	return nil

}

// upgrade 升级协议，返回socket连接
func (s *WebsocketServer) upgradeRequest(w http.ResponseWriter, r *http.Request) (conn *wsConn, err error) {
	ws, err := s.upgrade.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	conn = &wsConn{
		Conn:    ws,
		msgType: websocket.BinaryMessage,
	}
	return
}

func (s *WebsocketServer) handle(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgradeRequest(w, r)
	if err != nil {
		log.Printf("upgrade failed:%v\n", err)
		return
	}

	if s.conf.Debug {
		lAddr := conn.LocalAddr().String()
		rAddr := conn.RemoteAddr().String()
		log.Printf("start serve %s with %s\n", lAddr, rAddr)
	}

	connection := NewConn(conn, s.conf.QueueSize) // 实例化会话
	defer connection.Close()

	s.OnConnect(connection)

	// 分发响应数据
	go connection.dispatchResponse()

	// 处理接收数据
	connection.handleRequest(s.engine.contextPool(rand.Intn(10000)), conn.ReadLine)

	s.OnDisconnect(connection)
}

// wsConn 实现net.Conn接口
type wsConn struct {
	*websocket.Conn
	msgType int
}

func (conn *wsConn) ReadLine() ([]byte, error) {
	_, msg, err := conn.ReadMessage()
	return msg, err
}

func (conn *wsConn) Read(b []byte) (n int, err error) {
	msgType, msg, err := conn.ReadMessage()
	if err != nil {
		return 0, err
	}
	if msgType != conn.msgType {
		return 0, errors.New("invalid message type")
	}

	b = msg
	return len(msg), nil
}

func (conn *wsConn) Write(b []byte) (n int, err error) {
	err = conn.WriteMessage(conn.msgType, b)
	n = len(b)
	return
}

func (conn *wsConn) SetDeadline(t time.Time) error {
	return conn.UnderlyingConn().SetDeadline(t)
}
