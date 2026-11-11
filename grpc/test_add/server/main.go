package main

import (
	"context"
	"fmt"
	"net"
	"server/pb"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedGreeterServer
}

func (s *server) Add(ctx context.Context, in *pb.AddRequest) (*pb.AddResponse, error) {
	reply := in.GetX() + in.GetY()
	return &pb.AddResponse{Reply: reply}, nil
}

func main() {
	// 监听 TCP 8972 端口，等待 gRPC 客户端连接
	l, err := net.Listen("tcp", ":8972")
	if err != nil {
		fmt.Printf("failed net.Listen, err:%v\n", err)
		return
	}
	// 创建 gRPC 服务器实例
	s := grpc.NewServer()
	// 将 service 注册为 Greeter 服务的实现，处理对应的 rpc 请求
	pb.RegisterGreeterServer(s, &server{})
	// 阻塞运行服务器；返回错误说明服务异常退出
	err = s.Serve(l)
	if err != nil {
		fmt.Printf("failed s.Serve(l), err:%v\n", err)
		return
	}
}
