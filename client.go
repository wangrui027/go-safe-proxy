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
