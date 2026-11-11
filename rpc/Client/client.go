// rpc demo/client.go

package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
)

type Args struct {
	X, Y int
}

func main() {
	// 建立HTTP连接
	// client, err := rpc.DialHTTP("tcp", "127.0.0.1:9091")
	// client, err := rpc.Dial("tcp", "127.0.0.1:9091")

	// 使用json协议
	conn, err := net.Dial("tcp", "127.0.0.1:9091")
	// 必须先判 err 再用 conn：连接失败时 conn 为 nil，
	// 直接传给 NewClientCodec 会先崩在这里，下面的 log.Fatal 根本走不到
	if err != nil {
		log.Fatal("dialing:", err)
	}
	client := rpc.NewClientWithCodec(jsonrpc.NewClientCodec(conn))

	// 同步调用
	args := &Args{10, 20}
	var reply int
	err = client.Call("ServiceA.Add", args, &reply)
	if err != nil {
		log.Fatal("ServiceA.Add error:", err)
	}
	fmt.Printf("ServiceA.Add: %d+%d=%d\n", args.X, args.Y, reply)

	// 异步调用
	var reply2 int
	divCall := client.Go("ServiceA.Add", args, &reply2, nil)
	replyCall := <-divCall.Done // 接收调用结果
	fmt.Println(replyCall.Error)
	fmt.Println(reply2)
}
