package main

import (
	"context"
	"fmt"
	"math"
	"net"
	"server/pb"

	"google.golang.org/grpc"
	// codes/status 用于返回带标准状态码的 gRPC 错误，客户端可用 status.Code(err) 判断
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	pb.UnimplementedGreeterServer
}

func (s *server) Add(ctx context.Context, in *pb.AddRequest) (*pb.AddResponse, error) {
	x, y := in.GetX(), in.GetY()
	// int32 相加溢出会静默回绕（大正数变负数），Go 既不报错也不 panic，
	// 只能在计算前判断：y 为正时看是否超上限，y 为负时看是否低于下限，y 为 0 不可能溢出。
	// 这里返回 InvalidArgument 而不是自己截断，是为了让客户端明确知道结果不可信。
	if (y > 0 && x > math.MaxInt32-y) || (y < 0 && x < math.MinInt32-y) {
		return nil, status.Errorf(codes.InvalidArgument, "result overflows int32: %d + %d", x, y)
	}
	return &pb.AddResponse{Reply: x + y}, nil
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
