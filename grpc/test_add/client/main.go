package main

import (
	"client/pb"
	"context"
	"flag"
	"log"
	"time"

	"google.golang.org/grpc"
	// insecure 提供不加密的传输凭证，本地调试用；生产环境应换成 TLS 凭证
	"google.golang.org/grpc/credentials/insecure"
)

// 命令行参数。flag.Int 返回的是 *int，要等 flag.Parse() 之后才写入，读取时需解引用。
// proto 里的 x/y 是 int32，而 flag 包没有 Int32，所以调用前要显式转换。
var (
	x = flag.Int("x", 1, "加数 x")
	y = flag.Int("y", 2, "加数 y")
)

func main() {
	// 解析命令行参数，必须在读取 *x / *y 之前调用
	flag.Parse()

	// 创建 gRPC 客户端连接。
	// grpc.NewClient 是懒连接：这里只建立 ClientConn 对象，首次 RPC 调用时才真正拨号。
	conn, err := grpc.NewClient("127.0.0.1:8972", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		// log.Fatalf 打印日志后直接以非 0 状态码退出进程，无需再 return
		log.Fatalf("grpc.NewClient failed, err:%v\n", err)
	}
	// 函数返回前关闭连接，释放底层的 TCP 连接和 goroutine
	defer conn.Close()

	// 用 protoc 生成的桩代码把连接包装成 Greeter 服务的客户端，
	// 之后就能像调用本地方法一样调用远端的 Add。
	c := pb.NewGreeterClient(conn)

	// 为本次调用设置 1 秒超时：超时或服务端未响应时 RPC 返回 DeadlineExceeded，
	// 避免客户端一直阻塞。cancel 必须调用，否则计时器资源不会及时释放。
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 发起一元（unary）RPC，阻塞等待服务端返回。
	// flag.Int 给出的是 int，而 proto 字段是 int32，Go 不做隐式转换，必须显式 int32()。
	// 返回值是 *pb.AddResponse 整个响应对象，不是求和结果本身。
	response, err := c.Add(ctx, &pb.AddRequest{X: int32(*x), Y: int32(*y)})
	if err != nil {
		// 用 Printf 而不是 Fatalf：连接已建立，调用失败不算致命错误，正常退出即可
		log.Printf("add failed, err:%v\n", err)
		return
	}

	// GetReply() 才是 int32 类型的求和结果；直接打印 response 会输出结构体指针
	log.Printf("%d + %d = %d", *x, *y, response.GetReply())
}
