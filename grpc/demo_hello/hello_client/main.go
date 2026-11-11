package main

import (
	"context"
	"flag"
	"hello_client/proto"
	"log"
	"time"

	"google.golang.org/grpc"
	// insecure 提供不加密的传输凭证，本地调试用；生产环境应换成 TLS 凭证
	"google.golang.org/grpc/credentials/insecure"
)

// name 指向命令行参数 -name 的存储地址。
// flag.String 返回的是 *string 而不是 string：参数值要等 flag.Parse() 执行后才写入，
// 所以这里只能先拿到指针，读取时必须解引用（*name）。
// 三个参数依次是：flag 名、默认值（不传 -name 时使用）、-h 输出里的说明文字。
// 放在包级变量是因为 flag.String 会在 init 阶段完成注册，早于 main 执行。
var name = flag.String("name", "max", "通过 -name 告诉你是谁")

func main() {
	// 解析 os.Args，把命令行上的 -name=xxx 写进上面注册的指针。
	// 必须在读取任何 flag 之前调用，否则拿到的永远是默认值。
	// 遇到 -h/--help 会自动打印用法并退出；遇到未定义的 flag 会报错退出。
	flag.Parse()
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

	// 构造请求参数并发起一元（unary）RPC，阻塞等待服务端返回。
	// *name 是对 flag 指针解引用，取出解析后的字符串值；
	// HelloRequest 的 Name 字段是 string，所以这里不能直接传 name。
	response, err := c.SayHello(ctx, &proto.HelloRequest{Name: *name})
	if err != nil {
		// 这里用 Printf 而不是 Fatalf：连接已经建立，调用失败不算致命错误，正常退出即可
		log.Printf("c.SayHello failed, err:%v\n", err)
		return
	}

	// GetReply 是 protoc 为字段生成的 getter，接收者为 nil 时也安全，不会 panic
	log.Printf("response:%v", response.GetReply())
}
