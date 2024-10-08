package main

import (
	"flag"
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"strings"
	"time"
)

var (
	src  string
	dst  string
	path string
)

func init() {
	flag.StringVar(&src, "src", "127.0.0.1:2181", "src zk addr")
	flag.StringVar(&dst, "dst", "127.0.0.1:2181", "dst zk addr")
	flag.StringVar(&path, "path", "/", "zkpath")
	flag.Parse()
}

func main() {
	dstZkAddr := strings.Split(dst, ",")
	dstZk, _, err := zk.Connect(dstZkAddr, time.Second*5)
	if err != nil {
		fmt.Println(err)
		return
	}

	srcZkAddr := strings.Split(src, ",")
	srcZk, _, err := zk.Connect(srcZkAddr, time.Second*5)
	if err != nil {
		fmt.Println(err)
		return
	}

	iterChild(srcZk, dstZk, path)
}

func iterChild(src, dst *zk.Conn, path string) {
	childs, _, _ := src.Children(path)
	if len(childs) > 0 {
		dstCheckPath(dst, path)
		for _, child := range childs {
			if path == "/" {
				path = ""
			}
			iterChild(src, dst, fmt.Sprintf("%s/%s", path, child))
		}
	} else {
		data, stat, _ := src.Get(path)
		exist, _, _ := dst.Exists(path)
		if exist {
			_, err := dst.Set(path, data, stat.Version)
			if err != nil {
				fmt.Printf("path %s err: %s \n", path, err)
			} else {
				fmt.Printf("path %s replace ok\n", path)
			}
		} else {
			_, err := dst.Create(path, data, 0, zk.WorldACL(zk.PermAll))
			if err != nil {
				fmt.Printf("path %s err: %s \n", path, err)
			} else {
				fmt.Printf("path %s create ok\n", path)
			}
		}
	}
}

func dstCheckPath(dst *zk.Conn, path string) {
	exist, _, _ := dst.Exists(path)
	if !exist {
		_, err := dst.Create(path, []byte{}, 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			fmt.Printf("path %s err: %s \n", path, err)
		} else {
			fmt.Printf("path %s create ok\n", path)
		}
	}
}
