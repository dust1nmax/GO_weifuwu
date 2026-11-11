package main

import (
	"context"
	"hello_client/proto"
	"log"
	"time"

	"google.golang.org/grpc"
	// insecure 提供不加密的传输凭证，本地调试用；生产环境应换成 TLS 凭证
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// 创建 gRPC 客户端连接。
	// grpc.NewClient 是懒连接：这里只建立 ClientConn 对象，真正的首次 RPC 调用时才会去拨号。
	// 第二个参数指定传输凭证，insecure 表示明文传输（不走 TLS）。
	conn, err := grpc.NewClient("127.0.0.1:8972", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		// log.Fatalf 打印日志后直接以非 0 状态码退出进程，无需再 return
		log.Fatalf("grpc.NewClient failed, err:%v\n", err)
	}
	// 函数返回前关闭连接，释放底层的 TCP 连接和 goroutine
	defer conn.Close()

	// 用 protoc 生成的桩代码把连接包装成 Greeter 服务的客户端，
	// 之后就能像调用本地方法一样调用远端的 SayHello。
	c := proto.NewGreeterClient(conn)

	// 为本次调用设置 1 秒超时：超时或服务端未响应时，RPC 会返回 DeadlineExceeded 错误，
	// 避免客户端一直阻塞。cancel 必须调用，否则计时器资源不会及时释放。
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 构造请求参数并发起一元（unary）RPC，阻塞等待服务端返回
	name := "max"
	response, err := c.SayHello(ctx, &proto.HelloRequest{Name: name})
	if err != nil {
		// 这里用 Printf 而不是 Fatalf：连接已经建立，调用失败不算致命错误，正常退出即可
		log.Printf("c.SayHello failed, err:%v\n", err)
		return
	}

	// GetReply 是 protoc 为字段生成的 getter，接收者为 nil 时也安全，不会 panic
	log.Printf("response:%v", response.GetReply())
}
