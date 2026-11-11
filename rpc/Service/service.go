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
		conn, _ := l.Accept()
		//rpc.ServeConn(conn)

		// 使用json协议
		rpc.ServeCodec(jsonrpc.NewServerCodec(conn))
	}
}
