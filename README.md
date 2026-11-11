# GO_weifuwu

一个围绕 **RPC（远程过程调用）** 展开的 Go 学习仓库。目标是把「像调用本地函数一样调用远端函数」这件事从底层原理到工业级方案完整走一遍。

## 这个项目在做什么

RPC 的核心问题是三件事：**怎么找到远端方法**、**参数怎么序列化过网络**、**连接怎么管理**。仓库里的三个模块按由浅入深的顺序，分别给出不同答案：

| 模块 | 方案 | 解决的重点 |
|---|---|---|
| `rpc/` | Go 标准库 `net/rpc` | RPC 的最小骨架：服务注册、反射调用、同步/异步 |
| `rpc/` + JSON-RPC | `net/rpc/jsonrpc` 编解码器 | **协议与语言解耦**：换成通用编码后，Python 也能调 Go 服务 |
| `grpc/demo_hello/` | gRPC + Protocol Buffers | 工业级方案：IDL 定义契约、代码生成、HTTP/2、强类型 |
| `grpc/test_add/` | gRPC + Protocol Buffers | 换个数据类型和多参数场景巩固，理解 `int32` 与 Go `int` 的边界 |

演进逻辑很清晰：标准库 `net/rpc` 用的是 Go 专有的 `gob` 编码，只有 Go 客户端能调；把编解码换成 JSON-RPC 之后，跨语言就通了（`client.py` 就是证明）；但 JSON 是无类型的，字段拼错只能运行时发现 —— 于是引入 Protocol Buffers，用 `.proto` 文件把接口契约固化下来，编译期就能校验。gRPC 在此之上再补上 HTTP/2 多路复用、流式调用、超时传播和拦截器。

## 目录结构

```
.
├── grpc/
│   ├── demo_hello/            gRPC 入门：字符串参数的 SayHello
│   │   ├── hello_service/     服务端（module hello_service）
│   │   └── hello_client/      客户端（module hello_client）
│   └── test_add/              gRPC 练习：整数求和 Add
│       ├── server/            服务端（module server）
│       └── client/            客户端（module client）
└── rpc/                       Go 标准库 net/rpc + JSON-RPC
    ├── Service/service.go     服务端：注册 ServiceA，暴露 Add 方法
    └── Client/
        ├── client.go          Go 客户端：同步 Call + 异步 Go
        └── client.py          Python 客户端：手写 JSON-RPC 报文
```

## 环境要求

| 项 | 版本 | 用途 |
|---|---|---|
| Go | 1.27.4 | 全部模块 |
| google.golang.org/grpc | v1.85.0 | `grpc/` 下两个示例 |
| google.golang.org/protobuf | v1.36.12 | `grpc/` 下两个示例 |
| protoc | v7.36.2 | 仅重新生成 pb 代码时需要 |
| protoc-gen-go / protoc-gen-go-grpc | v1.6.2 | 同上 |
| Python 3 | 任意 | 仅 `rpc/Client/client.py` |

`rpc/` 目录只用 Go 标准库，零第三方依赖。

> **重要**：`grpc/` 下的四个子目录是**各自独立的 Go module**，必须 `cd` 进对应目录再执行 `go run`，在仓库根目录下直接跑会报找不到包。`rpc/` 没有 `go.mod`，属于游离示例文件，在根目录用 `go run ./rpc/xxx.go` 即可。

## 快速开始

### 1. rpc/ —— 标准库 net/rpc

服务端把 `ServiceA` 注册到 `net/rpc`，暴露 `Add(args *Args, reply *int) error`，监听 TCP `:9091`，编解码走 JSON-RPC。

```bash
# 终端 1：启动服务端
go run ./rpc/Service/service.go

# 终端 2：Go 客户端
go run ./rpc/Client/client.go
# 输出 ServiceA.Add: 10+20=30，随后演示异步调用的结果

# 或者用 Python 客户端
python3 ./rpc/Client/client.py
```

Go 客户端里能看到 RPC 的两种调用形态：

- **同步** `client.Call("ServiceA.Add", args, &reply)` —— 阻塞直到结果返回
- **异步** `client.Go(...)` —— 立即返回 `*rpc.Call`，从 `divCall.Done` 这个 channel 收结果

方法名 `"ServiceA.Add"` 是 `包内注册名.方法名` 的字符串，服务端靠反射找到它。这也是标准库方案的脆弱点：写错了要到运行时才知道。

Python 客户端手写的报文，就是 JSON-RPC 在线上的真实样子：

```json
{"id": 0, "params": [{"x": 10, "y": 20}], "method": "ServiceA.Add"}
```

`params` 里的字段名要能对应上 Go 的 `Args` 结构体（大小写不敏感）。这一步说明：**只要双方约定好编码格式，客户端和服务端用什么语言写完全无所谓。**

### 2. grpc/demo_hello/ —— gRPC 入门

最小可运行的 gRPC 示例。契约写在 `hello.proto` 里：

