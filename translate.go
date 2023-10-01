//go:build ignore
// +build ignore

package main

import (
	"fmt"

	translator "github.com/Conight/go-googletrans"
)

var content = `你好，世界！`

func main() {
	c := translator.Config{
		Proxy:       "socks5://zxwy:2082327995@192.168.10.12:1080",
		UserAgent:   []string{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36"},
		ServiceUrls: []string{"translate.google.com.hk"},
	}
	t := translator.New(c)
	result, err := t.Translate(content, "auto", "en")
	if err != nil {
		panic(err)
	}
	fmt.Println(result.Text)
}
