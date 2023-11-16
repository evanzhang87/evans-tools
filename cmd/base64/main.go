package main

import (
	"encoding/base64"
	"flag"
	"fmt"
)

var (
	str string
	opt string
)

func init() {
	flag.StringVar(&str, "s", "", "input string")
	flag.StringVar(&opt, "o", "d", "d=decode, e=encode")
}

func main() {
	flag.Parse()
	if opt == "d" {
		decode, err := base64.StdEncoding.DecodeString(str)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(string(decode))
		}
	} else {
		encoded := base64.StdEncoding.EncodeToString([]byte(str))
		fmt.Println(encoded)
	}
}