```protobuf
service Greeter {
    rpc SayHello (HelloRequest) returns (HelloResponse);
}

message HelloRequest  { string name  = 1; }
message HelloResponse { string reply = 1; }
```

protoc 据此生成服务端接口和客户端桩代码，业务代码只需要实现接口 / 调用桩方法，序列化和网络传输全由框架接管。

```bash
# 终端 1：启动服务端（监听 :8972）
cd grpc/demo_hello/hello_service
go run .

# 终端 2：启动客户端，-name 指定名字，默认 "max"
cd grpc/demo_hello/hello_client
go run . -name alice
```

客户端为每次调用设置了 1 秒超时（`context.WithTimeout`），服务端没起来或响应过慢会返回 `DeadlineExceeded` 而不是一直阻塞 —— 这是 gRPC 相比裸 socket 的实用之处，超时能随 ctx 一路传到服务端。

传输用的是 `insecure` 明文凭证，仅限本地调试，生产环境要换成 TLS。

### 3. grpc/test_add/ —— gRPC 求和练习

结构和 demo_hello 一致，换成数值类型和双参数，用来熟悉 proto 的 `int32` 与 Go 类型的对应关系：

```protobuf
service Greeter {
    rpc Add (AddRequest) returns (AddResponse);
}

message AddRequest  { int32 x = 1; int32 y = 2; }
message AddResponse { int32 reply = 1; }
```

```bash
# 终端 1：启动服务端（监听 :8972）
cd grpc/test_add/server
go run .

# 终端 2：启动客户端
cd grpc/test_add/client
go run .                      # 默认 1 + 2 = 3
go run . -x 3 -y 4            # 3 + 4 = 7
go run . -x=-10 -y=5          # -10 + 5 = -5
```

实测输出：

| 输入 | 输出 |
|---|---|
| 无参数（默认 1, 2） | `1 + 2 = 3` |
| `-x 3 -y 4` | `3 + 4 = 7` |
| `-x=-10 -y=5` | `-10 + 5 = -5` |
| `-x 2000000000 -y 100000000` | `2000000000 + 100000000 = 2100000000` |

这个示例里有两个值得注意的语言边界：

**proto `int32` ≠ Go `int`。** 生成的字段类型是 `int32`，而 `flag.Int` 给的是 `int`，Go 不做隐式数值转换，构造请求时必须显式写 `int32(*x)`。

**负数参数要用等号。** `-x -10` 会被 `flag` 包理解成「`-x` 没带值，`-10` 是下一个参数」，得写成 `-x=-10`。

顺带一提，`int32` 相加超过 2147483647 会静默回绕成负数，服务端不做检查。上面最后一组用例接近上限但没超，所以结果正确。

## 关于 .proto 的两个易混概念

写 gRPC 时最容易搞混的是 `package` 和 `go_package`，它们作用在完全不同的层面：

- **`package pb;`** 决定 RPC 方法在线上的全限定名，即 `/pb.Greeter/Add`。**客户端和服务端必须完全一致**，否则调用会返回 `Unimplemented` —— 这是跨端联调时最常见的坑。
- **`option go_package`** 只决定生成代码在 Go 里的导入路径，两端各自独立（`server/pb` 和 `client/pb`），改它不影响协议兼容性。

本仓库的做法是给 client 和 server 各存一份 `.proto`，只改 `go_package`，其余保持完全一致。

### 重新生成 pb 代码

改了 `.proto` 之后需要重新生成，在对应 module 目录下执行（以 test_add 的 client 为例）：

```bash
cd grpc/test_add/client
protoc --go_out=. --go_opt=module=client \
       --go-grpc_out=. --go-grpc_opt=module=client \
       pb/add.proto
```

`--go_opt=module=client` 的作用是把 `go_package` 里的模块前缀剥掉，让文件落在 `pb/` 而不是 `client/pb/`。

demo_hello 同理，把 `module=client` 换成 `module=hello_client`、proto 路径换成 `proto/hello.proto`。

生成出来的 `.pb.go` 文件头部会记录 protoc 和插件的版本，方便日后对齐环境。

## 端口占用

| 端口 | 使用方 |
|---|---|
| `:8972` | `grpc/demo_hello`、`grpc/test_add` |
| `:9091` | `rpc/` |

两个 gRPC 示例共用 `:8972`，**不能同时启动**，否则后启动的会报 `address already in use`。`rpc/` 用 `:9091`，和 gRPC 示例互不冲突，可以并行跑。

## 说明

这是学习用的练习仓库，代码以「把概念讲清楚」为优先，没有做生产级的错误处理、连接池、优雅退出和并发控制。比如 `rpc/Service/service.go` 是在主 goroutine 里串行处理连接的，一个客户端不断开后面的就排不上；gRPC 示例也都用的明文凭证。参考时请注意这一点。

各 module 目录下 `go build` 产生的可执行文件已通过 `.gitignore` 排除。
