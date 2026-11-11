package main

import (
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
)

type Args struct {
	X, Y int
}

type ServiceA struct{}

func (s *ServiceA) Add(args *Args, reply *int) error {
	*reply = args.X + args.Y
	return nil
}

func main() {
	service := new(ServiceA)
	rpc.Register(service)
	// rpc.HandleHTTP()
	l, err := net.Listen("tcp", ":9091")
	if err != nil {
		log.Fatal("Listen Fatal", err)
	}
	// http.Serve(l, nil)
	// 基于TCP 处理服务
	for {
		conn, err := l.Accept()
		if err != nil {
			// Accept 报错说明监听 socket 已不可用（被关闭或 fd 耗尽）。
			// 这里必须退出而不是 continue：忽略错误继续循环会在不接受任何连接的情况下
			// 空转，把 CPU 打满。交给 systemd/docker 之类的进程管理器重启更合适。
			log.Fatal("Accept Fatal", err)
		}
		//rpc.ServeConn(conn)

		// 使用json协议。
		// ServeCodec 会阻塞到该连接关闭，所以必须放到独立 goroutine 里，
		// 否则 Accept 循环无法继续，一个客户端不断开后面的连接就全都排不上。
		go rpc.ServeCodec(jsonrpc.NewServerCodec(conn))
	}
}
