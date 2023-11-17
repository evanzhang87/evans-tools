package main

import (
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
	"os"
	"strings"
	"time"
)

var dir string

func init() {
	flag.StringVar(&dir, "d", "", "dir")
}

func main() {
	flag.Parse()
	if dir == "" {
		dir, _ = os.Getwd()
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if event.Op == fsnotify.Write {
					continue
				}
				if event.Op == fsnotify.Rename {
					fmt.Println("path", event.Name)
				}
				if !ok {
					return
				}
				if strings.Contains(event.String(), "RENAME") {
					err := watcher.Remove(dir)
					if err != nil {
						fmt.Println("remove error", err)
					}
					time.Sleep(time.Second)
					err2 := watcher.Add(dir)
					if err2 != nil {
						fmt.Println("add error", err2)
					}
				}
				if !strings.Contains(event.String(), "WRITE") {
					fmt.Println("event:", event)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("error:", err)
			}
		}
	}()
	_ = watcher.Add(dir)
	err = watcher.Add(dir)
	if err != nil {
		log.Fatal(err)
	}

	<-make(chan struct{})
}
