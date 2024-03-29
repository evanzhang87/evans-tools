package main

import (
	"fmt"
	"github.com/evanzhang87/evans-tools/pkg/config"
)

type Config struct {
	Id   int    `yaml:"id" required:"true"`
	Name string `yaml:"name" default:"evan"`
	Age  int    `yaml:"age" min:"1"`
	Num  int    `yaml:"num" max:"100"`
}

func (c *Config) print() {
	fmt.Println(c.Id, c.Name, c.Age, c.Num)
}

func main() {
	var conf Config
	err := config.LoadConfig(&conf, "config.yaml")
	if err != nil {
		fmt.Println(err.Error())
	}
	conf.print()
}
