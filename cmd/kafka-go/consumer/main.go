package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/segmentio/kafka-go"
	"strings"
	"time"
)

var (
	topic         string
	brokers       string
	groupId       string
	minBytes      int
	maxBytes      int
	latest        bool
	startOffset   int64
	startTime     string
	startUnixTime int64
	maxMessages   int
)

func init() {
	flag.StringVar(&topic, "topic", "", "kafka topic")
	flag.StringVar(&brokers, "brokers", "", "kafka broker")
	flag.StringVar(&groupId, "group", "", "consumer group id, no value means Fetch")
	flag.IntVar(&minBytes, "minBytes", 1024, "min bytes, default 1024")
	flag.IntVar(&maxBytes, "maxBytes", 10485760, "max bytes, default 10485760 = 1024 * 1024 * 10")
	flag.BoolVar(&latest, "latest", true, "if start from latest, default true")
	flag.Int64Var(&startOffset, "startOffset", 0, "start offset, default 0")
	flag.StringVar(&startTime, "startTime", "", "start time, RFC3339 format (2006-01-02T15:04:05Z07:00)")
	flag.Int64Var(&startUnixTime, "startUnixTime", 0, "start unix time, Mill second")
	flag.IntVar(&maxMessages, "maxMessages", 10, "consume message count, default 10")
}

func main() {
	flag.Parse()
	if len(brokers) == 0 {
		panic("no Kafka bootstrap brokers defined, please set the -brokers flag")
	}
	if len(topic) == 0 {
		panic("no topics given to be consumed, please set the -topic flag")
	}
	brokerList := strings.Split(brokers, ",")
	start := kafka.LastOffset
	if !latest {
		start = kafka.FirstOffset
	}

	if groupId == "" {
		// fetch
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:     brokerList,
			Topic:       topic,
			MinBytes:    minBytes,
			MaxBytes:    maxBytes,
			StartOffset: start,
		})
		defer reader.Close()

		var startFrom time.Time
		if startUnixTime > 0 {
			startFrom = time.UnixMilli(startUnixTime)
		}
		if startTime != "" {
			startTemp, err := time.Parse(time.RFC3339, startTime)
			if err == nil {
				startFrom = startTemp
			}
		}
		ctx := context.Background()

		if startOffset != 0 {
			_ = reader.SetOffset(startOffset)
		} else if !startFrom.IsZero() {
			_ = reader.SetOffsetAt(ctx, startFrom)
		}
		count := 0
		for {
			msg, err := reader.FetchMessage(ctx)
			if err != nil {
				fmt.Println("read msg err:", err.Error())
				continue
			}
			fmt.Println("get msg:", msg.Topic, string(msg.Value))
			count++
			if maxMessages > 0 && count >= maxMessages {
				fmt.Println("get message counts: ", count)
				return
			}
		}
	} else {
		// consume
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:     brokerList,
			GroupID:     groupId,
			Topic:       topic,
			MinBytes:    minBytes,
			MaxBytes:    maxBytes,
			StartOffset: start,
		})
		defer reader.Close()
		for {
			msg, err := reader.ReadMessage(context.Background())
			if err != nil {
				fmt.Println("read msg err:", err.Error())
				continue
			}
			fmt.Println("get msg:", msg.Topic, string(msg.Value))
		}
	}

}
