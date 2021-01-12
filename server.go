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
