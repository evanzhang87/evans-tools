package main

import "evans-tools/pkg/logger"

func main() {
	config := &logger.LogConfig{}
	logger.InitLogger(config)
	log := logger.GetLogger()
	log.Info("INFO: aaa")
	log.Warn("WARN: bbb")
}
