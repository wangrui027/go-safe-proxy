# golang 安全代理程序

## 服务端

### 功能

服务端默认启动`43088`端口，接收客户端的请求，解析加密后的url参数，拿到url参数后使用AES算法对其解密，得到客户端希望请求的真实http地址

由服务端发起真正的url地址请求，得到返回结果，将返回结果使用AES算法加密后返回给客户端

### 文档

```bash
# server.exe -help
Usage of server.exe:
  -key string
        AES密钥 (default "1234567890abcdef")
  -p int
        服务端口号 (default 43088)
```

## 客户端

### 功能

客户端接受服务端地址、AES密钥和真实的url访问地址（使用当前的AES密钥加密后并做Base64字符串编码）三个参数

客户端将真实的url地址使用AES算法加密后作为参数向服务端发起请求，得到返回后使用AES算法解密返回结果得到真实的响应数据

### 文档

```bash
# client.exe -help
Usage of client.exe:
  -key string
        AES密钥 (default "1234567890abcdef")
  -server string
        服务端地址
  -url string
        请求访问地址，使用当前的AES密钥加密后并做Base64编码
```

## 编译

由于我的服务端最终是要部署到路由器上使用的，路由器型号网件R6900，路由器固件梅林380.70_0-X7.9.1，所以要编译为arm架构程序

```go
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=5 go build server.go
```

至于客户端，当然是本地运行了，编译为x86结构就OK了

```go
go build client.go
```

