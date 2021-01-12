# golang 安全代理程序

## 服务端

### 功能

服务端默认启动`43088`端口，接收客户端的请求，解析加密后的url参数，拿到url参数后使用AES算法对其解密，得到客户端希望请求的真实http地址

由服务端发起真正的url地址请求，得到返回结果，将返回结果使用AES算法加密后返回给客户端

### 源码

```GO
package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

var key []byte

func padding(src []byte, blocksize int) []byte {
	padnum := blocksize - len(src)%blocksize
	pad := bytes.Repeat([]byte{byte(padnum)}, padnum)
	return append(src, pad...)
}

func unpadding(src []byte) []byte {
	n := len(src)
	unpadnum := int(src[n-1])
	return src[:n-unpadnum]
}

func decryptAES(src []byte) []byte {
	block, _ := aes.NewCipher(key)
	blockmode := cipher.NewCBCDecrypter(block, key)
	blockmode.CryptBlocks(src, src)
	src = unpadding(src)
	return src
}

func encryptAES(src []byte) []byte {
	block, _ := aes.NewCipher(key)
	src = padding(src, block.BlockSize())
	blockmode := cipher.NewCBCEncrypter(block, key)
	blockmode.CryptBlocks(src, src)
	return src
}

func get(url string) string {
	res, err := http.Get(url)
	if err != nil {
		return "请求错误"
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return "请求错误"
	}
	return string(body)
}

// 先加密后编码
func index(w http.ResponseWriter, r *http.Request) {
	queryForm, err := url.ParseQuery(r.URL.RawQuery)
	if err == nil && len(queryForm["url"]) > 0 {
		url := queryForm["url"][0]
		bytes, _ := base64.URLEncoding.DecodeString(url)
		url = string(decryptAES(bytes))
		log.Println("url: " + url)
		response := get(url)
		result := encryptAES([]byte(response))
		fmt.Fprintf(w, base64.URLEncoding.EncodeToString(result))
	} else {
		fmt.Fprintf(w, "url缺失")
	}
}

func main() {
	var port int
	var keyText string

	flag.IntVar(&port, "p", 43088, "服务端口号")
	flag.StringVar(&keyText, "key", "1234567890abcdef", "AES密钥")
	flag.Parse()

	key = []byte(keyText)

	log.Println("服务端口号：" + strconv.Itoa(port))
	log.Println("AES密钥：" + keyText)

	http.HandleFunc("/", index)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}
```

### 文档

```bash
# server --help
Usage of server:
  -key string
        AES密钥 (default "1234567890abcdef")
  -p int
        服务端口号 (default 43088)
```

## 客户端

### 功能

客户端接受服务端地址、AES密钥和真实的url访问地址（使用当前的AES密钥加密后并做Base64字符串编码）三个参数

客户端将真实的url地址使用AES算法加密后作为参数向服务端发起请求，得到返回后使用AES算法解密返回结果得到真实的响应数据

### 源码

```go
package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
)

var key []byte

func padding(src []byte, blocksize int) []byte {
	padnum := blocksize - len(src)%blocksize
	pad := bytes.Repeat([]byte{byte(padnum)}, padnum)
	return append(src, pad...)
}

func unpadding(src []byte) []byte {
	n := len(src)
	unpadnum := int(src[n-1])
	return src[:n-unpadnum]
}

func decryptAES(src []byte) []byte {
	block, _ := aes.NewCipher(key)
	blockmode := cipher.NewCBCDecrypter(block, key)
	blockmode.CryptBlocks(src, src)
	src = unpadding(src)
	return src
}

func encryptAES(src []byte) []byte {
	block, _ := aes.NewCipher(key)
	src = padding(src, block.BlockSize())
	blockmode := cipher.NewCBCEncrypter(block, key)
	blockmode.CryptBlocks(src, src)
	return src
}

func get(url string) string {
	res, err := http.Get(url)
	if err != nil {
		return "请求错误"
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return "请求错误"
	}
	return string(body)
}

// 先解码后解密
func main() {
	var server string
	var keyText string
	var url string

	flag.StringVar(&server, "server", "", "服务端地址")
	flag.StringVar(&keyText, "key", "1234567890abcdef", "AES密钥")
	flag.StringVar(&url, "url", "", "需要服务端代理请求的访问地址，客户端后台会使用当前的AES密钥加密后并做Base64编码")
	flag.Parse()

	key = []byte(keyText)
	url = base64.URLEncoding.EncodeToString(encryptAES([]byte(url)))
	response := get(server + "?url=" + url)

	byte, _ := base64.URLEncoding.DecodeString(response)
	result := decryptAES(byte)
	fmt.Println(string(result))
}
```

### 文档

```bash
# client.exe --help
Usage of client.exe:
  -key string
        AES密钥 (default "1234567890abcdef")
  -server string
        服务端地址
  -url string
        需要服务端代理请求的访问地址，客户端后台会使用当前的AES密钥加密后并做Base64编码
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

