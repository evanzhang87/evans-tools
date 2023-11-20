package main

import (
	"flag"
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"strings"
	"time"
)

var (
	addr string
	op   string
	path string
)

func init() {
	flag.StringVar(&addr, "addr", "127.0.0.1:2181", "zk addr")
	flag.StringVar(&op, "op", "get", "option")
	flag.StringVar(&path, "path", "/", "zkpath")
}

func main() {
	flag.Parse()
	hosts := strings.Split(addr, ",")
	conn, _, err := zk.Connect(hosts, time.Second*5)
	if err != nil {
		fmt.Println(err)
		return
	}

	switch op {
	case "get":
		data, _, err := conn.Get(path)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(data))
	case "watch":
		for {
			data, _, ch, err := conn.GetW(path)
			if err != nil {
				fmt.Println("err", err)
				time.Sleep(time.Second * 10)
				continue
			}
			fmt.Println("init data:", string(data))
			for event := range ch {
				fmt.Println(event.Type.String())
				data, _, err := conn.Get(path)
				if err != nil {
					fmt.Println(err)
					break
				}
				fmt.Println(string(data))
			}
		}
	default:
		fmt.Println("unexpected op ", op)
		return
	}
}
