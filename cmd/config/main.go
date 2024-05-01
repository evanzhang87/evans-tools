package main

import (
	"fmt"
	"github.com/evanzhang87/evans-tools/pkg/config"
	"os"
	"os/signal"
	"time"
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
	reloader := config.NewReloader(&conf)
	err := reloader.WatchPath("config.yaml")
	if err != nil {
		fmt.Println(err.Error())
	}
	signChan := make(chan os.Signal, 1)
	signal.Notify(signChan, os.Kill, os.Interrupt)
	configChan := reloader.SubscribeConfig()
	reloadTicker := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-signChan:
			return
		case configOutlet := <-configChan:
			if configOutlet != nil {
				configOutlet.(*Config).print()
			}
		case <-reloadTicker.C:
			confFetch := reloader.FetchConfig()
			if confFetch != nil {
				confFetch.(*Config).print()
			}
		}
	}
}
