/*
	1.编写protobuf文件
	2.生成代码
	3.编写业务逻辑
*/
package main

import (
	"context"
	"fmt"
	"hello_service/proto" // protoc 生成的代码，包含 Greeter 服务接口和消息结构体
	"net"

	"google.golang.org/grpc"
)

// service 实现 proto 中定义的 Greeter 服务
type service struct {
	// 嵌入默认实现，自动满足接口中未实现的方法（向前兼容新增的 rpc）
	proto.UnimplementedGreeterServer
}

// SayHello 实现 Greeter 服务的 SayHello 方法，由 gRPC 框架在收到请求时调用
func (s *service) SayHello(ctx context.Context, in *proto.HelloRequest) (*proto.HelloResponse, error) {
	// 拼接问候语，GetName() 是 protoc 为 name 字段生成的取值方法
	reply := "hello " + in.GetName()
	// 返回响应消息，Reply 对应 proto 中 HelloResponse 的 reply 字段
	return &proto.HelloResponse{Reply: reply}, nil
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
	proto.RegisterGreeterServer(s, &service{})
	// 阻塞运行服务器；返回错误说明服务异常退出
	err = s.Serve(l)
	if err != nil {
		fmt.Printf("failed s.Serve(l), err:%v\n", err)
		return
	}
}
