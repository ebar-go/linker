package linker

import "net"

type HandleFunc func(ctx IContext)

// IContext 上下文
type IContext interface {
	ContextParam

	Request() IRequest
	Channel() IConnection
	//Output(decoder Decoder)
	Output(msg []byte)
	Next()  // 执行下一步
	Abort() // 中断
}

type ContextParam interface {
	Set(key string, value interface{})
	Get(key string) (value interface{}, exists bool)
	GetInt(key string) (i int)
	GetString(key string) (s string)
}

// IRequest 请求
type IRequest interface {
	Body() []byte
	//Bind(decode Decoder) error // 解析结构体
}

// IConnection 连接
type IConnection interface {
	IP() string // 获取客户端IP
	//Push(encoder Encoder)    // 推送消息
	Push(msg []byte)
	Connection() net.Conn // 原始连接
	Close()               // 关闭
}

// IServer TCP服务,负责启动tcp服务
type IServer interface {
	Start() error //启动服务器方法
	Use(handlers ...HandleFunc)

	SetOnConnect(func(connection IConnection))    // 注册连接时的回调
	SetOnDisconnect(func(connection IConnection)) // 注册断开连接时的回调
	SetOnReceive(hookFunc HandleFunc)             // 注册receive回调
}
